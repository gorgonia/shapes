package shapes

import (
	"fmt"

	"github.com/pkg/errors"
)

type rr struct{ start, end rune }

// generator holds the alphabet range allowed for variables
var generator = [...]rr{
	{'a', 'z'},
	{'α', 'ω'},
}

// Abstract is an abstract shape
type Abstract []Sizelike

// Gen creates an abstract with the provided dims.
// This is particularly useful for generating abstracts for higher order functions.
func Gen(d int) (retVal Abstract) {
	if d <= 0 {
		panic("Cannot generate an Abstract with d <= 0")
	}
	type state struct {
		rangeNum int
		letter   rune
	}
	s := state{0, 'a'}
	for i := 0; i < d; i++ {
		retVal = append(retVal, Var(s.letter))

		s.letter++
		if s.letter > generator[s.rangeNum].end {
			s.rangeNum++
			s.letter = generator[s.rangeNum].start
		}
	}
	return retVal
}

func (a Abstract) Cons(other Conser) (retVal Conser) {
	switch ot := other.(type) {
	case Shape:
		l := len(a)
		r := append(a, make(Abstract, len(ot))...)
		r = r[:l]
		for _, v := range ot {
			r = append(r, Size(v))
		}
		retVal = r
	case Abstract:
		retVal = append(a, ot...)
	}
	return retVal
}

func (a Abstract) isConser() {}

func (a Abstract) ToShape() (s Shape, ok bool) {
	s = make(Shape, len(a)) // TODO: perf - borrow
	for i := range a {
		sz, ok := a[i].(Size)
		if !ok {
			return nil, ok
		}
		s[i] = int(sz)
	}
	return s, true
}

func (a Abstract) Clone() Abstract {
	retVal := make(Abstract, len(a))
	copy(retVal, a)
	return retVal
}

func (a Abstract) isExpr() {}

// Dims returns the number of dimensions in the shape
func (a Abstract) Dims() int { return len(a) }

func (a Abstract) TotalSize() int { panic("Unable to get TotalSize for Abstract") }

func (a Abstract) DimSize(dim int) (Sizelike, error) {
	if a.Dims() <= dim {
		return nil, errors.Errorf("Cannot get Dim %d of %v", dim, a)
	}
	return a[dim], nil
}

func (a Abstract) T(axes ...Axis) (newShape Shapelike, err error) {
	retVal := make(Abstract, len(a))
	copy(retVal, a)
	err = genericUnsafePermute(axesToInts(axes), retVal)
	newShape = retVal
	return
}

func (a Abstract) S(slices ...Slice) (newShape Shapelike, err error) {
	if shp, ok := a.ToShape(); ok {
		return shp.S(slices...)
	}

	opDims := len(a)
	if len(slices) > opDims {
		err = errors.Errorf(dimsMismatch, opDims, len(slices))
		return
	}

	retVal := a.Clone()
	for d, size := range a {

		var sl Slice // default is a nil Slice
		if d <= len(slices)-1 {
			sl = slices[d]
		}
		if sl == nil {
			retVal[d] = size
			continue
		}

		switch s := size.(type) {
		case Size:
			var x int
			if x, err = sliceSize(sl, int(s)); err != nil {
				return nil, errors.Wrapf(err, "Unable to slice %v. Dim %d caused an error.", a, d)
			}
			retVal[d] = Size(x)

		case Var:
			retVal[d] = sizelikeSliceOf{SliceOf{toRange(sl), s}}
		case BinOp:
			retVal[d] = sizelikeSliceOf{SliceOf{toRange(sl), exprBinOp{s}}}
		case UnaryOp:
			retVal[d] = sizelikeSliceOf{SliceOf{toRange(sl), s}}
		default:
			return nil, errors.Errorf("%dth sizelike %v of %T is unsupported by S(). Perhaps make a pull request?", d, size, size)
		}

		// attempt to resolve if possible
		if s, ok := retVal[d].(sizeOp); ok && s.isValid() {
			if sz, err := s.resolveSize(); err == nil {
				retVal[d] = sz
			}
		}

	}

	// drop any dimension with size 1, except the last dimension
	offset := 0
	dims := a.Dims()
	for d := 0; d < dims; d++ {
		if sz, ok := retVal[d].(Size); ok && sz == 1 && offset+d <= len(slices)-1 && slices[offset+d] != nil {
			retVal = append(retVal[:d], retVal[d+1:]...)
			d--
			dims--
			offset++
		}
	}

	if shp, ok := retVal.ToShape(); ok {
		if shp.IsScalar() {
			return ScalarShape(), nil
		}
		return shp, nil
	}

	return retVal, nil
}

