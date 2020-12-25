package sequencepool

import "errors"

// Errors holds package-level variables that represent different errors related
// to batches and the transaction pool.
var (
	// ErrNil means that there was a nil pointer dereference
	ErrNil = errors.New("nil pointer dereference")

	// ErrNotFoundElement means no element found for specified key
	ErrNotFoundElement = errors.New("not found element")

	// ErrMismatchElement means found a duplicate element with different key
	ErrMismatchElement = errors.New("found mismatch element")
)
