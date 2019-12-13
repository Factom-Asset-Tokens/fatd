package runtime

import "fmt"

const ErrorExecLimitExceededString = "Execution limit exceeded."

type ErrorExecLimitExceeded struct {
	Func string
}

func (err ErrorExecLimitExceeded) Error() string {
	return ErrorExecLimitExceededString
}

type ErrorRevert struct {
	Reason string
}

func (err ErrorRevert) Error() string {
	return fmt.Sprintf("revert: %v", err.Reason)
}

type ErrorSelfDestruct struct{}

func (err ErrorSelfDestruct) Error() string {
	return "contract self destruct"
}
