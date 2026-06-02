package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omniful/payment-platform/services/auth-service/internal/domain"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, email, passwordHash string) (*domain.User, error) {
	id := uuid.New().String()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)`,
		id, email, passwordHash,
	)
	if err != nil {
		return nil, err
	}
	roleID := "00000000-0000-0000-0000-000000000001"
	_, _ = r.pool.Exec(ctx,
		`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`,
		id, roleID,
	)
	return &domain.User{ID: id, Email: email, PasswordHash: passwordHash, Roles: []string{"user"}}, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, created_at FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx,
		`SELECT r.name FROM roles r JOIN user_roles ur ON r.id = ur.role_id WHERE ur.user_id = $1`, u.ID,
	)
	if err != nil {
		return &u, nil
	}
	defer rows.Close()
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err == nil {
			u.Roles = append(u.Roles, role)
		}
	}
	if len(u.Roles) == 0 {
		u.Roles = []string{"user"}
	}
	return &u, nil
}

func (r *UserRepository) SaveRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at) VALUES ($1, $2, $3, $4::timestamp)`,
		uuid.New().String(), userID, tokenHash, expiresAt,
	)
	return err
}

func (r *UserRepository) RevokeRefreshTokens(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1 AND revoked = FALSE`, userID,
	)
	return err
}
