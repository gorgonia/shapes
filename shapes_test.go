package shapes

import (
	"fmt"
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/assert"
)

func TestShapes_Clone(t *testing.T) {
	f := func(s Shape) bool {
		s2 := s.Clone()
		if !s.Eq(s2) {
			return false
		}
		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

var shapesEqCases = []struct {
	a, b Shape
}{
	{Shape{}, Shape{}},      // scalar is always the same
	{Shape{2}, Shape{2, 1}}, // vector "soft equality"
	{Shape{1, 2}, Shape{2}}, // vector "soft equality"
	{Shape{1, 2, 3}, Shape{1, 2, 3}},
}

var shapesNeCases = []struct {
	a, b Shape
}{
	{Shape{}, Shape{1}}, // differing lengths
	{Shape{2}, Shape{1, 3}},
	{Shape{1, 2, 3}, Shape{1, 2, 4}},
}

func TestShapes_Eq(t *testing.T) {
	for _, c := range shapesEqCases {
		if !c.a.Eq(c.b) {
			t.Errorf("Expected %v = %v", c.a, c.b)
		}
		if !c.b.Eq(c.a) {
			t.Errorf("Expected %v = %v", c.b, c.a)
		}
	}

	for _, c := range shapesNeCases {
		if c.a.Eq(c.b) {
			t.Errorf("Expected %v != %v", c.a, c.b)
		}
		if c.b.Eq(c.a) {
			t.Errorf("Expected %v != %v", c.b, c.a)
		}
	}
}

func ExampleShape_IsScalarEquiv() {
	s := Shape{1, 1, 1, 1, 1, 1}
	fmt.Printf("%v is scalar equiv: %t\n", s, s.IsScalarEquiv())

	s = Shape{}
	fmt.Printf("%v is scalar equiv: %t\n", s, s.IsScalarEquiv())

	s = Shape{2, 3}
	fmt.Printf("%v is scalar equiv: %t\n", s, s.IsScalarEquiv())

	s = Shape{0, 0, 0}
	fmt.Printf("%v is scalar equiv: %t\n", s, s.IsScalarEquiv())

	s = Shape{1, 2, 0, 3}
	fmt.Printf("%v is scalar equiv: %t\n", s, s.IsScalarEquiv())

	// Output:
	// (1, 1, 1, 1, 1, 1) is scalar equiv: true
	// () is scalar equiv: true
	// (2, 3) is scalar equiv: false
	// (0, 0, 0) is scalar equiv: true
	// (1, 2, 0, 3) is scalar equiv: false

}

func TestShapeIsX(t *testing.T) {
	assert := assert.New(t)
	var s Shape

	// scalar shape
	s = Shape{}
	assert.True(s.IsScalar())
	assert.True(s.IsScalarEquiv())
	assert.False(s.IsVector())
	assert.False(s.IsColVec())
	assert.False(s.IsRowVec())

	// vectors

	// scalar-equiv vector
	s = Shape{1}
	assert.False(s.IsScalar())
	assert.True(s.IsScalarEquiv())
	assert.True(s.IsVector())
	assert.True(s.IsVectorLike())
	assert.False(s.IsColVec())
	assert.False(s.IsRowVec())

	// vanila vector
	s = Shape{2}
	assert.False(s.IsScalar())
	assert.True(s.IsVector())
	assert.False(s.IsColVec())
	assert.False(s.IsRowVec())

	// col vec
	s = Shape{2, 1}
	assert.False(s.IsScalar())
	assert.True(s.IsVector())
	assert.True(s.IsVectorLike())
	assert.True(s.IsColVec())
	assert.False(s.IsRowVec())

	// row vec
	s = Shape{1, 2}
	assert.False(s.IsScalar())
	assert.True(s.IsVector())
	assert.True(s.IsVectorLike())
	assert.False(s.IsColVec())
	assert.True(s.IsRowVec())

	// matrix and up
	s = Shape{2, 2}
	assert.False(s.IsScalar())
	assert.False(s.IsVector())
	assert.False(s.IsColVec())
	assert.False(s.IsRowVec())

	// scalar equiv matrix
	s = Shape{1, 1}
	assert.False(s.IsScalar())
	assert.True(s.IsScalarEquiv())
	assert.True(s.IsVectorLike())
	assert.False(s.IsVector())
}

func TestShapeEquality(t *testing.T) {
	assert := assert.New(t)
	var s1, s2 Shape

	// scalar
	s1 = Shape{}
	s2 = Shape{}
	assert.True(s1.Eq(s2))
	assert.True(s2.Eq(s1))

	// scalars and scalar equiv are not the same!
	s1 = Shape{1}
	s2 = Shape{}
	assert.False(s1.Eq(s2))
	assert.False(s2.Eq(s1))

	// vector
	s1 = Shape{3}
	s2 = Shape{5}
	assert.False(s1.Eq(s2))
	assert.False(s2.Eq(s1))

	s1 = Shape{2, 1}
	s2 = Shape{2, 1}
	assert.True(s1.Eq(s2))
	assert.True(s2.Eq(s1))

	s2 = Shape{2}
	assert.True(s1.Eq(s2))
	assert.True(s2.Eq(s1))

	s2 = Shape{1, 2}
	assert.False(s1.Eq(s2))
	assert.False(s2.Eq(s1))

	s1 = Shape{2}
	assert.True(s1.Eq(s2))
	assert.True(s2.Eq(s1))

	s2 = Shape{2, 3}
	assert.False(s1.Eq(s2))
	assert.False(s2.Eq(s1))

	// matrix
	s1 = Shape{2, 3}
	assert.True(s1.Eq(s2))
	assert.True(s2.Eq(s1))

	s2 = Shape{3, 2}
	assert.False(s1.Eq(s2))
	assert.False(s2.Eq(s1))

	// just for that green coloured code
	s1 = Shape{2}
	s2 = Shape{1, 3}
	assert.False(s1.Eq(s2))
	assert.False(s2.Eq(s1))
}

var shapeSliceTests = []struct {
	name string
	s    Shape
	sli  []Slice

	expected Shape
	err      bool
}{
	{"slicing a scalar shape", ScalarShape(), nil, ScalarShape(), false},
	{"slicing a scalar shape", ScalarShape(), []Slice{Range{0, 0, 0}}, nil, true},
	{"vec[0]", Shape{2}, []Slice{Range{0, 1, 1}}, ScalarShape(), false},
	{"vec[3]", Shape{2}, []Slice{Range{3, 4, 1}}, nil, true},
	{"vec[:, 0]", Shape{2}, []Slice{nil, Range{0, 1, 1}}, nil, true},
	{"vec[1:4:2]", Shape{5}, []Slice{Range{1, 4, 2}}, ScalarShape(), false},
	{"tensor[0, :, :]", Shape{1, 2, 2}, []Slice{Range{0, 1, 1}, nil, nil}, Shape{2, 2}, false},
	{"tensor[:, 0, :]", Shape{1, 2, 2}, []Slice{nil, Range{0, 1, 1}, nil}, Shape{1, 2}, false},
	{"tensor[0, :, :, :]", Shape{1, 1, 2, 2}, []Slice{Range{0, 1, 1}, nil, nil, nil}, Shape{1, 2, 2}, false},
	{"tensor[0,]", Shape{1, 1, 2, 2}, []Slice{Range{0, 1, 1}}, Shape{1, 2, 2}, false},
}

func TestShape_Slice(t *testing.T) {
	for i, ssts := range shapeSliceTests {
		newShape, err := ssts.s.S(ssts.sli...)
		if checkErr(t, ssts.err, err, "Shape slice", i) {
			continue
		}
		shp, ok := newShape.(Shape)
		if !ok {
			t.Errorf("Test %v Expected newShape to be a Shape. Got %v of %T instead.", ssts.name, newShape, newShape)
			continue
		}

		if !ssts.expected.Eq(shp) {
			t.Errorf("Test %q: Expected shape %v. Got %v instead", ssts.name, ssts.expected, newShape)
		}
	}
}

var shapeRepeatTests = []struct {
	name    string
	s       Shape
	repeats []int
	axis    Axis

	expected        Shape
	expectedRepeats []int
	expectedSize    int
	err             bool
}{

	{"scalar repeat on axis 0", ScalarShape(), []int{3}, 0, Shape{3}, []int{3}, 1, false},
	{"scalar repeat on axis 1", ScalarShape(), []int{3}, 1, Shape{1, 3}, []int{3}, 1, false},
	{"vector repeat on axis 0", Shape{2}, []int{3}, 0, Shape{6}, []int{3, 3}, 2, false},
	{"vector repeat on axis 1", Shape{2}, []int{3}, 1, Shape{2, 3}, []int{3}, 1, false},
	{"colvec repeats on axis 0", Shape{2, 1}, []int{3}, 0, Shape{6, 1}, []int{3, 3}, 2, false},
	{"colvec repeats on axis 1", Shape{2, 1}, []int{3}, 1, Shape{2, 3}, []int{3}, 1, false},
	{"rowvec repeats on axis 0", Shape{1, 2}, []int{3}, 0, Shape{3, 2}, []int{3}, 1, false},
	{"rowvec repeats on axis 1", Shape{1, 2}, []int{3}, 1, Shape{1, 6}, []int{3, 3}, 2, false},
	{"3-Tensor repeats", Shape{2, 3, 2}, []int{1, 2, 1}, 1, Shape{2, 4, 2}, []int{1, 2, 1}, 3, false},
	{"3-Tensor generic repeats", Shape{2, 3, 2}, []int{2}, AllAxes, Shape{24}, []int{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}, 12, false},
	{"3-Tensor generic repeat, axis specified", Shape{2, 3, 2}, []int{2}, 2, Shape{2, 3, 4}, []int{2, 2}, 2, false},

	// stupids
	{"nonexisting axis 2", Shape{2, 1}, []int{3}, 2, nil, nil, 0, true},
	{"mismatching repeats", Shape{2, 3, 2}, []int{3, 1, 2}, 0, nil, nil, 0, true},
}

func TestShape_Repeat(t *testing.T) {
	assert := assert.New(t)
	for i, srts := range shapeRepeatTests {
		newShape, reps, size, err := srts.s.Repeat(srts.axis, srts.repeats...)

		if checkErr(t, srts.err, err, srts.name, i) {
			continue
		}
		shp, ok := newShape.(Shape)
		if !ok {
			t.Errorf("Test %v. Expected newShape to be Shape. Got %v of %T instead", srts.name, newShape, newShape)
			continue
		}

		assert.True(srts.expected.Eq(shp), "Test %q:  Want: %v. Got %v", srts.name, srts.expected, newShape)
		assert.Equal(srts.expectedRepeats, reps, "Test %q - Expected Repeats %v. Got %v ", srts.name, srts.expectedRepeats, reps)
		assert.Equal(srts.expectedSize, size, "Test %q: Expected size: %d. Got %d ", srts.name, srts.expectedSize, size)
	}
}

var shapeConcatTests = []struct {
	name string
	s    Shape
	axis Axis
	ss   []Shape

	expected Shape
	err      bool
}{
	{"standard, axis 0 ", Shape{2, 2}, 0, []Shape{{2, 2}, {2, 2}}, Shape{6, 2}, false},
	{"standard, axis 1 ", Shape{2, 2}, 1, []Shape{{2, 2}, {2, 2}}, Shape{2, 6}, false},
	{"standard, axis AllAxes ", Shape{2, 2}, AllAxes, []Shape{{2, 2}, {2, 2}}, Shape{6, 2}, false},
	{"concat to empty", Shape{2}, 0, nil, Shape{2}, false},

	{"stupids: different dims", Shape{2, 2}, 0, []Shape{{2, 3, 2}}, nil, true},
	{"stupids: negative axes", Shape{2, 2}, -5, []Shape{{2, 2}}, nil, true},
	{"stupids: toobig axis", Shape{2, 2}, 5, []Shape{{2, 2}}, nil, true},
	{"subtle stupids: dim mismatch", Shape{2, 2}, 0, []Shape{{2, 2}, {2, 3}}, nil, true},
}

func TestShape_Concat(t *testing.T) {
	assert := assert.New(t)
	for _, scts := range shapeConcatTests {
		sls := ShapesToShapelikes(scts.ss)
		newShape, err := scts.s.Concat(scts.axis, sls...)
		switch {
		case scts.err:
			if err == nil {
				t.Error("Expected an error")
			}
			continue
		case !scts.err && err != nil:
			t.Errorf("Test %v err %v", scts.name, err)
			continue
		}
		assert.Equal(scts.expected, newShape)
	}
}

var shapeDimTests = []struct {
	name string
	s    Shape
	dim  int

	expected int
	err      bool
}{
	{"standard behaviour - 0", Shape{2, 3, 4}, 0, 2, false},
	{"standard behaviour - 1", Shape{2, 3, 4}, 1, 3, false},
	{"standard behaviour - 2", Shape{2, 3, 4}, 2, 4, false},
	{"non-standard behaviour - -1", Shape{2, 3, 4}, -1, 4, false},
	{"non-standard behaviour - -2", Shape{2, 3, 4}, -2, 3, false},
	{"non-standard behaviour - -3", Shape{2, 3, 4}, -3, 2, false},
	{"standard on scalar", Shape{}, 0, 0, false},

	// naughty
	{"scalar but dim ≠ 0", Shape{}, -1, -1, true},
	{"d > dims", Shape{2, 3, 4}, 3, -1, true},
	{"d < -dims", Shape{2, 3, 4}, -4, -1, true},
}

func TestShape_Dim(t *testing.T) {
	assert := assert.New(t)
	for _, sdt := range shapeDimTests {
		t.Run(sdt.name, func(t *testing.T) {
			sz, err := sdt.s.Dim(sdt.dim)
			switch {
			case sdt.err:
				assert.NotNil(err, "%v should cause an error", sdt.name)
				return
			default:
				if !assert.Nil(err, "%v should not have an error", sdt.name) {
					return
				}
				assert.Equal(sdt.expected, sz)
			}
		})
	}
}
