package shapes

import (
	"fmt"

	"github.com/pkg/errors"
)

// wrappers are wrappers for specific types to convert them from one type to another

var _ sizeOp = sizelikeSliceOf{}
var _ sizeOp = E2{}
var _ fmt.Formatter = sizelikeSliceOf{}

type sizelikeSliceOf struct{ SliceOf }

func (s sizelikeSliceOf) isSizelike() {}

func (s sizelikeSliceOf) apply(ss substitutions) substitutable {
	so := s.SliceOf.apply(ss).(SliceOf)
	return sizelikeSliceOf{so}
}

func (s sizelikeSliceOf) resolveSize() (retVal Size, err error) {
	if !s.isValid() {
		return -1, errors.Errorf("Cannot resolveSize - some variables remain")
	}
	switch a := s.SliceOf.A.(type) {
	case Size:
		var x int
		if x, err = sliceSize(s.SliceOf.Slice.(Slice), int(a)); err != nil {
			return -1, errors.Wrapf(err, "Unable to resolve %v into a Size", s)
		}
		return Size(x), nil
	case sizeOp:
		var x int
		if retVal, err = a.resolveSize(); err != nil {
			return -1, errors.Wrapf(err, "Unable to resolve %v into a Size", s)
		}

		if x, err = sliceSize(s.SliceOf.Slice.(Slice), int(retVal)); err != nil {
			return -1, errors.Wrapf(err, "Unable to resolve %v into a Size", s)
		}
		return Size(x), nil
	}
	panic("Unreachable")
}

type E2 struct{ BinOp }

func (e E2) isExpr()    {}
func (e E2) depth() int { return max(e.A.depth(), e.B.depth()) + 1 }

func (e E2) apply(ss substitutions) substitutable {
	bo := e.BinOp.apply(ss).(BinOp)
	return E2{bo}
}
