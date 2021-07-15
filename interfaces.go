package shapes

var (
	_ Shapelike = Abstract{}
	_ Shapelike = Shape{}
)

//go-sumtype:decl Sizelike

// Shapelike is anything that performs all the things you can do with a Shape.
// The following types provided by this library are Shapelike:
// 	Shape | Abstract
type Shapelike interface {
	Dims() int
	TotalSize() int // needed?
	DimSize(dim int) (Sizelike, error)
	T(axes ...Axis) (newShape Shapelike, err error)
	S(slices ...Slice) (newShape Shapelike, err error)
	Repeat(axis Axis, repeats ...int) (newShape Shapelike, finalRepeats []int, size int, err error)
	Concat(axis Axis, others ...Shapelike) (newShape Shapelike, err error)
}

// intslike is anything that can return a []int
// The following types provided byt this library are intslike:
// 	Shape | Sizes | Axes
type intslike interface {
	AsInts() []int
}

// Shaper is anything that can return a Shape.
type Shaper interface {
	Shape() Shape
}

// Exprer is anything that can return a Shape Expr.
type Exprer interface {
	Shape() Expr
}

var (
	_ Sizelike = Size(0)
	_ Sizelike = Var('a')
	_ Sizelike = BinOp{}
	_ Sizelike = UnaryOp{}
)

// Sizelike represents something that can go into a Abstract. The following types are Sizelike:
// 	Size | Var | BinOp | UnaryOp
type Sizelike interface {
	isSizelike()
}

// Conser is anything that can be used to construct another Conser. The following types are Conser:
//	Shape | Abstract
type Conser interface {
	Cons(Conser) Conser
	isConser()
}

// substitutable is anything that can apply a list of subsitution and then return a substitutable.
//
// The following implements substitutable:
//
// Expressions:
// 	Shape | Abstrct | Arrow | Sizes | Size | Axes | Axis | Var
// Operations:
// 	BinaryOp | UnaryOp
//	RepeatOf | ConcatOf | SliceOf | TransposeOf | IndexOf
// Compound expressions:
//	Compound | SubjectTo
// Constraints:
//	constraints | exprConstraint
type substitutable interface {
	apply(substitutions) substitutable
	freevars() varset // set of free variables
}

// same as substitutable, except doesn't apply to internal constraints (exprConstraint and constraints)
type substitutableExpr interface {
	substitutable
	subExprs() []substitutableExpr
}

// Operation represents an operation (BinOp or UnaryOp)
type Operation interface {
	isValid() bool
	substitutableExpr
}

type boolOp interface {
	Operation
	resolveBool() (bool, error)
}

type sizeOp interface {
	Operation
	resolveSize() (Size, error)
}

//  resolver is anything that can resolve an expression
//
// e.g. "built-in" unary terms like TransposeOf, ConcatOf, SliceOf, RepeatOf
type resolver interface {
	resolve() (Expr, error)
}

// Slicelike is anything like a slice. The following types implement Slicelike:
// 	Range | Var
type Slicelike interface {
	substitutableExpr
	isSlicelike()
}
