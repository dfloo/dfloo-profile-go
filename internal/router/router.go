package router

import (
	"net/http"

	"github.com/dfloo/dfloo-profile-go/internal/email"
	"github.com/dfloo/dfloo-profile-go/internal/handler"
	"github.com/dfloo/dfloo-profile-go/internal/middleware"
	"github.com/dfloo/dfloo-profile-go/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(pool *pgxpool.Pool) *http.ServeMux {
	mux := http.NewServeMux()

	emailSvc := email.NewService()

	profileRepo := repository.NewDBProfileRepository(pool)
	profileHandler := handler.NewProfileHandler(profileRepo)
	resumeRepo := repository.NewDBResumeRepository(pool)
	resumeHandler := handler.NewResumeHandler(resumeRepo)
	jobApplicationRepo := repository.NewDBJobApplicationRepository(pool)
	jobApplicationHandler := handler.NewJobApplicationHandler(jobApplicationRepo)
	meetingRequestRepo := repository.NewDBMeetingRequestRepository(pool)
	meetingRequestHandler := handler.NewMeetingRequestHandler(meetingRequestRepo, emailSvc)
	signupRequestRepo := repository.NewDBSignupRequestRepository(pool)
	signupRequestHandler := handler.NewSignupRequestHandler(signupRequestRepo, emailSvc)
	healthHandler := handler.NewHealthHandler(pool)
	f1Repo := repository.NewDBF1Repository(pool)
	f1Handler := handler.NewF1Handler(f1Repo)

	mux.HandleFunc("GET /health", healthHandler.HealthCheck)

	mux.HandleFunc("/api/meeting-requests", func(w http.ResponseWriter, r *http.Request) {
		middleware.Core(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodPost:
					meetingRequestHandler.PostMeetingRequest(w, r)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			}),
		).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/signup-requests", func(w http.ResponseWriter, r *http.Request) {
		middleware.Core(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodPost:
					signupRequestHandler.PostSignupRequest(w, r)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			}),
		).ServeHTTP(w, r)
	})

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

	mux.HandleFunc("/api/resumes/download/{resumeId}", func(w http.ResponseWriter, r *http.Request) {
		middleware.CoreAuthenticated(
			http.HandlerFunc(resumeHandler.DownloadResumePDF),
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
			http.HandlerFunc(resumeHandler.DownloadGuestResumePDF),
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

	mux.HandleFunc("/api/f1/championships", func(w http.ResponseWriter, r *http.Request) {
		middleware.Core(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					f1Handler.GetChampionships(w, r)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			}),
		).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/f1/drivers", func(w http.ResponseWriter, r *http.Request) {
		middleware.Core(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					f1Handler.GetDrivers(w, r)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			}),
		).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/f1/drivers/{id}", func(w http.ResponseWriter, r *http.Request) {
		middleware.Core(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					f1Handler.GetDriverDetails(w, r)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			}),
		).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/f1/constructors", func(w http.ResponseWriter, r *http.Request) {
		middleware.Core(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					f1Handler.GetConstructors(w, r)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			}),
		).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/f1/events", func(w http.ResponseWriter, r *http.Request) {
		middleware.Core(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					f1Handler.GetEvents(w, r)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			}),
		).ServeHTTP(w, r)
	})

	return mux
}
