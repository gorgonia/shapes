// package shapes provides an algebra for dealing with shapes.
package shapes // import "gorgonia.org/shapes"

import (
	"fmt"

	"github.com/pkg/errors"
)

var (
	scalarShape = Shape{}
)

// ScalarShape returns a shape that represents a scalar shape.
//
// Usually `nil` will also be considered a scalar shape
// (because a `nil` of type `Shape` has a length of 0 and will return true when `.IsScalar` is called)
func ScalarShape() Shape { return scalarShape }

// Shape represents the shape of a multidimensional array.
type Shape []int

// Shape implements Shaper - it returns itself
func (s Shape) Shape() Shape { return s }

// Cons is an associative construction of shapes
func (s Shape) Cons(other Conser) Conser {
	switch ot := other.(type) {
	case Shape:
		return append(s, ot...)
	case Abstract:
		retVal := s.toAbs(len(s) + len(ot))
		retVal = append(retVal, ot...)
		return retVal
	}
	panic("Unreachable")
}

func (s Shape) isConser() {}

func (s Shape) toAbs(hint int) Abstract {
	if hint <= 0 {
		hint = len(s)
	}
	retVal := make(Abstract, 0, hint)
	for i := range s {
		retVal = append(retVal, Size(s[i]))
	}
	return retVal
}

func (s Shape) Clone() Shape {
	retVal := make(Shape, len(s))
	copy(retVal, s)
	return retVal
}

func (s Shape) AsInts() []int { return []int(s) }

// Eq indicates if a shape is equal with another. There is a soft concept of equality when it comes to vectors.
//
// If s is a column vector and other is a vanilla vector, they're considered equal if the size of the column dimension is the same as the vector size;
// if s is a row vector and other is a vanilla vector, they're considered equal if the size of the row dimension is the same as the vector size
func (s Shape) Eq(other Shape) bool {
	if s.IsScalar() && other.IsScalar() {
		return true
	}

	if s.IsVector() && other.IsVector() {
		switch {
		case len(s) == 2 && len(other) == 1:
			if (s.IsColVec() && s[0] == other[0]) || (s.IsRowVec() && s[1] == other[0]) {
				return true
			}
			return false
		case len(s) == 1 && len(other) == 2:
			if (other.IsColVec() && other[0] == s[0]) || (other.IsRowVec() && other[1] == s[0]) {
				return true
			}
			return false
		}
	}

	if len(s) != len(other) {
		return false
	}

	for i, v := range s {
		if other[i] != v {
			return false
		}
	}
	return true
}

// IsScalar returns true if the access pattern indicates it's a scalar value
func (s Shape) IsScalar() bool { return len(s) == 0 }

// IsScalarEquiv returns true if the access pattern indicates it's a scalar-like value
func (s Shape) IsScalarEquiv() bool {
	if s.IsScalar() {
		return true
	}
	asInts := []int(s)
	if allEq(asInts, 0) {
		return true
	}
	return prodInts(asInts) == 1
}

// IsVector returns whether the access pattern falls into one of three possible definitions of vectors:
//
//	vanilla vector (not a row or a col)
//	column vector
//	row vector
func (s Shape) IsVector() bool { return s.IsColVec() || s.IsRowVec() || (len(s) == 1) }

// IsColVec returns true when the access pattern has the shape (x, 1)
func (s Shape) IsColVec() bool { return len(s) == 2 && (s[1] == 1 && s[0] > 1) }

// IsRowVec returns true when the access pattern has the shape (1, x)
func (s Shape) IsRowVec() bool { return len(s) == 2 && (s[0] == 1 && s[1] > 1) }

// IsVectorLike returns true when the shape looks like a vector
// e.g. a number that is surrounded by 1s:
//
//	(1, 1, ... 1, 10, 1, 1... 1)
func (s Shape) IsVectorLike() bool {
	var nonOnes int
	for _, i := range s {
		if i != 1 {
			nonOnes++
		}
	}
	return nonOnes == 1 || nonOnes == 0 // if there is only one non-one then it's a vector or a scalarlike.
}

// IsMatrix returns true if it's a matrix. This is mostly a convenience method. RowVec and ColVecs are also considered matrices
func (s Shape) IsMatrix() bool { return len(s) == 2 }

// Dims returns the number of dimensions in the shape
func (s Shape) Dims() int { return len(s) }

func (s Shape) TotalSize() int { return prodInts([]int(s)) }

// DimSize returns the dimension wanted.
func (s Shape) DimSize(d int) (retVal Sizelike, err error) {
	ret, err := s.Dim(d)
	if err != nil {
		return nil, err
	}
	return Size(ret), nil
}

// Dim returns the dimension wanted,
func (s Shape) Dim(d int) (retVal int, err error) {
	if (s.IsScalar() && d != 0) || (!s.IsScalar() && d >= len(s)) {
		return -1, errors.Errorf(dimMismatch, len(s), d)
	}
	switch {
	case s.IsScalar():
		return 0, nil
	case d < 0:
		od := d
		d = s.Dims() + d
		if d < 0 {
			return -1, errors.Errorf(dimMismatch, len(s), od)
		}
		fallthrough
	default:
		return s[d], nil
	}
}

