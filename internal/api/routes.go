package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (a *Api) BindRoutes() {
	// Aplicar middlewares globais - RequestID, Recoverer, Logger
	a.Router.Use(middleware.RequestID, middleware.Recoverer, middleware.Logger)

	a.Router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:3080/swagger/doc.json"),
	))

	a.Router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	a.Router.Route("/api", func(r chi.Router) {
		// API v1 - JWT Authentication
		r.Route("/v1", func(r chi.Router) {
			// Aplicar middleware WebSocket JWT para v1
			r.Use(a.WebSocketAuthMiddleware)

			r.Route("/users", func(r chi.Router) {
				r.Post("/signup", a.handleSignupUser)
				r.Post("/login", a.handleLoginUser)
			})
			r.Route("/products", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(a.JwtAuthMiddleware)
					r.Post("/", a.handleCreateProduct)
					r.Get("/ws/subscribe/{product_id}", a.handleSubscribeUserToAuction)
				})
			})
		})

		// API v2 - Cookie/Session Authentication
		r.Route("/v2", func(r chi.Router) {
			// Aplicar middleware de sessão para v2
			r.Use(a.Sessions.LoadAndSave)

			r.Route("/users", func(r chi.Router) {
				r.Post("/signup", a.handleSignupUser)
				r.Post("/login", a.handleLoginUserV2)
				r.Group(func(r chi.Router) {
					r.Use(a.SessionAuthMiddleware)
					r.Post("/logout", a.handleLogOutUserV2)
				})
			})
			r.Route("/products", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(a.SessionAuthMiddleware, a.WebSocketSessionAuthMiddleware)
					r.Post("/", a.handleCreateProduct)
					r.Get("/ws/subscribe/{product_id}", a.handleSubscribeUserToAuctionV2)
				})
			})
		})
	})
}
