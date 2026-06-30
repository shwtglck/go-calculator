package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"newstart/internal/auth"
	"newstart/internal/model"
	"strconv"
)

// Repository описывает операции хранения данных.
// Сейчас реализовано через PostgreSQL, позже можно добавить кэш или другие источники.
type Repository interface {
	SaveCalculation(ctx context.Context, userID int, a, b float64, operator string, result float64) (model.Calculation, error)
	ListCalculations(ctx context.Context, userID int) ([]model.Calculation, error)
	CreateUser(ctx context.Context, username, passwordHash string) (model.User, error)
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
	mux.HandleFunc("/register", h.register)
	mux.HandleFunc("/health", h.health)
}

type saveRequest struct {
	A        float64 `json:"a"`
	B        float64 `json:"b"`
	Operator string  `json:"operator"`
	Result   float64 `json:"result"`
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
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

	calc, err := h.repo.SaveCalculation(r.Context(), 1, req.A, req.B, req.Operator, req.Result)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "не удалось сохранить запись"})
		return
	}

	writeJSON(w, http.StatusCreated, calc)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "нужен POST запрос"})
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "неверный JSON"})
		return
	}

	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "username и password обязательны"})
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "не удалось захешировать пароль"})
		return
	}

	user, err := h.repo.CreateUser(r.Context(), req.Username, passwordHash)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "не удалось создать пользователя"})
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *Handler) listCalculations(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")

userID, err := strconv.Atoi(userIDStr)
if err != nil {
	writeJSON(w, http.StatusBadRequest, errorResponse{
		Error: "неверный X-User-ID",
	})
	return
}
	calculations, err := h.repo.ListCalculations(r.Context(), userID)
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
