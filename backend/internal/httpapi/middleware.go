package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (w *responseRecorder) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseRecorder) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			recorder := &responseRecorder{ResponseWriter: w}
			next.ServeHTTP(recorder, r)
			status := recorder.status
			if status == 0 {
				status = http.StatusOK
			}
			logger.Info("http request", "method", r.Method, "path", r.URL.Path, "status", status, "duration_ms", time.Since(started).Milliseconds(), "request_id", middleware.GetReqID(r.Context()))
		})
	}
}

func recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recover() != nil {
					logger.Error("http request panicked", "request_id", middleware.GetReqID(r.Context()))
					writeErrorResponse(w, http.StatusInternalServerError, "internal_error", "Internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
