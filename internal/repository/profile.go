package repository

import (
	"context"

	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileRepository interface {
	GetProfileByUserID(ctx context.Context, userID string) (*model.Profile, error)
	CreateUserProfile(ctx context.Context, profile *model.Profile, userID string) error
	UpdateProfile(ctx context.Context, profile *model.Profile, userID string) error
}

type DBProfileRepository struct {
	Pool *pgxpool.Pool
}

func NewDBProfileRepository(pool *pgxpool.Pool) *DBProfileRepository {
	return &DBProfileRepository{Pool: pool}
}

func (r *DBProfileRepository) GetProfileByUserID(
	ctx context.Context,
	userID string,
) (*model.Profile, error) {
	var profile model.Profile
	row := r.Pool.QueryRow(
		ctx,
		`SELECT
			profile_id,
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
			social_accounts,
			created,
			updated
		 FROM profile WHERE user_id = $1;`,
		userID,
	)
	err := row.Scan(
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
	return &profile, nil
}

func (r *DBProfileRepository) CreateUserProfile(
	ctx context.Context,
	profile *model.Profile,
	userID string,
) error {
	err := r.Pool.QueryRow(
		ctx,
		`INSERT INTO profile (
            user_id,
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
         ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		 RETURNING profile_id, created, updated`,
		userID,
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
	).Scan(
		&profile.ProfileID,
		&profile.Created,
		&profile.Updated,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *DBProfileRepository) UpdateProfile(
	ctx context.Context,
	profile *model.Profile,
	userID string,
) error {
	err := r.Pool.QueryRow(
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
        WHERE user_id = $13 RETURNING created, updated`,
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
		userID,
	).Scan(&profile.Created, &profile.Updated)
	if err != nil {
		return err
	}

	return nil
}
