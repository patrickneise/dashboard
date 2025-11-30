package widgetkit

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/patrickneise/dashboard/internal/cache"
)

type Handler[T any] struct {
	Name  string
	TTL   time.Duration
	Cache *cache.TTL[T]

	Fetch  func(ctx context.Context) (T, error)
	Render func(data T) templ.Component
	Error  func(err error) templ.Component

	MarkStale func(v T, staleBy time.Duration) T

	Log *slog.Logger
}

func (h Handler[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	now := time.Now()
	reqID := middleware.GetReqID(r.Context())
	hx := r.Header.Get("HX-Request") == "true"
	path := r.URL.Path

	log := h.reqLogger(reqID, hx, path)

	var (
		cached     T
		cacheExp   time.Time
		cacheState cache.State = cache.Miss
	)

	// Cached version is current
	if h.Cache != nil {
		cached, cacheExp, cacheState = h.Cache.Get(now)
		if cacheState == cache.Fresh {
			if err := h.Render(cached).Render(r.Context(), w); err != nil {
				h.renderError(w, r, err)
			}
			return
		}
	}

	// Cache is stale or missing: attempt refresh
	v, err := h.Fetch(r.Context())
	if err != nil {
		// If we have stale data, serve it instead of erroring the widget.
		if cacheState == cache.Stale {
			staleBy := now.Sub(cacheExp)
			if log != nil {
				log.Warn("widget_fetch_failed_serving_stale",
					slog.Duration("stale_by", staleBy),
					slog.Any("err", err))
			}
			w.Header().Set("X-Widget-Stale", "true")
			toRender := cached
			if h.MarkStale != nil {
				toRender = h.MarkStale(cached, staleBy)
			}
			if err := h.Render(toRender).Render(r.Context(), w); err != nil {
				h.renderError(w, r, err)
			}
			return
		}

		// No cache to fall back to
		if log != nil {
			log.Error("widget_fetch_failed", slog.Any("err", err))
		}

		w.WriteHeader(http.StatusBadGateway)
		if h.Error != nil {
			_ = h.Error(err).Render(r.Context(), w)
		}
		return
	}

	// Refresh succeeded: update cache + render
	if h.Cache != nil && h.TTL > 0 {
		h.Cache.Set(v, now.Add(h.TTL))
	}

	if err := h.Render(v).Render(r.Context(), w); err != nil {
		h.renderError(w, r, err)
	}
}

func (h Handler[T]) renderError(w http.ResponseWriter, r *http.Request, err error) {
	reqID := middleware.GetReqID(r.Context())
	hx := r.Header.Get("HX-Request") == "true"
	path := r.URL.Path

	log := h.reqLogger(reqID, hx, path)
	if log != nil {
		log.Error("widget_render_failed", slog.Any("err", err))
	}
	http.Error(w, "render error", http.StatusInternalServerError)
}

func (h Handler[T]) reqLogger(reqID string, hx bool, path string) *slog.Logger {
	if h.Log == nil {
		return nil
	}
	return h.Log.With(
		slog.String("widget", h.Name),
		slog.String("req_id", reqID),
		slog.Bool("hx", hx),
		slog.String("path", path),
	)
}