func (s Shape) T(axes ...Axis) (newShape Shapelike, err error) {
	retVal := make(Shape, len(s))
	copy(retVal, s)
	err = UnsafePermute(axesToInts(axes), []int(retVal))
	newShape = retVal
	return
}

// S gives the new shape after a shape has been sliced.
func (s Shape) S(slices ...Slice) (newShape Shapelike, err error) {
	opDims := len(s)
	if len(slices) > opDims {
		err = errors.Errorf(dimsMismatch, opDims, len(slices))
		return
	}

	retVal := s.Clone()

	for d, size := range s {
		var sl Slice // default is a nil Slice
		if d <= len(slices)-1 {
			sl = slices[d]
		}
		if retVal[d], err = sliceSize(sl, size); err != nil {
			return nil, errors.Wrapf(err, "Unable to slice shape %v. Dim %d caused an error", s, d)
		}

	}

	// drop any dimension with size 1, except the last dimension
	offset := 0
	dims := s.Dims()
	for d := 0; d < dims; d++ {
		if retVal[d] == 1 && offset+d <= len(slices)-1 && slices[offset+d] != nil /*&& d != t.dims-1  && dims > 2*/ {
			retVal = append(retVal[:d], retVal[d+1:]...)
			d--
			dims--
			offset++
		}
	}

	if retVal.IsScalar() {
		//ReturnInts(retVal)
		return ScalarShape(), nil
	}
	newShape = retVal
	return
}

// Repeat returns the expected new shape given the repetition parameters
func (s Shape) Repeat(axis Axis, repeats ...int) (retVal Shapelike, finalRepeats []int, size int, err error) {
	var newShape Shape
	switch {
	case axis == AllAxes:
		size = s.TotalSize()
		newShape = Shape{size}
		axis = 0
	case s.IsScalar():
		size = 1
		// special case for row vecs
		newShape = make(Shape, axis+1)
		for i := range newShape {
			newShape[i] = 1
		}

		// if axis == 1 {
		// 	newShape = Shape{1, 0}
		// } else {
		// 	// otherwise it will be repeated into a vanilla vector
		// 	newShape = Shape{0}
		// }
	case s.IsVector() && !s.IsRowVec() && !s.IsColVec() && axis == 1:
		size = 1
		newShape = s.Clone()
		newShape = append(newShape, 1)
	default:
		if int(axis) >= s.Dims() {
			// error
			err = errors.Errorf(invalidAxis, axis, s.Dims())
			return
		}
		size = s[axis]
		newShape = s.Clone()
	}
	// special case to allow generic repeats
	if len(repeats) == 1 {
		rep := repeats[0]
		repeats = make([]int, size)
		for i := range repeats {
			repeats[i] = rep
		}
	}
	reps := len(repeats)
	if reps != size {
		err = errors.Errorf(broadcastError, size, reps)
		return
	}

	newSize := sumInts(repeats)
	newShape[axis] = newSize
	finalRepeats = repeats

	retVal = newShape
	return
}

func (s Shape) Concat(axis Axis, ss ...Shapelike) (retVal Shapelike, err error) {
	dims := s.Dims()

	// check that all the concatenates have the same dimensions
	for _, shp := range ss {
		if shp.Dims() != dims {
			err = errors.Errorf(dimMismatch, dims, shp.Dims())
			return
		}
	}

	// special case
	if axis == AllAxes {
		axis = 0
	}

	// nope... no negative indexing here.
	if axis < 0 {
		err = errors.Errorf(invalidAxis, axis, len(s))
		return
	}

	if int(axis) >= dims {
		err = errors.Errorf(invalidAxis, axis, len(s))
		return
	}

	newShape := s.Clone()
	for _, sl := range ss {
		shp := sl.(Shape) // will panic if not a shape
		for d := 0; d < dims; d++ {
			if d == int(axis) {
				newShape[d] += shp[d]
			} else {
				// validate that the rest of the dimensions match up
				if newShape[d] != shp[d] {
					err = errors.Wrapf(errors.Errorf(dimMismatch, newShape[d], shp[d]), "Axis: %d, dimension it failed at: %d", axis, d)
					return
				}
			}
		}
	}
	retVal = newShape
	return
}

// Format implements fmt.Formatter, and formats a shape nicely
func (s Shape) Format(st fmt.State, r rune) {
	switch r {
	case 'v', 's':
		st.Write([]byte("("))
		for i, v := range s {
			fmt.Fprintf(st, "%d", v)
			if i < len(s)-1 {
				st.Write([]byte(", "))
			}
		}
		st.Write([]byte(")"))
	default:
		fmt.Fprintf(st, "%v", []int(s))
	}
}

// apply doesn't apply any substitutions to Shape because there will not be anything to substitution.
func (s Shape) apply(_ substitutions) substitutable { return s }

// freevar returns nil because there are no free variables in a Shape.
func (s Shape) freevars() varset { return nil }

func (s Shape) isExpr() {}

// subExprs returns the shape as a slice of Expr (specifically, it becomes a slice of Size)
func (s Shape) subExprs() (retVal []substitutableExpr) {
	retVal = make([]substitutableExpr, 0, len(s))
	for i := range s {
		retVal = append(retVal, Size(s[i]))
	}
	return retVal
}
