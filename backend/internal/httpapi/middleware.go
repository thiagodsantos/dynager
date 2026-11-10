package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
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

func requestLogger(logger *zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			recorder := &responseRecorder{ResponseWriter: w}
			next.ServeHTTP(recorder, r)
			status := recorder.status
			if status == 0 {
				status = http.StatusOK
			}
			logEvent := logger.Info()
			if status >= http.StatusInternalServerError {
				logEvent = logger.Error()
			} else if status >= http.StatusBadRequest {
				logEvent = logger.Warn()
			}
			logEvent.Str("method", r.Method).Str("path", r.URL.Path).Int("status", status).Int64("duration_ms", time.Since(started).Milliseconds()).Str("request_id", middleware.GetReqID(r.Context())).Msg("http request")
		})
	}
}

func recoverer(logger *zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recover() != nil {
					logger.Error().Str("request_id", middleware.GetReqID(r.Context())).Msg("http request panicked")
					writeErrorResponse(w, http.StatusInternalServerError, "internal_error", "Internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
