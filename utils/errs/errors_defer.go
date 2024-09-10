package errs

import (
	"errors"
	"fmt"
)

// Capture runs errFunc and assigns the error, if any, to *errPtr. Preserves the
// original error by wrapping with errors.Join if the errFunc err is non-nil.
func Capture(errPtr *error, errFunc func() error, msg string) {
	err := errFunc()
	if err == nil {
		return
	}
	*errPtr = errors.Join(*errPtr, fmt.Errorf("%s: %w", msg, err))
}

// CaptureGeneric runs errFunc with a generic type K and assigns the error, if any, to *errPtr.
func CaptureGeneric[K any](errPtr *error, errFunc func(value K) error, value K, msg string) {
	err := errFunc(value)
	if err == nil {
		return
	}
	*errPtr = errors.Join(*errPtr, fmt.Errorf("%s: %w", msg, err))
}
