package httpapi

import (
	"encoding/json"
	"errors"
	"io"

	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/aws/smithy-go"
	"github.com/go-chi/chi/v5"
	"github.com/thiagosantos/dynamodb-manager/backend/internal/dynamo"
	"github.com/thiagosantos/dynamodb-manager/backend/internal/usecase"
)

type connectionRequest struct {
	Name     string `json:"name"`
	Region   string `json:"region"`
	Profile  string `json:"profile,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
}

func (r connectionRequest) input() usecase.ConnectionInput {
	return usecase.ConnectionInput{Name: r.Name, Region: r.Region, Profile: r.Profile, Endpoint: r.Endpoint}
}

func listConnections(service *usecase.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connections, err := service.ListConnections(r.Context())
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"connections": connections})
	}
}

func createConnection(service *usecase.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request connectionRequest
		if err := decodeJSON(w, r, &request); err != nil {
			writeDecodeError(w, err, "Request body must be valid connection JSON")
			return
		}
		connection, err := service.CreateConnection(r.Context(), request.input())
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, connection)
	}
}

func testConnection(service *usecase.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request connectionRequest
		if err := decodeJSON(w, r, &request); err != nil {
			writeDecodeError(w, err, "Request body must be valid connection JSON")
			return
		}
		if err := service.TestConnection(r.Context(), request.input()); err != nil {
			var applicationError *usecase.Error
			if errors.As(err, &applicationError) {
				writeError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"success": false, "message": safeTestMessage(err)})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Connection successful"})
	}
}

func deleteConnection(service *usecase.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := service.DeleteConnection(r.Context(), chi.URLParam(r, "connectionID")); err != nil {
			writeError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func listTables(service *usecase.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tables, err := service.ListTables(r.Context(), chi.URLParam(r, "connectionID"))
		if err != nil {
			writeError(w, err)
			return
		}
		result := make([]map[string]string, len(tables))
		for index, name := range tables {
			result[index] = map[string]string{"name": name}
		}
		writeJSON(w, http.StatusOK, map[string]any{"tables": result})
	}
}

func describeTable(service *usecase.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		description, err := service.DescribeTable(r.Context(), chi.URLParam(r, "connectionID"), chi.URLParam(r, "tableName"))
		if err != nil {
			writeError(w, err)
			return
		}
		response := map[string]any{"tableName": chi.URLParam(r, "tableName"), "partitionKey": description.PartitionKey}
		if description.SortKey != nil {
			response["sortKey"] = description.SortKey
		}
		writeJSON(w, http.StatusOK, response)
	}
}

type filterInput struct {
	Attribute string          `json:"attribute"`
	Operator  string          `json:"operator"`
	Value     json.RawMessage `json:"value"`
}

func parseFilter(value string) (*dynamo.Filter, error) {
	if value == "" {
		return nil, nil
	}
	var input filterInput
	decoder := json.NewDecoder(strings.NewReader(value))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, errors.New("invalid filter JSON")
	}
	if input.Attribute == "" || input.Operator != dynamo.FilterOperatorEqual || len(input.Value) == 0 {
		return nil, errors.New("filter requires attribute, operator eq, and value")
	}
	encoded, err := json.Marshal(map[string]json.RawMessage{"value": input.Value})
	if err != nil {
		return nil, err
	}
	var typed dynamo.Item
	if err := json.Unmarshal(encoded, &typed); err != nil {
		return nil, err
	}
	return &dynamo.Filter{Attribute: input.Attribute, Operator: dynamo.FilterOperatorEqual, Value: typed["value"]}, nil
}

func listItems(service *usecase.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "invalid_argument", "Query string is invalid")
			return
		}
		limit := int64(25)
		if raw := query.Get("limit"); raw != "" {
			parsed, err := strconv.ParseInt(raw, 10, 32)
			if err != nil || parsed < 1 || parsed > 100 {
				writeErrorResponse(w, http.StatusBadRequest, "invalid_argument", "Limit must be between 1 and 100")
				return
			}
			limit = parsed
		}
		filter, err := parseFilter(query.Get("filter"))
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "invalid_filter", "Filter must be JSON with attribute, operator eq, and a typed value")
			return
		}
		page, err := service.ListItems(r.Context(), chi.URLParam(r, "connectionID"), chi.URLParam(r, "tableName"), int32(limit), query.Get("nextToken"), filter)
		if err != nil {
			writeError(w, err)
			return
		}
		token, err := usecase.EncodeToken(page.NextKey)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "internal_error", "Could not encode pagination token")
			return
		}
		response := map[string]any{"items": page.Items, "count": page.Count, "scannedCount": page.ScannedCount}
		if token != "" {
			response["nextToken"] = token
		}
		writeJSON(w, http.StatusOK, response)
	}
}

type putItemRequest struct {
	Key  dynamo.Item `json:"key"`
	Item dynamo.Item `json:"item"`
}

type deleteItemRequest struct {
	Key dynamo.Item `json:"key"`
}

func putItem(service *usecase.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request putItemRequest
		if err := decodeJSON(w, r, &request); err != nil {
			writeDecodeError(w, err, "Request body must include typed key and item maps")
			return
		}
		if request.Key == nil || request.Item == nil {
			writeErrorResponse(w, http.StatusBadRequest, "invalid_argument", "Key and item are required")
			return
		}
		if err := service.PutItem(r.Context(), chi.URLParam(r, "connectionID"), chi.URLParam(r, "tableName"), request.Key, request.Item); err != nil {
			writeError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func deleteItem(service *usecase.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request deleteItemRequest
		if err := decodeJSON(w, r, &request); err != nil {
			writeDecodeError(w, err, "Request body must include a typed key map")
			return
		}
		if request.Key == nil {
			writeErrorResponse(w, http.StatusBadRequest, "invalid_argument", "Key is required")
			return
		}
		if err := service.DeleteItem(r.Context(), chi.URLParam(r, "connectionID"), chi.URLParam(r, "tableName"), request.Key); err != nil {
			writeError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func safeTestMessage(err error) string {
	var apiError smithy.APIError
	if errors.As(err, &apiError) {
		switch apiError.ErrorCode() {
		case "AccessDeniedException", "UnrecognizedClientException", "InvalidSignatureException":
			return "AWS denied the request; check the selected profile's credentials and permissions"
		case "ValidationException":
			return "DynamoDB rejected the connection settings; check the region and endpoint"
		case "ResourceNotFoundException":
			return "DynamoDB endpoint was reached, but the requested resource was not found"
		}
	}
	return "Could not connect to DynamoDB; check the region, AWS credentials, endpoint, and network"
}
