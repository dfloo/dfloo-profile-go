package repository

import (
	"context"

	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InvitationRequestRepository interface {
	CreateInvitationRequest(ctx context.Context, req *model.InvitationRequest) error
}

type DBInvitationRequestRepository struct {
	Pool *pgxpool.Pool
}

func NewDBInvitationRequestRepository(pool *pgxpool.Pool) *DBInvitationRequestRepository {
	return &DBInvitationRequestRepository{Pool: pool}
}

func (r *DBInvitationRequestRepository) CreateInvitationRequest(ctx context.Context, req *model.InvitationRequest) error {
	return r.Pool.QueryRow(
		ctx,
		`INSERT INTO invitation_request (name, email, reason) VALUES ($1, $2, $3) RETURNING created`,
		req.Name,
		req.Email,
		req.Reason,
	).Scan(&req.Created)
}
