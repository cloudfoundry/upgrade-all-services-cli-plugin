package upgrader

import (
	"errors"
)

type MultiError struct {
	Errors []error
}

// Append adds a non-nil error to the MultiError.
// If the provided error is nil, it will be ignored, preventing
// nil errors from being included in the error slice.
// This method simplifies error aggregation by centralizing nil checks.
func (m *MultiError) Append(err error) {
	if err != nil {
		m.Errors = append(m.Errors, err)
	}
}

// Error implements the error interface. It returns a concatenated
// string of all the non-nil errors it contains. If there are no non-nil
// errors, it returns an empty string.
// This method allows MultiError to seamlessly integrate with typical error handling in Go.
func (m *MultiError) Error() string {
	return errors.Join(m.Errors...).Error()
}
