package shapes

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"
)

// intrinsic.go describes all the intrinsic operations that a shape can do.
// By intrinsic operation, I mean that these are symbolic versions of the shape operations.
// All comments/documentation will contain a phrase "symbolic version of: XXX"

// IndexOf gets the size of a given shape (expression) at the given index.
//
// IndexOf is the symbolic version of doing s[i], where s is a Shape.
type IndexOf struct {
	I Size
	A Expr
}

func (i IndexOf) isExpr()                    {}
func (i IndexOf) Format(s fmt.State, r rune) { fmt.Fprintf(s, "%v[%d]", i.A, i.I) }
func (i IndexOf) apply(ss substitutions) substitutable {
	return IndexOf{
		I: i.I,
		A: i.A.apply(ss).(Expr),
	}
}
func (i IndexOf) freevars() varset { return i.A.freevars() }
func (i IndexOf) subExprs() []substitutableExpr {
	return []substitutableExpr{i.I, i.A.(substitutableExpr)}
}

func (i IndexOf) isValid() bool { return true }
func (i IndexOf) resolveSize() (Size, error) {
	if len(i.A.freevars()) > 0 {
		return 0, errors.Errorf("Cannot resolve IndexOf %v - free vars found", i)
	}
	switch at := i.A.(type) {
	case Shapelike:
		if at.Dims() <= int(i.I) {
			return 0, errors.Errorf("Expression %v has %d Dims. Want to get index of %d", at, at.Dims(), i.I)
		}
		sz, err := at.DimSize(int(i.I))
		if err != nil {
			return 0, errors.Wrapf(err, "Cannot get Index %d of %v", i.I, i.A)
		}
		switch s := sz.(type) {
		case Size:
			return s, nil
		case sizeOp:
			return s.resolveSize()
		default:
			return 0, errors.Errorf("Sizelike of %v (Index %d of %v)is unresolvable ", sz, i.I, i.A)
		}
	default:
		return 0, errors.Errorf("Cannot resolve IndexOf %v - expression of %T is unhandled", i.A, i.A)
	}
}

// TransposeOf is the symbolic version of doing s.T(axes...)
type TransposeOf struct {
	Axes Axes
	A    Expr
}

func (t TransposeOf) isExpr()                    {}
func (t TransposeOf) Format(s fmt.State, r rune) { fmt.Fprintf(s, "T%x %v", t.Axes, t.A) }
func (t TransposeOf) apply(ss substitutions) substitutable {
	return TransposeOf{
		Axes: t.Axes,
		A:    t.A.apply(ss).(Expr),
	}
}
func (t TransposeOf) freevars() varset { return t.A.freevars() }
func (t TransposeOf) subExprs() []substitutableExpr {
	return []substitutableExpr{t.Axes, t.A.(substitutableExpr)}
}
func (t TransposeOf) resolve() (Expr, error) {
	switch at := t.A.(type) {
	case Shapelike:
		retVal, err := at.T(t.Axes...)
		_, ok := err.(NoOpError)
		if !ok && err != nil {
			return nil, err
		}
		return retVal.(Expr), nil
	default:
		return nil, errors.Errorf("Cannot transpose Expression %v of %T", t.A, t.A)
	}
	panic("Unreachable")
}

// SliceOf is an intrinsic operation, symbolically representing a slicing operation.
type SliceOf struct {
	Slice Slicelike
	A     Expr
}

func (s SliceOf) isExpr() {}
func (s SliceOf) Format(st fmt.State, r rune) {
	switch s.Slice.(type) {
	case Slice:
		fmt.Fprintf(st, "%v%v", s.A, s.Slice)
	case Slices:
		fmt.Fprintf(st, "%v%v", s.A, s.Slice)
	case Var:
		fmt.Fprintf(st, "%v[%v]", s.A, s.Slice)

	}

}
func (s SliceOf) apply(ss substitutions) substitutable {
	return SliceOf{
		Slice: s.Slice.apply(ss).(Slicelike),
		A:     s.A.apply(ss).(Expr),
	}
}
func (s SliceOf) freevars() varset { return s.A.freevars() }
func (s SliceOf) subExprs() []substitutableExpr {
	return []substitutableExpr{s.Slice, s.A.(substitutableExpr)}
}

func (s SliceOf) resolve() (Expr, error) {
	switch at := s.A.(type) {
	case Shapelike:
		switch sl := s.Slice.(type) {
		case Slice:
			retVal, err := at.S(sl)
			if err != nil {
				return nil, errors.Wrapf(err, "Unable to resolve %v - .S() failed", s)
			}
			return retVal.(Expr), nil
		case Slices:
			retVal, err := at.S(sl...)
			if err != nil {
				return nil, errors.Wrapf(err, "Unable to resolve %v - .S() failed", s)
			}
			return retVal.(Expr), nil
		default:
			return s, nil

		}

	default:
		return nil, errors.Errorf("Cannot slice Expression %v of %T", s.A, s.A)
	}
}

