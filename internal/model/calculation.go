package model

import "time"

// Calculation — одна запись в базе данных (одно вычисление).
type Calculation struct {
	ID        int       `json:"id"`
	OperandA  float64   `json:"a"`
	OperandB  float64   `json:"b"`
	Operator  string    `json:"operator"`
	Result    float64   `json:"result"`
	CreatedAt time.Time `json:"created_at"`
}
