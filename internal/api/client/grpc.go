package client

import (
	"context"
	"fmt"

	"newstart/internal/model"
)

// GRPCStorageClient будет обращаться к storage-сервису по gRPC.
// Пока не реализован — заготовка под proto/storage/v1/storage.proto.
type GRPCStorageClient struct{}

// NewGRPCStorageClient создаёт gRPC-клиент storage-сервиса.
func NewGRPCStorageClient(_ string) (Storage, error) {
	return nil, fmt.Errorf("gRPC transport is not implemented yet")
}

func (c *GRPCStorageClient) SaveCalculation(_ context.Context, _ float64, _ float64, _ string, _ float64) (model.Calculation, error) {
	return model.Calculation{}, fmt.Errorf("gRPC transport is not implemented yet")
}

func (c *GRPCStorageClient) ListCalculations(_ context.Context) ([]model.Calculation, error) {
	return nil, fmt.Errorf("gRPC transport is not implemented yet")
}
