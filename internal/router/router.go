package router

import (
	"net/http"

	"github.com/dfloo/dfloo-profile-go/internal/handler"
	"github.com/dfloo/dfloo-profile-go/internal/middleware"
	"github.com/dfloo/dfloo-profile-go/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func New(pool *pgxpool.Pool) *http.ServeMux {
	mux := http.NewServeMux()

	profileRepo := repository.NewDBProfileRepository(pool)
	profileHandler := handler.NewProfileHandler(profileRepo)
	resumeRepo := repository.NewDBResumeRepository(pool)
	resumeHandler := handler.NewResumeHandler(resumeRepo)
	jobApplicationRepo := repository.NewDBJobApplicationRepository(pool)
	jobApplicationHandler := handler.NewJobApplicationHandler(jobApplicationRepo)
	healthHandler := handler.NewHealthHandler(pool)

	mux.HandleFunc("GET /health", healthHandler.HealthCheck)
	mux.Handle("GET /metrics", promhttp.Handler())

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

	mux.HandleFunc("/api/resumes", func(w http.ResponseWriter, r *http.Request) {
		middleware.CoreAuthenticated(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					resumeHandler.GetUserResumes(w, r)
				case http.MethodPost:
					resumeHandler.PostResume(w, r)
				case http.MethodPut:
					resumeHandler.PutResume(w, r)
				case http.MethodDelete:
					resumeHandler.DeleteResumes(w, r)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			}),
		).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/resumes/default", func(w http.ResponseWriter, r *http.Request) {
		middleware.CoreAuthenticated(
			http.HandlerFunc(resumeHandler.SetDefaultResume),
		).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/download/resume/default", func(w http.ResponseWriter, r *http.Request) {
		middleware.Core(
			http.HandlerFunc(resumeHandler.DownloadDefaultResumePDF),
		).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/download/resume", func(w http.ResponseWriter, r *http.Request) {
		middleware.Core(
			http.HandlerFunc(resumeHandler.DownloadResumePDF),
		).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/job-applications", func(w http.ResponseWriter, r *http.Request) {
		middleware.CoreAuthenticated(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					jobApplicationHandler.GetUserJobApplications(w, r)
				case http.MethodPost:
					jobApplicationHandler.PostJobApplication(w, r)
				case http.MethodPut:
					jobApplicationHandler.PutJobApplications(w, r)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			}),
		).ServeHTTP(w, r)
	})

	return mux
}
