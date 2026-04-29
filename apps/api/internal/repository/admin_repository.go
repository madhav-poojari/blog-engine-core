package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/typewriter-blog/api/internal/model"
)

type AdminRepository struct {
	db *sqlx.DB
}

func NewAdminRepository(db *sqlx.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) GetByUsername(ctx context.Context, username string) (*model.Admin, error) {
	var admin model.Admin
	err := r.db.GetContext(ctx, &admin, `SELECT * FROM admins WHERE username = $1`, username)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get admin by username: %w", err)
	}
	return &admin, nil
}

func (r *AdminRepository) Create(ctx context.Context, admin *model.Admin) error {
	query := `INSERT INTO admins (id, username, password_hash, created_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, admin.ID, admin.Username, admin.PasswordHash, admin.CreatedAt)
	if err != nil {
		return fmt.Errorf("create admin: %w", err)
	}
	return nil
}
