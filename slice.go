package shapes

import "fmt"

// Slice represents a slicing range.
type Slice interface {
	Start() int
	End() int
	Step() int
}

// Range is a shape expression representing a slicing range. Coincidentally, Range also implements Slice.
//
// A Range is a shape expression but it doesn't stand alone - resolving it will yield an error.
type Range struct {
	start, end, step int
}

func (s Range) isExpr() {}

func (s Range) apply(ss substitutions) substitutable { return s }
func (s Range) freevars() varset                     { return nil }

// Exprs returns nil because we want a Sli to be treated as a monolithic expression term with nothing to unify on the inside.
func (s Range) subExprs() []substitutableExpr { return nil }

// Format allows Sli to implement fmt.Formmatter
func (s Range) Format(st fmt.State, r rune) {
	if st.Flag('#') {
		fmt.Fprintf(st, "{%d:%d:%d}", s.start, s.end, s.step)
		return
	}
	fmt.Fprintf(st, "[%d", s.start)
	if s.end-s.start > 1 {
		fmt.Fprintf(st, ":%d", s.end)
	}
	if s.step > 1 {
		fmt.Fprintf(st, "~:%d", s.step)
	}
	st.Write([]byte("]"))
}

/* Sli implements Slice */

// Start returns the start of the slicing range
func (s Range) Start() int { return s.start }

// End returns the end of the slicing range
func (s Range) End() int { return s.end }

// Step returns the steps/jumps to make in the slicing range.
func (s Range) Step() int { return s.step }

// isSlicelike makes Range implement slicelike
func (s Range) isSlicelike() {}

// S creates a Slice. Internally it uses the Range type provided.
func S(start int, opt ...int) *Range {
	var end, step int
	if len(opt) > 0 {
		end = opt[0]
	} else {
		end = start + 1
	}

	step = 1
	if len(opt) > 1 {
		step = opt[1]
	}
	return &Range{
		start: start,
		end:   end,
		step:  step,
	}
}

// toRange creates a Sli from a Slice.
func toRange(s Slice) Range {
	if ss, ok := s.(Range); ok {
		return ss
	}
	if ss, ok := s.(*Range); ok {
		return *ss
	}
	return Range{s.Start(), s.End(), s.Step()}
}

// sliceSize is a support function for slicing a number
func sliceSize(sl Slice, sz int) (retVal int, err error) {

	var start, end, step int
	if start, end, step, err = SliceDetails(sl, sz); err != nil {
		return
	}

	if step > 1 {
		retVal = (end - start) / step

		//fix
		if retVal <= 0 {
			retVal = 1
		}
	} else {
		retVal = (end - start)
	}
	return
}

// ToSlicelike is a utility function for turning a slice into a Slicelike.
func ToSlicelike(s Slice) Slicelike { return toRange(s) }
