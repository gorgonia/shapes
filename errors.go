package shapes

const (
	dimsMismatch      = "Dimension mismatch. Expected %v. Got  %v instead."
	invalidAxis       = "Invalid axis %d for ndarray with %d dimensions."
	repeatedAxis      = "repeated axis %d in permutation pattern."
	invalidSliceIndex = "Invalid slice index. Start: %d, End: %d."
	unaryOpResolveErr = "Cannot resolve %v to a Size."
	broadcastErr      = "Cannot broadcast %v with %v. %d-th dimension does not match or is not a 1."
)

// NoOpError is a useful for operations that have no op.
type NoOpError interface {
	NoOp() bool
}

type noopError struct{}

func (e noopError) NoOp() bool    { return true }
func (e noopError) Error() string { return "NoOp" }

const (
	broadcastError = "Canot broadcast together. Resulting shape will be at least (%d, 1). Repeats is (%d, 1)"
	dimMismatch    = "Dimension mismatch. Expected %d, got %d"
)
