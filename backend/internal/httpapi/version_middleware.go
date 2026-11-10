package httpapi

import (
	"net/http"
	"strings"
)

func apiVersion(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		version := strings.TrimSpace(r.Header.Get(HeaderAPIVersion))
		if version != APIVersionV1 {
			writeErrorResponse(w, http.StatusBadRequest, "unsupported_api_version", "X-API-Version must be v1")
			return
		}
		next.ServeHTTP(w, r)
	})
}