func (a Abstract) Repeat(axis Axis, repeats ...int) (retVal Shapelike, finalRepeats []int, size int, err error) {
	var newShape Abstract
	var sz Sizelike
	switch {
	case axis == AllAxes:
		sz = UnaryOp{Prod, a}
		newShape = Abstract{sz}
		axis = 0
	case a.Dims() == 1 && axis == 1: // "vector"
		sz = Size(1)
		newShape = a.Clone()
		newShape = append(newShape, Size(1))
	default:
		if int(axis) >= a.Dims() {
			// error
			err = errors.Errorf(invalidAxis, axis, a.Dims())
			return
		}
		sz = a[axis]
		newShape = a.Clone()
	}

	size = -1
	switch s := sz.(type) {
	case Size:
		size = int(s)
		// special case to allow generic repeats
		if size > 0 && len(repeats) == 1 {
			rep := repeats[0]
			repeats = make([]int, size)
			for i := range repeats {
				repeats[i] = rep
			}
		}
		// optimistically check
		reps := len(repeats)
		if size > 0 && reps != size {
			err = errors.Errorf(broadcastError, size, reps)
			return
		}
		newSize := sumInts(repeats)
		newShape[axis] = Size(newSize)

		// set return values
		finalRepeats = repeats
	default:
		// special case to allow generic repeats
		if len(repeats) == 1 {
			rep := Size(repeats[0])
			newSize := BinOp{Mul, sz.(Expr), rep}
			newShape[axis] = newSize

			// set return values
			finalRepeats = repeats
		} else {
			// cannot check if newShape[axis] == len(repeats)
			// gotta take it on faith
			newSize := sumInts(repeats)
			newShape[axis] = Size(newSize)
			// don't set finalRepeats. Should be nil
		}
	}

	retVal = newShape

	// try to resolve
	if x, err := newShape.resolve(); err == nil {
		retVal = x.(Shapelike)
	}

	return
}

func (a Abstract) Concat(axis Axis, others ...Shapelike) (newShape Shapelike, err error) {
	panic("not implemented") // TODO: Implement
}

// Format implements fmt.Formatter, and formats a shape nicely
func (s Abstract) Format(st fmt.State, r rune) {
	switch r {
	case 'v', 's':
		st.Write([]byte("("))
		for i, v := range s {
			switch vt := v.(type) {
			case Size:
				fmt.Fprintf(st, "%d", int(vt))
			case Var:
				fmt.Fprintf(st, "%c", rune(vt))
			default:
				fmt.Fprintf(st, "%v", v)
			}
			if i < len(s)-1 {
				st.Write([]byte(", "))
			}
		}
		st.Write([]byte(")"))
	default:
		fmt.Fprintf(st, "%v", []Sizelike(s))
	}
}

func (s Abstract) apply(ss substitutions) substitutable {
	retVal := make(Abstract, len(s))
	copy(retVal, s)
	for i, a := range s {
		st := a.(substitutable)
		retVal[i] = st.apply(ss).(Sizelike)
	}
	return retVal
}

func (s Abstract) freevars() (retVal varset) {
	for _, a := range s {
		if v, ok := a.(Var); ok {
			retVal = append(retVal, v)
		}
	}
	return unique(retVal)
}

func (s Abstract) subExprs() (retVal []substitutableExpr) {
	for i := range s {
		retVal = append(retVal, s[i].(substitutableExpr))
	}
	return retVal
}

func (s Abstract) resolve() (Expr, error) {
	retVal := s.Clone()
	for i, v := range s {
		switch r := v.(type) {
		case sizeOp:
			if !r.isValid() {
				retVal[i] = v
				continue
			}
			sz, err := r.resolveSize()
			if err != nil {
				return nil, errors.Errorf("%dth sizelike of %v is not resolveable to a Size", i, s)
			}
			retVal[i] = sz

		default:
			return nil, errors.Errorf("Sizelike of %T is unhandled by Abstract", v)
		}

	}
	if shp, ok := retVal.ToShape(); ok {
		return shp, nil
	}
	return retVal, nil
}
