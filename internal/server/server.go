package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/patrickneise/dashboard/internal/ui/components"
	"github.com/patrickneise/dashboard/internal/ui/pages"
	"github.com/patrickneise/dashboard/internal/widgetkit"
)

func RegisterRoutes(r chi.Router, reg *widgetkit.Registry) {
	if reg == nil {
		panic("server.RegisterRoutes: registry is nil")
	}

	// Static assets
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Build dashboard cards once (registry is startup-time config)
	specs := reg.List()
	cards := make([]components.WidgetCardProps, 0, len(specs))
	for _, s := range specs {
		cards = append(cards, components.WidgetCardProps{
			Title:    s.Title,
			Endpoint: "/widgets/" + s.Key,
			Trigger:  s.Trigger,
			Class:    s.Class,
		})
	}

	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		component := pages.DashboardPage(cards)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := component.Render(req.Context(), w); err != nil {
			http.Error(w, "render error", http.StatusInternalServerError)
		}
	})

	// Widgets auto-mounted under /widgets/<key>
	r.Route("/widgets", func(wr chi.Router) {
		for _, s := range specs {
			wr.Handle("/"+s.Key, s.Handler)
		}
	})
}
