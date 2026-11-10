package application

import "github.com/thiagosantos/dynamodb-manager/backend/internal/domain"

type ErrorCode = domain.ErrorCode
type Error = domain.Error

const (
	InvalidArgument = domain.InvalidArgument
	NotFound        = domain.NotFound
	Conflict        = domain.Conflict
	Upstream        = domain.Upstream
	Internal        = domain.Internal
)

func invalid(message string) error  { return domain.InvalidInput(message) }
func notFound(message string) error { return domain.NotFoundError(message) }
func internal(message string, cause error) error {
	return domain.InternalError(message, cause)
}
func upstream(cause error) error { return domain.UpstreamError(cause) }
