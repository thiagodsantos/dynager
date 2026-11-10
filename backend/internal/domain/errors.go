package domain

import (
	"errors"
	"fmt"
)

type ErrorCode string

type Error struct {
	Code    ErrorCode
	Message string
	Cause   error
}

const (
	InvalidArgument ErrorCode = "invalid_argument"
	NotFound        ErrorCode = "not_found"
	Conflict        ErrorCode = "conflict"
	Upstream        ErrorCode = "upstream_error"
	Internal        ErrorCode = "internal"
)

var errKeyValueType = errors.New("must be a string, number, or binary value")

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Cause }

func InvalidInput(message string) error  { return &Error{Code: InvalidArgument, Message: message} }
func NotFoundError(message string) error { return &Error{Code: NotFound, Message: message} }
func ConflictError(message string) error { return &Error{Code: Conflict, Message: message} }
func InternalError(message string, cause error) error {
	return &Error{Code: Internal, Message: message, Cause: cause}
}
func UpstreamError(cause error) error {
	return &Error{Code: Upstream, Message: "DynamoDB request failed", Cause: cause}
}
