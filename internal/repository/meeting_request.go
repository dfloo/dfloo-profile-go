package repository

import (
	"context"

	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MeetingRequestRepository interface {
	CreateMeetingRequest(ctx context.Context, req *model.MeetingRequest) error
}

type DBMeetingRequestRepository struct {
	Pool *pgxpool.Pool
}

func NewDBMeetingRequestRepository(pool *pgxpool.Pool) *DBMeetingRequestRepository {
	return &DBMeetingRequestRepository{Pool: pool}
}

func (r *DBMeetingRequestRepository) CreateMeetingRequest(ctx context.Context, req *model.MeetingRequest) error {
	return r.Pool.QueryRow(
		ctx,
		`INSERT INTO meeting_request (name, email, message) VALUES ($1, $2, $3) RETURNING created`,
		req.Name,
		req.Email,
		req.Message,
	).Scan(&req.Created)
}
