package shapes

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbstract_T(t *testing.T) {
	assert := assert.New(t)
	abstract := Abstract{Size(1), BinOp{Add, Size(1), Size(2)}}

	// noop
	a2, err := abstract.T(0, 1)
	if err == nil {
		t.Errorf("Expected a noop error")
	}
	if _, ok := err.(NoOpError); !ok {
		t.Errorf("Expected a noop error. Got %v instead", err)
	}
	assert.Equal(a2, abstract)

	a2, err = abstract.T(1, 0)
	if err != nil {
		t.Fatal(err)
	}
	correct := Abstract{abstract[1], abstract[0]}
	assert.Equal(correct, a2)
}

var absSliceTests = []struct {
	name string
	a    Abstract
	sli  []Slice

	expected Shapelike
	err      bool
}{
	{"all vars", Gen(2), nil, Gen(2), false},
	{"all vars", Gen(2), []Slice{nil, S(1)},
		Abstract{Var('a'), sizelikeSliceOf{SliceOf{*S(1), Var('b')}}}, false},

	{"all sizes (vector)", Abstract{Size(2)}, []Slice{S(0)}, ScalarShape(), false},
	{"all sizes (vector) - bad slice range", Abstract{Size(2)}, []Slice{S(3)}, nil, true},

	{"Mixed sizes and var", Abstract{Var('a'), Size(2)}, []Slice{S(2), S(0, 2)},
		Abstract{sizelikeSliceOf{SliceOf{*S(2), Var('a')}}, Size(2)},
		false,
	},
	{"Mixed",
		Abstract{Var('a'), BinOp{Add, Var('a'), Var('b')}, UnaryOp{Dims, Var('b')}},
		[]Slice{S(1, 5, 2), S(1, 5), S(1, 5)},
		Abstract{
			sizelikeSliceOf{SliceOf{*S(1, 5, 2), Var('a')}},
			sizelikeSliceOf{SliceOf{*S(1, 5), exprBinOp{BinOp{Add, Var('a'), Var('b')}}}},
			sizelikeSliceOf{SliceOf{*S(1, 5), UnaryOp{Dims, Var('b')}}},
		},
		false,
	},
}

func TestAbstract_S(t *testing.T) {
	assert := assert.New(t)
	for i, c := range absSliceTests {
		newShapelike, err := c.a.S(c.sli...)
		if checkErr(t, c.err, err, "Abs slice", i) {
			continue
		}
		assert.Equal(c.expected, newShapelike, c.name)
	}
}

func ExampleSlice_s() {
	param0 := Abstract{Var('a'), Var('b')}
	param1 := Abstract{Var('a'), Var('b'), BinOp{Add, Var('a'), Var('b')}, UnaryOp{Const, Var('b')}}
	expected, err := param1.S(S(1, 5), S(1, 5), S(1, 5), S(2, 5))
	if err != nil {
		fmt.Printf("Err %v\n", err)
		return
	}
	expr := MakeArrow(param0, param1, expected.(Expr))
	fmt.Printf("expr: %v\n", expr)

	fst := Shape{10, 20}
	result, err := InferApp(expr, fst)
	if err != nil {
		fmt.Printf("Err %v\n", err)
		return
	}
	fmt.Printf("%v @ %v ↠ %v\n", expr, fst, result)

	snd := Shape{10, 20, 30, 20}
	result2, err := InferApp(result, snd)
	if err != nil {
		fmt.Printf("Err %v\n", err)
		return
	}
	fmt.Printf("%v @ %v ↠ %v", result, snd, result2)

	// Output:
	// expr: (a, b) → (a, b, a + b, K b) → (a[1:5], b[1:5], a + b[1:5], K b[2:5])
	// (a, b) → (a, b, a + b, K b) → (a[1:5], b[1:5], a + b[1:5], K b[2:5]) @ (10, 20) ↠ (10, 20, 30, 20) → (4, 4, 4, 3)
	// (10, 20, 30, 20) → (4, 4, 4, 3) @ (10, 20, 30, 20) ↠ (4, 4, 4, 3)

}

var absRepeatTests = []struct {
	name    string
	a       Abstract
	repeats []int
	axis    Axis

	expected        Shapelike
	expectedRepeats []int
	expectedSize    int
	err             bool
}{
	{"vector repeat on axis 0", Abstract{Var('a')}, []int{3}, 0, Abstract{BinOp{Mul, Var('a'), Size(3)}}, []int{3}, -1, false},
	{"vector repeat on axis 1", Abstract{Var('a')}, []int{3}, 1, Abstract{Var('a'), Size(3)}, []int{3}, 1, false},
	{"var matrix repeat on axis 0", Abstract{Var('a'), Var('b')}, []int{1, 3}, 0, Abstract{Size(4), Var('b')}, nil, -1, false},
	{"var matrix repeat on axis 1", Abstract{Var('a'), Var('b')}, []int{1, 3}, 1, Abstract{Var('a'), Size(4)}, nil, -1, false},
	{"var matrix generic repeat on axis 0", Abstract{Var('a'), Var('b')}, []int{3}, 0, Abstract{BinOp{Mul, Var('a'), Size(3)}, Var('b')}, []int{3}, -1, false},
	{"var matrix generic repeat on axis 1", Abstract{Var('a'), Var('b')}, []int{3}, 1, Abstract{Var('a'), BinOp{Mul, Var('b'), Size(3)}}, []int{3}, -1, false},
}

func TestAbs_Repeat(t *testing.T) {
	assert := assert.New(t)
	for i, c := range absRepeatTests {
		newShape, reps, size, err := c.a.Repeat(c.axis, c.repeats...)
		if checkErr(t, c.err, err, c.name, i) {
			continue
		}

		assert.Equal(c.expected, newShape, "Test %v - Shape like not the same", c.name)
		assert.Equal(c.expectedRepeats, reps, "Test %v - Repeats not the same", c.name)
		assert.Equal(c.expectedSize, size, "Test %v - Size not the same", c.name)
	}
}
