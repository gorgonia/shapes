package shapes

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

var lexCases = map[string][]tok{
	"()":         []tok{{parenL, '(', 0}, {parenR, ')', 1}},
	"(a,)":       []tok{{parenL, '(', 0}, {letter, 'a', 1}, {comma, ',', 2}, {parenR, ')', 3}},
	"(1, 2, 34)": []tok{{parenL, '(', 0}, {digit, 1, 1}, {comma, ',', 2}, {digit, 2, 4}, {comma, ',', 5}, {digit, 34, 8}, {parenR, ')', 9}},
	"() -> ()":   []tok{{parenL, '(', 0}, {parenR, ')', 1}, {arrow, '→', 4}, {parenL, '(', 6}, {parenR, ')', 7}},
	"() → ()":    []tok{{parenL, '(', 0}, {parenR, ')', 1}, {arrow, '→', 3}, {parenL, '(', 5}, {parenR, ')', 6}},

	"1000": []tok{{digit, 1000, 3}},

	// unop
	"P a": []tok{{unop, 'Π', 0}, {letter, 'a', 2}},
	"S a": []tok{{unop, 'Σ', 0}, {letter, 'a', 2}},

	// binop, cmpop and logop
	"a + 1":  []tok{{letter, 'a', 0}, {binop, '+', 2}, {digit, 1, 4}},
	"a - 1":  []tok{{letter, 'a', 0}, {binop, '-', 2}, {digit, 1, 4}},
	"a = 2":  []tok{{letter, 'a', 0}, {cmpop, '=', 2}, {digit, 2, 4}},
	"a != 2": []tok{{letter, 'a', 0}, {cmpop, '≠', 3}, {digit, 2, 5}},
	"a > 1":  []tok{{letter, 'a', 0}, {cmpop, '>', 2}, {digit, 1, 4}},
	"a >= 1": []tok{{letter, 'a', 0}, {cmpop, '≥', 3}, {digit, 1, 5}},
	"a <= 1": []tok{{letter, 'a', 0}, {cmpop, '≤', 3}, {digit, 1, 5}},
	"a ≥ 1":  []tok{{letter, 'a', 0}, {cmpop, '≥', 2}, {digit, 1, 4}},
	"a ∧ 1":  []tok{{letter, 'a', 0}, {logop, '∧', 2}, {digit, 1, 4}},
	"a && 1": []tok{{letter, 'a', 0}, {logop, '∧', 3}, {digit, 1, 5}},
	"a || 1": []tok{{letter, 'a', 0}, {logop, '∨', 3}, {digit, 1, 5}},

	// constructions
	"{(a) -> () | (a > 2)}": []tok{
		{braceL, '{', 0},
		{parenL, '(', 1},
		{letter, 'a', 2},
		{parenR, ')', 3},
		{arrow, '→', 6},
		{parenL, '(', 8},
		{parenR, ')', 9},
		{pipe, '|', 11},
		{parenL, '(', 13},
		{letter, 'a', 14},
		{cmpop, '>', 16},
		{digit, 2, 18},
		{parenR, ')', 19},
		{braceR, '}', 20},
	},
	"[0:2:1]": []tok{{brackL, '[', 0}, {digit, 0, 1}, {colon, ':', 2}, {digit, 2, 3}, {colon, ':', 4}, {digit, 1, 5}, {brackR, ']', 6}},

	// dubious API design wise
	"& a": []tok{{letter, 'a', 2}}, // note that the singular '&' is ignored.
}

func TestLex(t *testing.T) {
	for k, v := range lexCases {
		toks, err := lex(k)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, v, toks)
	}
}

