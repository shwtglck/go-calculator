package client

import (
	"context"

	"newstart/internal/model"
)

// Storage описывает контракт обращения API-сервиса к storage-сервису.
// Сейчас используется HTTP-клиент, позже здесь появится gRPC-реализация.
type Storage interface {
	SaveCalculation(ctx context.Context, a, b float64, operator string, result float64) (model.Calculation, error)
	ListCalculations(ctx context.Context) ([]model.Calculation, error)
}
