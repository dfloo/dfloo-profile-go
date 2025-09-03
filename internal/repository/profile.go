package repository

import (
	"context"
	"os"

	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/jackc/pgx/v5"
)

type ProfileRepository struct {
	DBUrl string
}

func NewProfileRepository() *ProfileRepository {
	return &ProfileRepository{DBUrl: os.Getenv("DATABASE_URL")}
}

func (r *ProfileRepository) GetProfileByUserID(ctx context.Context, userID string) (*model.Profile, error) {
	conn, err := pgx.Connect(ctx, r.DBUrl)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	var profile model.Profile
	row := conn.QueryRow(
		ctx,
		`SELECT profile_id, phone_number, email, first_name, middle_name, last_name, address_1, address_2, city, state, zip_code, country, social_accounts
		 FROM profile WHERE user_id = $1;`,
		userID,
	)
	err = row.Scan(
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
	)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *ProfileRepository) CreateUserProfile(ctx context.Context, profile *model.Profile) error {
	conn, err := pgx.Connect(ctx, r.DBUrl)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(
		ctx,
		`INSERT INTO profile (
            user_id, phone_number, email, first_name, middle_name, last_name, address_1, address_2, city, state, zip_code, country, social_accounts
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		profile.UserID,
		profile.PhoneNumber,
		profile.Email,
		profile.FirstName,
		profile.MiddleName,
		profile.LastName,
		profile.Address1,
		profile.Address2,
		profile.City,
		profile.State,
		profile.ZipCode,
		profile.Country,
		profile.SocialAccounts,
	)
	return err
}

func (r *ProfileRepository) UpdateProfile(ctx context.Context, profile *model.Profile) error {
	conn, err := pgx.Connect(ctx, r.DBUrl)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(
		ctx,
		`UPDATE profile SET
            resume_id = $1,
            phone_number = $2,
            email = $3,
            first_name = $4,
            middle_name = $5,
            last_name = $6,
            address_1 = $7,
            address_2 = $8,
            city = $9,
            state = $10,
            zip_code = $11,
            country = $12,
            social_accounts = $13
        WHERE profile_id = $14`,
		profile.ResumeID,
		profile.PhoneNumber,
		profile.Email,
		profile.FirstName,
		profile.MiddleName,
		profile.LastName,
		profile.Address1,
		profile.Address2,
		profile.City,
		profile.State,
		profile.ZipCode,
		profile.Country,
		profile.SocialAccounts,
		profile.ProfileID,
	)
	return err
}
