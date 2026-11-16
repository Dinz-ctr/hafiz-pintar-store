package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAdminNotFound = errors.New("admin not found")

type Admin struct {
	ID           int64  `json:"id"`
	FullName     string `json:"full_name"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	IsActive     bool   `json:"is_active"`
}

type AdminRepository struct {
	pool *pgxpool.Pool
}

func NewAdminRepository(pool *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{pool: pool}
}

func (r *AdminRepository) GetByEmail(
	ctx context.Context,
	email string,
) (Admin, error) {
	var admin Admin

	err := r.pool.QueryRow(ctx, `
		SELECT id, full_name, email, password_hash, is_active
		FROM admins
		WHERE email = $1
	`, email).Scan(
		&admin.ID,
		&admin.FullName,
		&admin.Email,
		&admin.PasswordHash,
		&admin.IsActive,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Admin{}, ErrAdminNotFound
	}
	if err != nil {
		return Admin{}, err
	}

	return admin, nil
}

func (r *AdminRepository) GetByID(
	ctx context.Context,
	id int64,
) (Admin, error) {
	var admin Admin

	err := r.pool.QueryRow(ctx, `
		SELECT id, full_name, email, is_active
		FROM admins
		WHERE id = $1
	`, id).Scan(
		&admin.ID,
		&admin.FullName,
		&admin.Email,
		&admin.IsActive,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Admin{}, ErrAdminNotFound
	}
	if err != nil {
		return Admin{}, err
	}

	return admin, nil
}

func (r *AdminRepository) Create(
	ctx context.Context,
	fullName string,
	email string,
	passwordHash string,
) (Admin, error) {
	var admin Admin

	err := r.pool.QueryRow(ctx, `
		INSERT INTO admins (
			full_name,
			email,
			password_hash
		)
		VALUES ($1, $2, $3)
		RETURNING id, full_name, email, is_active
	`,
		fullName,
		email,
		passwordHash,
	).Scan(
		&admin.ID,
		&admin.FullName,
		&admin.Email,
		&admin.IsActive,
	)

	if err != nil {
		return Admin{}, err
	}

	return admin, nil
}
