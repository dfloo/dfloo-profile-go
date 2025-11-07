package repository

import (
	"context"

	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JobApplicationRepository interface {
	GetJobApplicationsByUserID(ctx context.Context, userID string) ([]*model.JobApplication, error)
	CreateJobApplication(ctx context.Context, jobApplication *model.JobApplication, userID string) error
	UpdateJobApplications(ctx context.Context, jobApplications []*model.JobApplication, userID string) error
}

type DBJobApplicationRepository struct {
	Pool *pgxpool.Pool
}

func NewDBJobApplicationRepository(pool *pgxpool.Pool) *DBJobApplicationRepository {
	return &DBJobApplicationRepository{Pool: pool}
}

func (r *DBJobApplicationRepository) GetJobApplicationsByUserID(
	ctx context.Context,
	userID string,
) ([]*model.JobApplication, error) {
	rows, err := r.Pool.Query(
		ctx,
		`SELECT
            job_application_id,
            status,
            sort_index,
            company,
            role,
            description,
            notes,
			resume_id,
            created,
            updated
         FROM job_application
         WHERE user_id = $1;`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobApplications []*model.JobApplication
	for rows.Next() {
		var jobApplication model.JobApplication
		var resumeID *string

		err = rows.Scan(
			&jobApplication.JobApplicationID,
			&jobApplication.Status,
			&jobApplication.SortIndex,
			&jobApplication.Company,
			&jobApplication.Role,
			&jobApplication.Description,
			&jobApplication.Notes,
			&resumeID,
			&jobApplication.Created,
			&jobApplication.Updated,
		)
		if err != nil {
			return nil, err
		}

		if resumeID != nil {
			jobApplication.ResumeID = *resumeID
		}

		jobApplications = append(jobApplications, &jobApplication)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return jobApplications, nil
}

func (r *DBJobApplicationRepository) CreateJobApplication(
	ctx context.Context,
	jobApplication *model.JobApplication,
	userID string,
) error {
	var resumeID *string
	if jobApplication.ResumeID != "" {
		resumeID = &jobApplication.ResumeID
	}
	err := r.Pool.QueryRow(
		ctx,
		`INSERT INTO job_application (
            user_id,
			status,
			sort_index,
			company,
			role,
			description,
			notes,
			resume_id
         ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		 RETURNING job_application_id, created, updated`,
		userID,
		jobApplication.Status,
		jobApplication.SortIndex,
		jobApplication.Company,
		jobApplication.Role,
		jobApplication.Description,
		jobApplication.Notes,
		resumeID,
	).Scan(
		&jobApplication.JobApplicationID,
		&jobApplication.Created,
		&jobApplication.Updated,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *DBJobApplicationRepository) UpdateJobApplications(
	ctx context.Context,
	jobApplications []*model.JobApplication,
	userID string,
) error {
	if len(jobApplications) == 0 {
		return nil
	}

	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, jobApplication := range jobApplications {
		var resumeID *string
		if jobApplication.ResumeID != "" {
			resumeID = &jobApplication.ResumeID
		}
		err := tx.QueryRow(
			ctx,
			`UPDATE job_application SET
                status = $1,
                sort_index = $2,
                company = $3,
                role = $4,
                description = $5,
                notes = $6,
                resume_id = $7
            WHERE job_application_id = $8 AND user_id = $9 
            RETURNING created, updated`,
			jobApplication.Status,
			jobApplication.SortIndex,
			jobApplication.Company,
			jobApplication.Role,
			jobApplication.Description,
			jobApplication.Notes,
			resumeID,
			jobApplication.JobApplicationID,
			userID,
		).Scan(&jobApplication.Created, &jobApplication.Updated)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
