package upgrader

import "fmt"

func newInstanceError(message string) error {
	return InstanceError{message: message}
}

func newInstanceErrorf(format string, args ...any) error {
	return InstanceError{message: fmt.Sprintf(format, args...)}
}

type InstanceError struct {
	message string
}

func (e InstanceError) Error() string {
	return e.message
}
