package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"newstart/internal/api/client"
	apikafka "newstart/internal/api/kafka"
	"newstart/internal/auth"
	"newstart/internal/calculator"
	"newstart/internal/model"
)

// Repository описывает операции с пользователями в базе данных.
type Repository interface {
	CreateUser(ctx context.Context, username, passwordHash string) (model.User, error)
	GetUserByUsername(ctx context.Context, username string) (model.User, error)
}

// Handler обрабатывает публичные HTTP-запросы API-сервиса.
type Handler struct {
	storage  client.Storage
	producer *apikafka.Producer
	repo     Repository
}

// New создаёт HTTP-обработчик API-сервиса.
func New(storage client.Storage, producer *apikafka.Producer, repo Repository) *Handler {
	return &Handler{
		storage:  storage,
		producer: producer,
		repo:     repo,
	}
}

// Register добавляет маршруты API-сервиса.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle(
		"/calculate",
		auth.AuthMiddleware(http.HandlerFunc(h.calculate)),
	)

	mux.Handle(
		"/calculations",
		auth.AuthMiddleware(http.HandlerFunc(h.listCalculations)),
	)

	mux.HandleFunc("/register", h.register)
	mux.HandleFunc("/login", h.login)
	mux.HandleFunc("/health", h.health)
}

type calculateRequest struct {
	A        float64 `json:"a"`
	B        float64 `json:"b"`
	Operator string  `json:"operator"`
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Message string `json:"message"`
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
	userID, ok := r.Context().Value("user_id").(int)
if !ok {
    writeJSON(w, http.StatusUnauthorized, errorResponse{
        Error: "неавторизованный пользователь",
    })
    return
}

	result, err := calculator.Calculate(req.A, req.B, req.Operator)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	calc := model.Calculation{
		UserID:   userID,
		OperandA: req.A,
		OperandB: req.B,
		Operator: req.Operator,
		Result:   result,
	}

	if err := h.producer.SendCalculation(r.Context(), calc); err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse{Error: "не удалось отправить в kafka"})
		return
	}

	writeJSON(w, http.StatusOK, calc)
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

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.repo.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "неверный логин или пароль"})
		return
	}

	if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "неверный логин или пароль"})
		return
	}

	token, err := auth.GenerateToken(
		user.ID,
		user.Username,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			errorResponse{
				Error: "не удалось создать токен",
			},
		)
		return
	}
	
	writeJSON(w, http.StatusOK, map[string]string{
		"token": token,
	})
}

func (h *Handler) listCalculations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "нужен GET запрос"})
		return
	}
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{
			Error: "не удалось определить пользователя",
		})
		return
	}
	calculations, err := h.storage.ListCalculations(r.Context(), userID)
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
