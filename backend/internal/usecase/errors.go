package usecase

import "fmt"

type ErrorCode string

const (
	InvalidArgument ErrorCode = "invalid_argument"
	NotFound        ErrorCode = "not_found"
	Conflict        ErrorCode = "conflict"
	Upstream        ErrorCode = "upstream_error"
	Internal        ErrorCode = "internal"
)

type Error struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Cause }

func invalid(message string) error  { return &Error{Code: InvalidArgument, Message: message} }
func notFound(message string) error { return &Error{Code: NotFound, Message: message} }
func internal(message string, cause error) error {
	return &Error{Code: Internal, Message: message, Cause: cause}
}
func upstream(cause error) error {
	return &Error{Code: Upstream, Message: "DynamoDB request failed", Cause: cause}
}
