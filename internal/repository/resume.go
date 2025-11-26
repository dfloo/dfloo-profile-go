package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ResumeRepository interface {
	GetResumesByUserID(ctx context.Context, userID string) ([]*model.Resume, error)
	GetResumeByID(ctx context.Context, resumeID, userID string) (*model.Resume, error)
	GetDefaultResume(ctx context.Context) (*model.Resume, error)
	CreateResume(ctx context.Context, resume *model.Resume, userID string) error
	UpdateResume(ctx context.Context, resume *model.Resume, userID string) error
	DeleteResumes(ctx context.Context, resumeIDs []string, userID string) ([]string, error)
}

type DBResumeRepository struct {
	Pool *pgxpool.Pool
}

func NewDBResumeRepository(pool *pgxpool.Pool) *DBResumeRepository {
	return &DBResumeRepository{Pool: pool}
}

func (r *DBResumeRepository) GetResumeByID(
	ctx context.Context,
	resumeID string,
	userID string,
) (*model.Resume, error) {
	var resume model.Resume
	var profile model.Profile
	row := r.Pool.QueryRow(
		ctx,
		`SELECT
            resume.resume_id,
            resume.sections,
            resume.summary,
            resume.skills,
            resume.experience,
            resume.education,
            resume.file_name,
            resume.template_settings,
			resume.description,
			resume.defaultResume,
			resume.created,
			resume.updated,
            profile.profile_id,
            profile.phone_number,
            profile.email,
            profile.first_name,
            profile.middle_name,
            profile.last_name,
            profile.address_1,
            profile.address_2,
            profile.city,
            profile.state,
            profile.zip_code,
            profile.country,
            profile.social_accounts,
			profile.created,
			profile.updated
         FROM resume
         LEFT JOIN profile ON resume.profile_id = profile.profile_id
         WHERE resume.resume_id = $1 AND resume.user_id = $2;`,
		resumeID,
		userID,
	)
	err := row.Scan(
		&resume.ResumeID,
		&resume.Sections,
		&resume.Summary,
		&resume.Skills,
		&resume.Experience,
		&resume.Education,
		&resume.FileName,
		&resume.TemplateSettings,
		&resume.Description,
		&resume.Default,
		&resume.Created,
		&resume.Updated,
		&profile.ProfileID,
		&profile.PhoneNumber,
		&profile.Email,
		&profile.FirstName,
		&profile.MiddleName,
		&profile.LastName,
		&profile.Address1,
		&profile.Address2,
		&profile.City,
		&profile.State,
		&profile.ZipCode,
		&profile.Country,
		&profile.SocialAccounts,
		&profile.Created,
		&profile.Updated,
	)
	if err != nil {
		return nil, err
	}
	resume.Profile = profile

	return &resume, nil
}

func (r *DBResumeRepository) GetDefaultResume(
	ctx context.Context,
) (*model.Resume, error) {
	var resume model.Resume
	var profile model.Profile
	row := r.Pool.QueryRow(
		ctx,
		`SELECT
            resume.resume_id,
            resume.sections,
            resume.summary,
            resume.skills,
            resume.experience,
            resume.education,
            resume.file_name,
            resume.template_settings,
			resume.description,
			resume.defaultResume,
			resume.created,
			resume.updated,
            profile.profile_id,
            profile.phone_number,
            profile.email,
            profile.first_name,
            profile.middle_name,
            profile.last_name,
            profile.address_1,
            profile.address_2,
            profile.city,
            profile.state,
            profile.zip_code,
            profile.country,
            profile.social_accounts,
			profile.created,
			profile.updated
         FROM resume
         LEFT JOIN profile ON resume.profile_id = profile.profile_id
         WHERE resume.defaultResume = TRUE;`,
	)
	err := row.Scan(
		&resume.ResumeID,
		&resume.Sections,
		&resume.Summary,
		&resume.Skills,
		&resume.Experience,
		&resume.Education,
		&resume.FileName,
		&resume.TemplateSettings,
		&resume.Description,
		&resume.Default,
		&resume.Created,
		&resume.Updated,
		&profile.ProfileID,
		&profile.PhoneNumber,
		&profile.Email,
		&profile.FirstName,
		&profile.MiddleName,
		&profile.LastName,
		&profile.Address1,
		&profile.Address2,
		&profile.City,
		&profile.State,
		&profile.ZipCode,
		&profile.Country,
		&profile.SocialAccounts,
		&profile.Created,
		&profile.Updated,
	)
	if err != nil {
		return nil, err
	}
	resume.Profile = profile

	return &resume, nil
}

