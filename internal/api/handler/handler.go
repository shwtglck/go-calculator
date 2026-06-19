package handler

import (
	"encoding/json"
	"net/http"

	"newstart/internal/api/client"
	"newstart/internal/calculator"
	"newstart/internal/model"
)

// Handler обрабатывает публичные HTTP-запросы API-сервиса.
type Handler struct {
	storage client.Storage
}

// New создаёт HTTP-обработчик API-сервиса.
func New(storage client.Storage) *Handler {
	return &Handler{storage: storage}
}

// Register добавляет маршруты API-сервиса.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/calculate", h.calculate)
	mux.HandleFunc("/calculations", h.listCalculations)
	mux.HandleFunc("/health", h.health)
}

type calculateRequest struct {
	A        float64 `json:"a"`
	B        float64 `json:"b"`
	Operator string  `json:"operator"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) calculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "нужен POST запрос"})
		return
	}

	var req calculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "неверный JSON, пример: {\"a\": 10, \"b\": 5, \"operator\": \"+\"}"})
		return
	}

	result, err := calculator.Calculate(req.A, req.B, req.Operator)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	calc, err := h.storage.SaveCalculation(r.Context(), req.A, req.B, req.Operator, result)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse{Error: "storage-сервис недоступен"})
		return
	}

	writeJSON(w, http.StatusOK, calc)
}

func (h *Handler) listCalculations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "нужен GET запрос"})
		return
	}

	calculations, err := h.storage.ListCalculations(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse{Error: "storage-сервис недоступен"})
		return
	}

	if calculations == nil {
		calculations = []model.Calculation{}
	}

	writeJSON(w, http.StatusOK, calculations)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "api"})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
