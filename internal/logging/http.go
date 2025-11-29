package logging

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusRecorder) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusRecorder) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func RequestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := &statusRecorder{ResponseWriter: w}
			next.ServeHTTP(ww, r)

			if ww.status == 0 {
				ww.status = http.StatusOK
			}

			routePattern := ""
			if rc := chi.RouteContext(r.Context()); rc != nil {
				routePattern = rc.RoutePattern()
			}

			reqID := middleware.GetReqID(r.Context())
			hx := r.Header.Get("HX-Request") == "true"

			log.Info("http_request",
				slog.String("req_id", reqID),
				slog.Bool("hx", hx),

				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("route", routePattern),

				slog.Int("status", ww.status),
				slog.Int("bytes", ww.bytes),
				slog.Duration("duration", time.Since(start)),

				slog.String("remote", r.RemoteAddr),
			)
		})
	}
}
