package main

import (
	"encoding/json"
	"net/http"
)

// calculateRequest — то, что клиент отправляет в POST /calculate.
type calculateRequest struct {
	A        float64 `json:"a"`
	B        float64 `json:"b"`
	Operator string  `json:"operator"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type handler struct {
	storage *Storage
}

func (h *handler) calculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "нужен POST запрос"})
		return
	}

	var req calculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "неверный JSON, пример: {\"a\": 10, \"b\": 5, \"operator\": \"+\"}"})
		return
	}

	result, err := Calculate(req.A, req.B, req.Operator)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	calc, err := h.storage.SaveCalculation(r.Context(), req.A, req.B, req.Operator, result)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "не удалось сохранить в базу"})
		return
	}

	writeJSON(w, http.StatusOK, calc)
}

func (h *handler) listCalculations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "нужен GET запрос"})
		return
	}

	calculations, err := h.storage.ListCalculations(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "не удалось прочитать из базы"})
		return
	}

	if calculations == nil {
		calculations = []Calculation{}
	}

	writeJSON(w, http.StatusOK, calculations)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
