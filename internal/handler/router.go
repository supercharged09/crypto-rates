package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter создает и настраивает HTTP роутер
func NewRouter(rateHandler *RateHandler) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Healthcheck
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	})
	r.Get("/analytics", rateHandler.GetAnalytics)

	// Маршруты для курсов
	r.Route("/rates", func(r chi.Router) {
		r.Get("/", rateHandler.GetAllRates)
		r.Get("/{cryptocurrency}", rateHandler.GetRate)
	})

	//график цены (HTML)
	r.Get("/chart/{cryptocurrency}", rateHandler.GetChart)

	return r
}
