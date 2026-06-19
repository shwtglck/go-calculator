package repository

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"newstart/internal/model"
)

// Postgres хранит и читает вычисления в PostgreSQL.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres подключается к базе по строке DATABASE_URL.
func NewPostgres(ctx context.Context) (*Postgres, error) {
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

	return &Postgres{pool: pool}, nil
}

// Close закрывает пул соединений с базой данных.
func (p *Postgres) Close() {
	p.pool.Close()
}

// SaveCalculation сохраняет результат вычисления в таблицу calculations.
func (p *Postgres) SaveCalculation(ctx context.Context, a, b float64, operator string, result float64) (model.Calculation, error) {
	const query = `
		INSERT INTO calculations (operand_a, operand_b, operator, result)
		VALUES ($1, $2, $3, $4)
		RETURNING id, operand_a, operand_b, operator, result, created_at
	`

	var calc model.Calculation
	err := p.pool.QueryRow(ctx, query, a, b, operator, result).Scan(
		&calc.ID,
		&calc.OperandA,
		&calc.OperandB,
		&calc.Operator,
		&calc.Result,
		&calc.CreatedAt,
	)
	if err != nil {
		return model.Calculation{}, fmt.Errorf("сохранение вычисления: %w", err)
	}

	return calc, nil
}

// ListCalculations возвращает все записи из базы, от новых к старым.
func (p *Postgres) ListCalculations(ctx context.Context) ([]model.Calculation, error) {
	const query = `
		SELECT id, operand_a, operand_b, operator, result, created_at
		FROM calculations
		ORDER BY created_at DESC
	`

	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("чтение вычислений: %w", err)
	}
	defer rows.Close()

	calculations := make([]model.Calculation, 0)
	for rows.Next() {
		var calc model.Calculation
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
