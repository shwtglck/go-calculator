package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Calculation — одна запись в базе данных (одно вычисление).
type Calculation struct {
	ID        int       `json:"id"`
	OperandA  float64   `json:"a"`
	OperandB  float64   `json:"b"`
	Operator  string    `json:"operator"`
	Result    float64   `json:"result"`
	CreatedAt time.Time `json:"created_at"`
}

// Storage отвечает за работу с PostgreSQL.
type Storage struct {
	pool *pgxpool.Pool
}

// NewStorage подключается к базе по строке DATABASE_URL из переменных окружения.
func NewStorage(ctx context.Context) (*Storage, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://calculator:calculator@localhost:5432/calculator?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("подключение к базе: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("проверка соединения с базой: %w", err)
	}

	return &Storage{pool: pool}, nil
}

func (s *Storage) Close() {
	s.pool.Close()
}

// SaveCalculation сохраняет результат вычисления в таблицу calculations.
func (s *Storage) SaveCalculation(ctx context.Context, a, b float64, operator string, result float64) (Calculation, error) {
	const query = `
		INSERT INTO calculations (operand_a, operand_b, operator, result)
		VALUES ($1, $2, $3, $4)
		RETURNING id, operand_a, operand_b, operator, result, created_at
	`

	var calc Calculation
	err := s.pool.QueryRow(ctx, query, a, b, operator, result).Scan(
		&calc.ID,
		&calc.OperandA,
		&calc.OperandB,
		&calc.Operator,
		&calc.Result,
		&calc.CreatedAt,
	)
	if err != nil {
		return Calculation{}, fmt.Errorf("сохранение вычисления: %w", err)
	}

	return calc, nil
}

// ListCalculations возвращает все записи из базы, от новых к старым.
func (s *Storage) ListCalculations(ctx context.Context) ([]Calculation, error) {
	const query = `
		SELECT id, operand_a, operand_b, operator, result, created_at
		FROM calculations
		ORDER BY created_at DESC
	`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("чтение вычислений: %w", err)
	}
	defer rows.Close()

	calculations := make([]Calculation, 0)
	for rows.Next() {
		var calc Calculation
		if err := rows.Scan(
			&calc.ID,
			&calc.OperandA,
			&calc.OperandB,
			&calc.Operator,
			&calc.Result,
			&calc.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("чтение строки: %w", err)
		}
		calculations = append(calculations, calc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("обход строк: %w", err)
	}

	return calculations, nil
}
