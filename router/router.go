package router

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/dfloo/dfloo-profile-go/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/julienschmidt/httprouter"
)

type Profile struct {
	ProfileID      string   `json:"profileId"`
	UserID         string   `json:"userId"`
	ResumeID       string   `json:"resumeId"`
	PhoneNumber    string   `json:"phoneNumber"`
	Email          string   `json:"email"`
	FirstName      string   `json:"firstName"`
	MiddleName     string   `json:"middleName"`
	LastName       string   `json:"lastName"`
	Address1       string   `json:"address1"`
	Address2       string   `json:"address2"`
	City           string   `json:"city"`
	State          string   `json:"state"`
	ZipCode        string   `json:"zipCode"`
	Country        string   `json:"country"`
	SocialAccounts []string `json:"socialAccounts"`
}

func New() *httprouter.Router {
	router := httprouter.New()

	router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Access-Control-Request-Method") != "" {
			w.Header().Set("Access-Control-Allow-Methods", w.Header().Get("Allow"))
			w.Header().Set("Access-Control-Allow-Origin", os.Getenv("CLIENT_ORIGIN"))
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "authorization, content-type")
		}

		w.WriteHeader(http.StatusNoContent)
	})

	router.GET("/api/profile", middleware.Logger(
		handlerToHandle(
			middleware.EnsureValidToken()(
				http.HandlerFunc(getUserProfileHandler),
			),
		),
	))
	router.POST("/api/profile", middleware.Logger(
		handlerToHandle(
			middleware.EnsureValidToken()(
				http.HandlerFunc(postUserProfileHandler),
			),
		),
	))
	router.PUT("/api/profile", middleware.Logger(
		handlerToHandle(
			middleware.EnsureValidToken()(
				http.HandlerFunc(putUserProfileHandler),
			),
		),
	))

	return router
}

func getUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Methods", w.Header().Get("Allow"))
	w.Header().Set("Access-Control-Allow-Origin", os.Getenv("CLIENT_ORIGIN"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "authorization, content-type")
	w.Header().Set("Content-Type", "application/json")

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Printf("database connection failed %v", err)
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer conn.Close(context.Background())

	claims, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	if !ok || claims == nil || claims.RegisteredClaims.Subject == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}

	var profile Profile
	row := conn.QueryRow(
		context.Background(),
		`SELECT profile_id, phoneNumber, email, first_name, middle_name, last_name, address_1, address_2, city, state, zip_code, country, social_accounts
		 FROM profile WHERE user_id = $1;`,
		claims.RegisteredClaims.Subject,
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
		log.Printf("select error %v", err)
		http.Error(w, "Failed to select profile detail", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(profile)
	if err != nil {
		http.Error(w, "Failed to encode profile detail", http.StatusInternalServerError)
	}
}

func postUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Methods", w.Header().Get("Allow"))
	w.Header().Set("Access-Control-Allow-Origin", os.Getenv("CLIENT_ORIGIN"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "authorization, content-type")
	var profile Profile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		log.Printf("json decoding failed %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Printf("database connection failed %v", err)
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer conn.Close(context.Background())

	claims, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	if !ok || claims == nil || claims.RegisteredClaims.Subject == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}

	_, err = conn.Exec(
		context.Background(),
		`INSERT INTO profile (
	        user_id, first_name, middle_name last_name, address_1, address_2, city, state, zip_code, country, email, social_accounts, phone_number
	    ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		claims.RegisteredClaims.Subject,
		profile.FirstName,
		profile.LastName,
		profile.Address1,
		profile.Address2,
		profile.City,
		profile.State,
		profile.ZipCode,
		profile.Country,
		profile.Email,
		profile.SocialAccounts,
		profile.PhoneNumber,
	)
	if err != nil {
		http.Error(w, "Failed to insert profile detail", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message":"Profile detail created successfully"}`))
}

func putUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Methods", w.Header().Get("Allow"))
	w.Header().Set("Access-Control-Allow-Origin", os.Getenv("CLIENT_ORIGIN"))
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "authorization, content-type")
	w.Header().Set("Content-Type", "application/json")
	var profile Profile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		log.Printf("json decoding failed %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Printf("database connection failed %v", err)
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer conn.Close(context.Background())

	claims, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	if !ok || claims == nil || claims.RegisteredClaims.Subject == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}

	_, err = conn.Exec(
		context.Background(),
		`UPDATE profile
		 SET first_name = $1, middle_name = $2, last_name = $3, address_1 = $4, address_2 = $5, city = $6, state = $7, zip_code = $8, country = $9, email = $10, social_accounts = $11, phone_number = $12
		 WHERE user_id = $13`,
		profile.FirstName,
		profile.MiddleName,
		profile.LastName,
		profile.Address1,
		profile.Address2,
		profile.City,
		profile.State,
		profile.ZipCode,
		profile.Country,
		profile.Email,
		profile.SocialAccounts,
		profile.PhoneNumber,
		claims.RegisteredClaims.Subject,
	)
	if err != nil {
		http.Error(w, "Failed to update profile detail", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message":"Profile detail updated successfully"}`))
}

func handlerToHandle(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		h.ServeHTTP(w, r)
	}
}