func (r *DBResumeRepository) GetResumesByUserID(
	ctx context.Context,
	userID string,
) ([]*model.Resume, error) {
	rows, err := r.Pool.Query(
		ctx,
		`SELECT
            resume.resume_id,
            resume.sections,
            resume.summary,
            resume.skills,
            resume.experience,
            resume.education,
            resume.file_name,
            resume.template_settings,
			resume.description,
			resume.defaultResume,
			resume.created,
			resume.updated,
            profile.profile_id,
            profile.phone_number,
            profile.email,
            profile.first_name,
            profile.middle_name,
            profile.last_name,
            profile.address_1,
            profile.address_2,
            profile.city,
            profile.state,
            profile.zip_code,
            profile.country,
            profile.social_accounts,
			profile.created,
			profile.updated
         FROM resume
         LEFT JOIN profile ON resume.profile_id = profile.profile_id
         WHERE resume.user_id = $1;`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resumes []*model.Resume
	for rows.Next() {
		var resume model.Resume
		var profile model.Profile
		err = rows.Scan(
			&resume.ResumeID,
			&resume.Sections,
			&resume.Summary,
			&resume.Skills,
			&resume.Experience,
			&resume.Education,
			&resume.FileName,
			&resume.TemplateSettings,
			&resume.Description,
			&resume.Default,
			&resume.Created,
			&resume.Updated,
			&profile.ProfileID,
			&profile.PhoneNumber,
			&profile.Email,
			&profile.FirstName,
			&profile.MiddleName,
			&profile.LastName,
			&profile.Address1,
			&profile.Address2,
			&profile.City,
			&profile.State,
			&profile.ZipCode,
			&profile.Country,
			&profile.SocialAccounts,
			&profile.Created,
			&profile.Updated,
		)
		if err != nil {
			return nil, err
		}
		resume.Profile = profile
		resumes = append(resumes, &resume)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return resumes, nil
}

func (r *DBResumeRepository) CreateResume(
	ctx context.Context,
	resume *model.Resume,
	userID string,
) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(
		ctx,
		`INSERT INTO profile (
			phone_number,
			email,
			first_name,
			middle_name,
			last_name,
			address_1,
			address_2,
			city,
			state,
			zip_code,
			country,
			social_accounts
		 ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		 RETURNING profile_id, created, updated`,
		resume.Profile.PhoneNumber,
		resume.Profile.Email,
		resume.Profile.FirstName,
		resume.Profile.MiddleName,
		resume.Profile.LastName,
		resume.Profile.Address1,
		resume.Profile.Address2,
		resume.Profile.City,
		resume.Profile.State,
		resume.Profile.ZipCode,
		resume.Profile.Country,
		resume.Profile.SocialAccounts,
	).Scan(
		&resume.Profile.ProfileID,
		&resume.Profile.Created,
		&resume.Profile.Updated,
	)
	if err != nil {
		return err
	}

	err = tx.QueryRow(
		ctx,
		`INSERT INTO resume (
			user_id,
			profile_id,
			sections,
			summary,
			skills,
			experience,
			education,
			file_name,
			template_settings,
			description
		 ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		 RETURNING resume_id, created, updated`,
		userID,
		resume.Profile.ProfileID,
		resume.Sections,
		resume.Summary,
		resume.Skills,
		resume.Experience,
		resume.Education,
		resume.FileName,
		resume.TemplateSettings,
		resume.Description,
	).Scan(
		&resume.ResumeID,
		&resume.Created,
		&resume.Updated,
	)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (r *DBResumeRepository) UpdateResume(
	ctx context.Context,
	resume *model.Resume,
	userID string,
) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(
		ctx,
		`UPDATE resume SET
			education = $1,
			experience = $2,
			skills = $3,
			description = $4,
			defaultResume = $5,
			summary = $6,
			file_name = $7,
			sections = $8,
			template_settings = $9
		 WHERE resume_id = $10 AND user_id = $11 RETURNING created, updated`,
		resume.Education,
		resume.Experience,
		resume.Skills,
		resume.Description,
		resume.Default,
		resume.Summary,
		resume.FileName,
		resume.Sections,
		resume.TemplateSettings,
		resume.ResumeID,
		userID,
	).Scan(&resume.Created, &resume.Updated)

	if err != nil {
		return err
	}

	err = tx.QueryRow(
		ctx,
		`UPDATE profile SET
            phone_number = $1,
            email = $2,
            first_name = $3,
            middle_name = $4,
            last_name = $5,
            address_1 = $6,
            address_2 = $7,
            city = $8,
            state = $9,
            zip_code = $10,
            country = $11,
            social_accounts = $12
         WHERE profile_id = $13 RETURNING created, updated`,
		resume.Profile.PhoneNumber,
		resume.Profile.Email,
		resume.Profile.FirstName,
		resume.Profile.MiddleName,
		resume.Profile.LastName,
		resume.Profile.Address1,
		resume.Profile.Address2,
		resume.Profile.City,
		resume.Profile.State,
		resume.Profile.ZipCode,
		resume.Profile.Country,
		resume.Profile.SocialAccounts,
		resume.Profile.ProfileID,
	).Scan(&resume.Profile.Created, &resume.Profile.Updated)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *DBResumeRepository) DeleteResumes(
	ctx context.Context,
	resumeIDs []string,
	userID string,
) ([]string, error) {
	deletedIDs := make([]string, 0)

	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	for i, resumeID := range resumeIDs {
		savePoint := pgx.Identifier{"sp_" + fmt.Sprintf("%d", i)}
		_, err := tx.Exec(ctx, "SAVEPOINT "+savePoint.Sanitize())
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec(
			ctx,
			`DELETE FROM resume WHERE resume_id = $1 AND user_id = $2`,
			resumeID,
			userID,
		)

		if err != nil {
			if IsForeignKeyConstraint(err) {
				_, rollbackErr := tx.Exec(
					ctx,
					"ROLLBACK TO SAVEPOINT "+savePoint.Sanitize(),
				)
				if rollbackErr != nil {
					return nil, rollbackErr
				}
				continue
			} else {
				return nil, err
			}
		}

		_, releaseErr := tx.Exec(
			ctx,
			"RELEASE SAVEPOINT "+savePoint.Sanitize(),
		)
		if releaseErr != nil {
			return nil, releaseErr
		}

		deletedIDs = append(deletedIDs, resumeID)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return deletedIDs, nil
}

func IsForeignKeyConstraint(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
