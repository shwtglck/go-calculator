package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"newstart/internal/model"
)

// Repository описывает операции хранения данных.
// Сейчас реализовано через PostgreSQL, позже можно добавить кэш или другие источники.
type Repository interface {
	SaveCalculation(ctx context.Context, a, b float64, operator string, result float64) (model.Calculation, error)
	ListCalculations(ctx context.Context) ([]model.Calculation, error)
}

// Handler обрабатывает HTTP-запросы storage-сервиса.
type Handler struct {
	repo Repository
}

// New создаёт HTTP-обработчик storage-сервиса.
func New(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// Register добавляет маршруты storage-сервиса.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/calculations", h.calculations)
	mux.HandleFunc("/health", h.health)
}

type saveRequest struct {
	A        float64 `json:"a"`
	B        float64 `json:"b"`
	Operator string  `json:"operator"`
	Result   float64 `json:"result"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) calculations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.saveCalculation(w, r)
	case http.MethodGet:
		h.listCalculations(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "нужен GET или POST запрос"})
	}
}

func (h *Handler) saveCalculation(w http.ResponseWriter, r *http.Request) {
	var req saveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "неверный JSON"})
		return
	}

	calc, err := h.repo.SaveCalculation(r.Context(), req.A, req.B, req.Operator, req.Result)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "не удалось сохранить запись"})
		return
	}

	writeJSON(w, http.StatusCreated, calc)
}

func (h *Handler) listCalculations(w http.ResponseWriter, r *http.Request) {
	calculations, err := h.repo.ListCalculations(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "не удалось прочитать записи"})
		return
	}

	if calculations == nil {
		calculations = []model.Calculation{}
	}

	writeJSON(w, http.StatusOK, calculations)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "storage"})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