// isValid makes SliceOf an Operation.
func (s SliceOf) isValid() bool {
	_, isVar := s.Slice.(Var)
	if isVar {
		return false
	}

	switch a := s.A.(type) {
	case Size:
		return true
	case Operation:
		return a.isValid()
	default:
		return false
	}
}

type ConcatOf struct {
	Along Axis
	A, B  Expr
}

func (c ConcatOf) isExpr()                    {}
func (c ConcatOf) Format(s fmt.State, r rune) { fmt.Fprintf(s, "%v :{%d}: %v", c.A, c.Along, c.B) }
func (c ConcatOf) apply(ss substitutions) substitutable {
	return ConcatOf{
		Along: c.Along,
		A:     c.A.apply(ss).(Expr),
		B:     c.B.apply(ss).(Expr),
	}
}
func (c ConcatOf) freevars() varset { return (exprtup{c.A, c.B}).freevars() }
func (c ConcatOf) subExprs() []substitutableExpr {
	return []substitutableExpr{c.Along, c.A.(substitutableExpr), c.B.(substitutableExpr)}
}

type RepeatOf struct {
	Along   Axis
	Repeats []Size
	A       Expr
}

func (r RepeatOf) isExpr() {}
func (r RepeatOf) Format(s fmt.State, ru rune) {
	fmt.Fprintf(s, "Repeat%x{%v} %v", r.Along, r.Repeats, r.A)
}
func (r RepeatOf) apply(ss substitutions) substitutable {
	return RepeatOf{
		Along:   r.Along,
		Repeats: r.Repeats,
		A:       r.A.apply(ss).(Expr),
	}
}
func (r RepeatOf) freevars() varset { return r.A.freevars() }
func (r RepeatOf) subExprs() []substitutableExpr {
	return []substitutableExpr{r.Along, r.A.(substitutableExpr)}
}

func (r RepeatOf) resolve() (Expr, error) {
	switch at := r.A.(type) {
	case Shapelike:
		retVal, _, _, err := at.Repeat(r.Along, sizesToInts(r.Repeats)...)
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to resolve %v. .Repeat() failed", r)
		}
		return retVal.(Expr), nil
	default:
		return nil, errors.Errorf("Cannot Repeat Expression %v of %T", r.A, r.A)
	}
}

// BroadcastOf represents the results of mutually broadcasting A and B expr.
type BroadcastOf struct {
	A, B Expr
}

func (b BroadcastOf) isExpr()                    {}
func (b BroadcastOf) Format(s fmt.State, r rune) { fmt.Fprintf(s, "(%v||%v)", b.A, b.B) }
func (b BroadcastOf) apply(ss substitutions) substitutable {
	return BroadcastOf{
		A: b.A.apply(ss).(Expr),
		B: b.B.apply(ss).(Expr),
	}
}
func (b BroadcastOf) freevars() varset { return append(b.A.freevars(), b.B.freevars()...) }
func (b BroadcastOf) subExprs() []substitutableExpr {
	return []substitutableExpr{b.A.(substitutableExpr), b.B.(substitutableExpr)}
}

// ReductOf represents the results of reducing A along the given axis.
type ReductOf struct {
	A     Expr
	Along Axis
}

func Reduce(a Expr, along Axes) ReductOf {
	ints := axesToInts(along)
	sort.Sort(sort.Reverse(sort.IntSlice(ints)))

	// innermost reductions first
	var retVal Expr = a
	for i := range ints {
		retVal = ReductOf{retVal, Axis(ints[i])}
	}
	return retVal.(ReductOf)
}

func (r ReductOf) isExpr()                    {}
func (r ReductOf) Format(s fmt.State, c rune) { fmt.Fprintf(s, "/%x%v", r.Along, r.A) }
func (r ReductOf) apply(ss substitutions) substitutable {
	return ReductOf{
		A:     r.A.apply(ss).(Expr),
		Along: r.Along,
	}
}

func (r ReductOf) freevars() varset { return r.A.freevars() }
func (r ReductOf) subExprs() []substitutableExpr {
	return []substitutableExpr{r.A.(substitutableExpr)}
}

func (r ReductOf) resolve() (Expr, error) {
	A := r.A
	for {
		switch at := A.(type) {
		case Shape:
			along := ResolveAxis(r.Along, at)
			if along == AllAxes {
				return ScalarShape(), nil
			}
			retVal := make(Shape, len(at)-1)
			copy(retVal, at[:along])
			copy(retVal[along:], at[along+1:])
			return retVal, nil
		case Abstract:
			along := ResolveAxis(r.Along, at)
			if along == AllAxes {
				return ScalarShape(), nil
			}
			retVal := make(Abstract, len(at)-1)
			copy(retVal, at[:along])
			copy(retVal[along:], at[along+1:])
			return retVal, nil
		case resolver:
			expr, err := at.resolve()
			if err != nil {
				return nil, err
			}
			A = expr
		default:
			return nil, errors.Errorf("Cannot reduce Expression %v of %T", r.A, r.A)
		}
	}

}
