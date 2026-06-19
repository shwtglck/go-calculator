package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"newstart/internal/model"
)

// HTTPStorageClient обращается к storage-сервису по HTTP.
type HTTPStorageClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPStorageClient создаёт HTTP-клиент для storage-сервиса.
func NewHTTPStorageClient(baseURL string) *HTTPStorageClient {
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}

	return &HTTPStorageClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
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

// SaveCalculation отправляет запрос на сохранение вычисления в storage-сервис.
func (c *HTTPStorageClient) SaveCalculation(ctx context.Context, a, b float64, operator string, result float64) (model.Calculation, error) {
	body, err := json.Marshal(saveRequest{
		A:        a,
		B:        b,
		Operator: operator,
		Result:   result,
	})
	if err != nil {
		return model.Calculation{}, fmt.Errorf("кодирование запроса: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/calculations", bytes.NewReader(body))
	if err != nil {
		return model.Calculation{}, fmt.Errorf("создание запроса: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.Calculation{}, fmt.Errorf("запрос к storage-сервису: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return model.Calculation{}, readServiceError(resp)
	}

	var calc model.Calculation
	if err := json.NewDecoder(resp.Body).Decode(&calc); err != nil {
		return model.Calculation{}, fmt.Errorf("декодирование ответа: %w", err)
	}

	return calc, nil
}

// ListCalculations запрашивает историю вычислений из storage-сервиса.
func (c *HTTPStorageClient) ListCalculations(ctx context.Context) ([]model.Calculation, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/calculations", nil)
	if err != nil {
		return nil, fmt.Errorf("создание запроса: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("запрос к storage-сервису: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, readServiceError(resp)
	}

	var calculations []model.Calculation
	if err := json.NewDecoder(resp.Body).Decode(&calculations); err != nil {
		return nil, fmt.Errorf("декодирование ответа: %w", err)
	}

	return calculations, nil
}

func readServiceError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var errResp errorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
		return fmt.Errorf("storage-сервис вернул %d: %s", resp.StatusCode, errResp.Error)
	}

	return fmt.Errorf("storage-сервис вернул %d", resp.StatusCode)
}

// Config задаёт способ подключения API-сервиса к storage-сервису.
type Config struct {
	Transport  string
	StorageURL string
}

// New создаёт клиент storage-сервиса по выбранному транспорту.
func New(cfg Config) (Storage, error) {
	switch cfg.Transport {
	case "", "http":
		url := cfg.StorageURL
		if url == "" {
			url = os.Getenv("STORAGE_URL")
		}
		return NewHTTPStorageClient(url), nil
	case "grpc":
		return NewGRPCStorageClient(cfg.StorageURL)
	default:
		return nil, fmt.Errorf("неизвестный транспорт %q, используйте http или grpc", cfg.Transport)
	}
}
