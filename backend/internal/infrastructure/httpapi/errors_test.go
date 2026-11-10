package httpapi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/smithy-go"
	"github.com/thiagosantos/dynamodb-manager/backend/internal/application"
)

func TestErrorsAreSanitizedAndConsistent(t *testing.T) {
	response := httptest.NewRecorder()
	writeError(response, &smithy.GenericAPIError{Code: "AccessDeniedException", Message: "secret token and credential detail"})
	if response.Code != http.StatusForbidden {
		t.Fatalf("status=%d", response.Code)
	}
	if strings.Contains(response.Body.String(), "secret token") || !strings.Contains(response.Body.String(), `"error"`) {
		t.Fatalf("error was not sanitized: %s", response.Body.String())
	}
	response = httptest.NewRecorder()
	writeError(response, &application.Error{Code: application.Upstream, Message: "DynamoDB request failed", Cause: errors.New("endpoint contains private details")})
	if response.Code != http.StatusBadGateway || strings.Contains(response.Body.String(), "private details") {
		t.Fatalf("upstream error was not mapped safely: status=%d body=%s", response.Code, response.Body.String())
	}
}
