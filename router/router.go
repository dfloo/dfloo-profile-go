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

type ProfileDetail struct {
	ID             string     `json:"profileId"`
	UserID         string     `json:"userId"`
	Email          string     `json:"email"`
	FirstName      string     `json:"firstName"`
	LastName       string     `json:"lastName"`
	Address1       string     `json:"address1"`
	Address2       string     `json:"address2"`
	AddressCity    string     `json:"city"`
	AddressState   string     `json:"state"`
	ZipCode        string     `json:"zipCode"`
	Country        string     `json:"country"`
	SocialAccounts [][]string `json:"socialAccounts"`
}

func New() *httprouter.Router {
	router := httprouter.New()

	router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Access-Control-Request-Method") != "" {
			w.Header().Set("Access-Control-Allow-Methods", w.Header().Get("Allow"))
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "authorization, content-type")
		}

		w.WriteHeader(http.StatusNoContent)
	})

	router.GET("/api/profile", middleware.Logger(
		handlerToHandle(
			middleware.EnsureValidToken()(
				http.HandlerFunc(getProfileDetailHandler),
			),
		),
	))
	router.POST("/api/profile", middleware.Logger(
		handlerToHandle(
			middleware.EnsureValidToken()(
				http.HandlerFunc(postProfileDetailHandler),
			),
		),
	))

	return router
}

func getProfileDetailHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("getProfileDetailHandler")
	w.Header().Set("Access-Control-Allow-Methods", w.Header().Get("Allow"))
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
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
	log.Printf("claims %v %v", claims, ok)
	if !ok || claims == nil || claims.RegisteredClaims.Subject == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}

	var profileDetail ProfileDetail
	row := conn.QueryRow(
		context.Background(),
		`SELECT id, email, first_name, last_name, address_1, address_2, address_city, address_state, zip_code, country, social_accounts
		 FROM profile_detail WHERE user_id = $1;`,
		claims.RegisteredClaims.Subject,
	)
	log.Printf("row: %v", row)
	err = row.Scan(
		&profileDetail.ID,
		&profileDetail.Email,
		&profileDetail.FirstName,
		&profileDetail.LastName,
		&profileDetail.Address1,
		&profileDetail.Address2,
		&profileDetail.AddressCity,
		&profileDetail.AddressState,
		&profileDetail.ZipCode,
		&profileDetail.Country,
		&profileDetail.SocialAccounts,
	)
	if err != nil {
		log.Printf("select error %v", err)
		http.Error(w, "Failed to select profile detail", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(profileDetail)
	if err != nil {
		http.Error(w, "Failed to encode profile detail", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func postProfileDetailHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("postProfileDetailHandler")
	w.Header().Set("Access-Control-Allow-Methods", w.Header().Get("Allow"))
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "authorization, content-type")
	var profile ProfileDetail
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
	log.Printf("claims %v %v", claims, ok)
	if !ok || claims == nil || claims.RegisteredClaims.Subject == "" {
		http.Error(w, "User ID not found in token", http.StatusUnauthorized)
		return
	}
	profile.UserID = claims.RegisteredClaims.Subject
	log.Printf("UserID: %v", profile.UserID)

	_, err = conn.Exec(
		context.Background(),
		`INSERT INTO profile_detail (
	        user_id, email, first_name, last_name, address_1, address_2, address_city, address_state, zip_code, country, social_accounts
	    ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		profile.UserID,
		profile.Email,
		profile.FirstName,
		profile.LastName,
		profile.Address1,
		profile.Address2,
		profile.AddressCity,
		profile.AddressState,
		profile.ZipCode,
		profile.Country,
		profile.SocialAccounts,
	)
	if err != nil {
		http.Error(w, "Failed to insert profile detail", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message":"Profile detail created successfully"}`))
}

func handlerToHandle(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		h.ServeHTTP(w, r)
	}
}
