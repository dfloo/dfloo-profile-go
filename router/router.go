package router

import (
	"net/http"

	"github.com/dfloo/dfloo-profile-go/middleware"
	"github.com/julienschmidt/httprouter"
)

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

	protectedHandler := middleware.Logger(
		handlerToHandle(
			middleware.EnsureValidToken()(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					getPrivateHandler(w, r, httprouter.ParamsFromContext(r.Context()))
				}),
			),
		),
	)

	router.GET("/api/private", protectedHandler)

	return router
}

func getPrivateHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	w.Header().Set("Access-Control-Allow-Methods", w.Header().Get("Allow"))
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "authorization, content-type")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Hello from a private endpoint! You need to be authenticated to see this."}`))
}

func handlerToHandle(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		h.ServeHTTP(w, r)
	}
}
