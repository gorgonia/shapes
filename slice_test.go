package shapes

import "testing"

type MySli int

func (s MySli) Start() int { return int(s) }
func (s MySli) End() int   { return int(s) + 1 }
func (s MySli) Step() int  { return 1 }

var _ Slice = MySli(1)

func TestToRange(t *testing.T) {
	var cases = []Slice{
		MySli(1),
		S(1),
		S(1, 2),
		S(1, 2, 3),
		&Range{1, 2, 3},
	}
	var correct = []Range{
		Range{1, 2, 1},
		Range{1, 2, 1},
		Range{1, 2, 1},
		Range{1, 2, 3},
		Range{1, 2, 3},
	}
	for i, c := range cases {
		r := toRange(c)
		if r != correct[i] {
			t.Errorf("Expected case %d to be correct: %#v. Got %#v instead.", i, c, r)
		}
	}
}
