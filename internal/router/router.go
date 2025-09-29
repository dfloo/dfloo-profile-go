package router

import (
	"net/http"

	"github.com/dfloo/dfloo-profile-go/internal/handler"
	"github.com/dfloo/dfloo-profile-go/internal/middleware"
	"github.com/dfloo/dfloo-profile-go/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(pool *pgxpool.Pool) *http.ServeMux {
	mux := http.NewServeMux()

	profileRepo := repository.NewProfileRepository(pool)
	profileHandler := handler.NewProfileHandler(profileRepo)

	mux.HandleFunc("/api/profiles", func(w http.ResponseWriter, r *http.Request) {
		middleware.CoreAuthenticated(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					profileHandler.GetUserProfile(w, r)
					return
				case http.MethodPost:
					profileHandler.PostUserProfile(w, r)
					return
				case http.MethodPut:
					profileHandler.PutUserProfile(w, r)
					return
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			}),
		).ServeHTTP(w, r)
	})

	return mux
}