var parseCases = map[string]Expr{

	"()":     Shape{},
	"(1,)":   Shape{1},
	"(a,b,)": Abstract{Var('a'), Var('b')},

	"(1,2,3,2325)": Shape{1, 2, 3, 2325},
	"(1, a, 2)":    Abstract{Size(1), Var('a'), Size(2)},

	"(a,b,c) → (a*b, b*c)": Arrow{
		Abstract{Var('a'), Var('b'), Var('c')},
		Abstract{
			BinOp{Mul, Var('a'), Var('b')},
			BinOp{Mul, Var('b'), Var('c')},
		},
	},

	// higher order functions, because why not
	"(a -> b) -> a -> b": Arrow{
		Arrow{Var('a'), Var('b')},
		Arrow{Var('a'), Var('b')},
	},

	// SubjectTo clause
	"{a -> b | (D a = D b)}": Compound{
		Expr: Arrow{Var('a'), Var('b')},
		SubjectTo: SubjectTo{
			Eq,
			UnaryOp{Dims, Var('a')},
			UnaryOp{Dims, Var('b')},
		},
	},
	// Axes
	"X[0 1 3 2]": Axes{0, 1, 3, 2},

	// TransposeOf
	"T X[1 0] a": TransposeOf{
		Axes{1, 0},
		Var('a'),
	},

	// Transpose:
	"{ a → X[0 1 3 2] → T X[0 1 3 2] a | (D X[0 1 3 2] = D a) }": Compound{
		Expr: Arrow{
			Var('a'),
			Arrow{
				Axes{0, 1, 3, 2},
				TransposeOf{
					Axes{0, 1, 3, 2},
					Var('a'),
				},
			},
		},
		SubjectTo: SubjectTo{
			Eq,
			UnaryOp{Dims, Axes{0, 1, 3, 2}},
			UnaryOp{Dims, Var('a')},
		},
	},

	// Indexing
	"a → b -> ()": Arrow{
		Var('a'),
		Arrow{Var('b'), Shape{}},
	},

	// Indexing (constrained)
	"{ a → b → () | ((D a = D b) ∧ (∀ b < ∀ a)) }": Compound{
		Expr: Arrow{
			Var('a'),
			Arrow{
				Var('b'),
				Shape{},
			},
		},
		SubjectTo: SubjectTo{
			And,
			SubjectTo{
				Eq,
				UnaryOp{Dims, Var('a')},
				UnaryOp{Dims, Var('b')},
			},
			SubjectTo{
				Lt,
				UnaryOp{ForAll, Var('b')},
				UnaryOp{ForAll, Var('a')},
			},
		},
	},

	// Slicing
	"{ a → [0:2] → a[0:2] | (a[0] ≥ 2) }": Compound{
		Expr: Arrow{
			Var('a'),
			Arrow{
				Range{0, 2, 1},
				SliceOf{
					Range{0, 2, 1},
					Var('a'),
				},
			},
		},
		SubjectTo: SubjectTo{
			OpType: Gte,
			A:      IndexOf{I: 0, A: Var('a')},
			B:      Size(2),
		},
	},

	// Slicing with steps
	"{ a → [0:2:1] → a[0:2:1] | (a[0] ≥ 2) }": Compound{
		Expr: Arrow{
			Var('a'),
			Arrow{
				Range{0, 2, 1},
				SliceOf{
					Range{0, 2, 1},
					Var('a'),
				},
			},
		},
		SubjectTo: SubjectTo{
			OpType: Gte,
			A:      IndexOf{I: 0, A: Var('a')},
			B:      Size(2),
		},
	},

	// Slicing with single slice
	"{ a → [0] → a[0] | (a[0] ≥ 2) }": Compound{
		Expr: Arrow{
			Var('a'),
			Arrow{
				Range{0, 1, 1},
				IndexOf{
					0,
					Var('a'),
				},
			},
		},
		SubjectTo: SubjectTo{
			OpType: Gte,
			A:      IndexOf{I: 0, A: Var('a')},
			B:      Size(2),
		},
	},

	// Reshaping
	"{ a → b → b | (Π a = Π b) }": Compound{
		Arrow{
			Var('a'),
			Arrow{
				Var('b'),
				Var('b'),
			},
		},
		SubjectTo{
			Eq,
			UnaryOp{Prod, Var('a')},
			UnaryOp{Prod, Var('b')},
		},
	},

	// Columnwise Sum Matrix
	"{ a → b | (D b = D a - 1) }": Compound{
		Arrow{
			Var('a'),
			Var('b'),
		},
		SubjectTo{
			Eq,
			UnaryOp{Dims, Var('b')},
			BinOp{
				Sub,
				UnaryOp{Dims, Var('a')},
				Size(1),
			},
		},
	},

	// Weird but acceptable inputs for Shapes and Abstracts
	"(1, (a,))":     Abstract{Size(1), Var('a')},
	"(1, (a, b,),)": Abstract{Size(1), Var('a'), Var('b')},
	"(1, (2,3),)":   Shape{1, 2, 3},
	"((1,), b)":     Abstract{Size(1), Var('b')},
	"((1,), 2)":     Shape{1, 2},
	"((1), (2))":    Shape{1, 2},

	"((1,), (a,))":   Abstract{Size(1), Var('a')},
	"((1,), (a,b,))": Abstract{Size(1), Var('a'), Var('b')},

	"((a, b), c)":     Abstract{Var('a'), Var('b'), Var('c')},
	"((), a)":         Abstract{Var('a')},
	"((a,b), (c, d))": Abstract{Var('a'), Var('b'), Var('c'), Var('d')},

	// please don't write something like this.
	"(),(0)": Shape{0},
}

/*
var knownFail = map[string]Expr{
	"(a,b,c) → (a*b+c, a*b+c)": Arrow{
		Abstract{Var('a'), Var('b'), Var('c')},
		Abstract{
			BinOp{Add, BinOp{Mul, Var('a'), Var('b')}, Var('c')},
			BinOp{Add, BinOp{Mul, Var('a'), Var('b')}, Var('c')},
		},
	},

   // open ended slicing
   "{ a → [1:] → a[1:] | (a[0] ≥ 2) }": Compound{
		Expr: Arrow{
			Var('a'),
			Arrow{
				Sli{0, 2, 1},
				SliceOf{
					Sli{0, 2, 1},
					Var('a'),
				},
			},
		},
		SubjectTo: SubjectTo{
			OpType: Gte,
			A:      IndexOf{I: 0, A: Var('a')},
			B:      Size(2),
		},
	},


}
*/

func TestParse(t *testing.T) {
	for k, v := range parseCases {
		expr, err := Parse(k)
		if err != nil {
			t.Fatalf("Unable to parse %q: %+v", k, err)
		}
		assert.Equal(t, v, expr, "Failed to parse %q", k)
	}
}

var badInputs = []string{

	"X1000",
	"0,0,0)0(0P0b0)0,0,0)0T",
	"0->0->->b[",
	">]]>0",
	":",
	"-0",
	"1<",
	"1∧",
	"{|}",
	"0*0",
	"0,>(0)",
	"0,>-0a->",
	"TX",
	"(",
	"(,y)0}0=",
	",c[SS[S ->S0",
	"0TX[",
	"[->]",
	"(0[])",
	"(a->a[]7476837158203120)",
	"(Y0-0)0{TX[0]",
	"TX[]-00(0)",
	"TX[]TX[]TX[][]",

	//	"0{(-0O)0||*0(|)}",
}

func TestParseBadInputs(t *testing.T) {
	var in string
	defer func() {
		if r := recover(); r != nil {
			log.Printf("in: %q", in)
			panic(r)
		}
	}()

	for _, in = range badInputs {
		if wtf, err := Parse(in); err == nil {
			t.Errorf("Expected errors when parsing %q. Got %v", in, wtf)
		}
	}
}
