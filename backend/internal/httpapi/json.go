package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/aws/smithy-go"
	"github.com/thiagosantos/dynamodb-manager/backend/internal/app"
)

type errorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeErrorResponse(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorEnvelope{Error: apiError{Code: code, Message: message}})
}

func writeError(w http.ResponseWriter, err error) {
	var apiError smithy.APIError
	if errors.As(err, &apiError) {
		switch apiError.ErrorCode() {
		case "ResourceNotFoundException":
			writeErrorResponse(w, http.StatusNotFound, "resource_not_found", "DynamoDB resource not found")
		case "AccessDeniedException", "UnrecognizedClientException", "InvalidSignatureException":
			writeErrorResponse(w, http.StatusForbidden, "aws_access_denied", "AWS denied the request; check the selected profile's permissions")
		case "ValidationException":
			writeErrorResponse(w, http.StatusBadRequest, "aws_validation_error", "DynamoDB rejected the request")
		case "ResourceInUseException":
			writeErrorResponse(w, http.StatusConflict, "resource_conflict", "DynamoDB resource is in use")
		case "ProvisionedThroughputExceededException", "ThrottlingException", "RequestLimitExceeded":
			writeErrorResponse(w, http.StatusTooManyRequests, "aws_throttled", "DynamoDB request was throttled; retry later")
		default:
			writeErrorResponse(w, http.StatusBadGateway, "aws_unavailable", "DynamoDB request failed")
		}
		return
	}
	var applicationError *app.Error
	if errors.As(err, &applicationError) {
		switch applicationError.Code {
		case app.InvalidArgument:
			writeErrorResponse(w, http.StatusBadRequest, string(applicationError.Code), applicationError.Message)
		case app.NotFound:
			writeErrorResponse(w, http.StatusNotFound, string(applicationError.Code), applicationError.Message)
		case app.Conflict:
			writeErrorResponse(w, http.StatusConflict, string(applicationError.Code), applicationError.Message)
		case app.Upstream:
			writeErrorResponse(w, http.StatusBadGateway, string(applicationError.Code), "DynamoDB request failed")
		default:
			writeErrorResponse(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		}
		return
	}
	var operationError *smithy.OperationError
	if errors.As(err, &operationError) {
		writeErrorResponse(w, http.StatusBadGateway, "aws_unavailable", "DynamoDB request failed")
		return
	}
	writeErrorResponse(w, http.StatusInternalServerError, "internal_error", "Internal server error")
}

func writeDecodeError(w http.ResponseWriter, err error, message string) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		writeErrorResponse(w, http.StatusRequestEntityTooLarge, "request_too_large", "Request body exceeds the 2 MiB limit")
		return
	}
	writeErrorResponse(w, http.StatusBadRequest, "invalid_json", message)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("request body must contain one JSON value")
		}
		return err
	}
	return nil
}
