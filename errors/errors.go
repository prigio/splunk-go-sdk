package errors

import (
	"fmt"
)

type ErrInvalidParam struct {
	Context string // function where error happened
	Msg     string // details message
	Err     error  // wrapped error
}

func (e *ErrInvalidParam) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("%s: invalid parameter %s", e.Context, e.Msg)
	}
	return fmt.Sprintf("%s: invalid parameter %s. %v", e.Context, e.Msg, e.Err)
}

// Unwrap
// https://github.com/golang/go/blob/release-branch.go1.17/src/errors/wrap.go
func (e *ErrInvalidParam) Unwrap() error {
	return e.Err
}

func NewErrInvalidParam(context string, err error, msg string, a ...interface{}) error {
	return &ErrInvalidParam{
		Context: context,
		Err:     err,
		Msg:     fmt.Sprintf(msg, a...),
	}
}
