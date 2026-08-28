package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/supercharged09/crypto-rates/internal/model"
	"github.com/supercharged09/crypto-rates/internal/service"
)

// RateHandler орабатывает HTTP запросы, связанные с курсами
// Принимает зависимость на сервис через конструктор (Dependency injection)
type RateHandler struct {
	rateService      service.RateServiceInterface
	analyticsService *service.AnalyticsService
	chartService     *service.ChartService
}

// NewRateHandler - конструктор хэндлера
func NewRateHandler(
	rateService service.RateServiceInterface, // ← интерфейс
	analyticsService *service.AnalyticsService,
	chartService *service.ChartService,
) *RateHandler {
	return &RateHandler{
		rateService:      rateService,
		analyticsService: analyticsService,
		chartService:     chartService,
	}
}

// GetAnalytics возвращает статистику пользователя
func (h *RateHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.analyticsService.GetStats(r.Context())
	if err != nil {
		log.Printf("ERROR: falied to get analytics: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get analytics",
		})
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// GetAllRates - обработчик GET /rates
func (h *RateHandler) GetAllRates(w http.ResponseWriter, r *http.Request) {
	var rates []interface{}

	for _, crypto := range model.SupportedCryptos {
		stats, err := h.rateService.GetRateStats(r.Context(), crypto.ID)
		if err != nil {
			log.Printf("WARNING: failed to get %s stats: %v", crypto.ID, err)
			continue
		}
		rates = append(rates, stats)
	}

	if len(rates) == 0 {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get rates",
		})
		return
	}
	//Формирование ответа
	response := map[string]interface{}{
		"rates": rates,
	}

	writeJSON(w, http.StatusOK, response)

}

// GetRate - обработчик GET /rates/{cryptocurrency}
func (h *RateHandler) GetRate(w http.ResponseWriter, r *http.Request) {
	//chi.URLParam извлекает параметр из URL: /rates/{cryptocurrency}
	crypto := chi.URLParam(r, "cryptocurrency")

	//ищем валюту по ID или символу
	info := model.GetCryptoInfo(crypto)
	if info == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "unsupported cryptocurrency: " + crypto,
		})
		return
	}

	stats, err := h.rateService.GetRateStats(r.Context(), info.ID)
	if err != nil {
		log.Printf("ERROR: failed to get %s stats: %v", info.ID, err)
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "rate not found for " + info.ID,
		})
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// writeJSON - вспомогательная функция для отправки JSON ответа
// Приватная, используется только внутри пакета handler
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	//Установка заголовка Content-Type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	//Кодирование data в JSON и запись в ResponseWriter
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("ERROR: failed to encode response: %v", err)
	}
}

// GetChart возвращает html с графиком
func (h *RateHandler) GetChart(w http.ResponseWriter, r *http.Request) {
	crypto := chi.URLParam(r, "cryptocurrency")

	info := model.GetCryptoInfo(crypto)
	if info == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "unsupported cryptocurrency: " + crypto,
		})
		return
	}

	html, err := h.chartService.GenerateChartHTML(r.Context(), info.ID)
	if err != nil {
		log.Printf("ERROR: failed to generate chart: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to generate chart",
		})
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=uft-8")
	w.Write([]byte(html))
}
