package repository

import (
	"context"

	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ResumeRepository struct {
	Pool *pgxpool.Pool
}

func NewResumeRepository(pool *pgxpool.Pool) *ResumeRepository {
	return &ResumeRepository{Pool: pool}
}

func (r *ResumeRepository) GetResumesByUserID(
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

func (r *ResumeRepository) CreateResume(
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

func (r *ResumeRepository) UpdateResume(
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
			summary = $5,
			file_name = $6,
			sections = $7,
			template_settings = $8
		 WHERE resume_id = $9 AND user_id = $10 RETURNING updated`,
		resume.Education,
		resume.Experience,
		resume.Skills,
		resume.Description,
		resume.Summary,
		resume.FileName,
		resume.Sections,
		resume.TemplateSettings,
		resume.ResumeID,
		userID,
	).Scan(&resume.Updated)

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
         WHERE profile_id = $13 RETURNING updated`,
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
	).Scan(&resume.Profile.Updated)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *ResumeRepository) DeleteResumes(
	ctx context.Context,
	resumeIDs []string,
	userID string,
) error {
	_, err := r.Pool.Exec(
		ctx,
		`DELETE FROM resume WHERE resume_id = ANY($1) AND user_id = $2`,
		resumeIDs,
		userID,
	)
	return err
}
