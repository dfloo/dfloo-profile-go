package repository

import (
	"context"

	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SignupRequestRepository interface {
	CreateSignupRequest(ctx context.Context, req *model.SignupRequest) error
}

type DBSignupRequestRepository struct {
	Pool *pgxpool.Pool
}

func NewDBSignupRequestRepository(pool *pgxpool.Pool) *DBSignupRequestRepository {
	return &DBSignupRequestRepository{Pool: pool}
}

func (r *DBSignupRequestRepository) CreateSignupRequest(ctx context.Context, req *model.SignupRequest) error {
	return r.Pool.QueryRow(
		ctx,
		`INSERT INTO signup_request (name, email, reason) VALUES ($1, $2, $3) RETURNING created`,
		req.Name,
		req.Email,
		req.Reason,
	).Scan(&req.Created)
}
