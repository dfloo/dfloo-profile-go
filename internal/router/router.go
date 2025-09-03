package router

import (
	"net/http"
	"os"

	"github.com/dfloo/dfloo-profile-go/internal/handler"
	"github.com/dfloo/dfloo-profile-go/internal/middleware"
	"github.com/dfloo/dfloo-profile-go/internal/repository"
	"github.com/julienschmidt/httprouter"
)

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

	repo := repository.NewProfileRepository()
	profileHandler := handler.NewProfileHandler(repo)

	router.GET("/api/profile", middleware.CORS(
		middleware.Logger(
			handlerToHandle(
				middleware.EnsureValidToken()(
					http.HandlerFunc(profileHandler.GetUserProfile),
				),
			),
		),
	))
	router.POST("/api/profile", middleware.CORS(
		middleware.Logger(
			handlerToHandle(
				middleware.EnsureValidToken()(
					http.HandlerFunc(profileHandler.PostUserProfile),
				),
			),
		),
	))

	router.PUT("/api/profile", middleware.CORS(
		middleware.Logger(
			handlerToHandle(
				middleware.EnsureValidToken()(
					http.HandlerFunc(profileHandler.PutUserProfile),
				),
			),
		),
	))

	return router
}

func handlerToHandle(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		h.ServeHTTP(w, r)
	}
}
