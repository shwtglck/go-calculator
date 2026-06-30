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
func (p *Postgres) SaveCalculation(
	ctx context.Context,
	userID int,
	a, b float64,
	operator string,
	result float64,
) (model.Calculation, error) {
	const query = `
		INSERT INTO calculations (user_id, operand_a, operand_b, operator, result)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, operand_a, operand_b, operator, result, created_at
	`

	var calc model.Calculation
	err := p.pool.QueryRow(ctx, query, userID, a, b, operator, result).Scan(
		&calc.ID,
		&calc.UserID,
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
func (p *Postgres) ListCalculations(ctx context.Context, userID int) ([]model.Calculation, error) {
	const query = `
		SELECT id, user_id, operand_a, operand_b, operator, result, created_at
		FROM calculations
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := p.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("чтение вычислений: %w", err)
	}
	defer rows.Close()

	calculations := make([]model.Calculation, 0)
	for rows.Next() {
		var calc model.Calculation
		if err := rows.Scan(
			&calc.ID,
			&calc.UserID,
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
	func (p *Postgres) CreateUser(ctx context.Context, username, passwordHash string) (model.User, error) {
		const query = `
			INSERT INTO users (username, password_hash)
			VALUES ($1, $2)
			RETURNING id, username, password_hash, created_at
		`
	
		var user model.User
	
		err := p.pool.QueryRow(ctx, query, username, passwordHash).Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
			&user.CreatedAt,
		)
	
		if err != nil {
			return model.User{}, fmt.Errorf("create user: %w", err)
		}
	
		return user, nil
	}
	
	// GetUserByUsername ищет пользователя по логину.
	func (p *Postgres) GetUserByUsername(ctx context.Context, username string) (model.User, error) {
		const query = `
			SELECT id, username, password_hash, created_at
			FROM users
			WHERE username = $1
		`
	
		var user model.User
	
		err := p.pool.QueryRow(ctx, query, username).Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
			&user.CreatedAt,
		)
	
		if err != nil {
			return model.User{}, fmt.Errorf("get user: %w", err)
		}
	
		return user, nil
	}


