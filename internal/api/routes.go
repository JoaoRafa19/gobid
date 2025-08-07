package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (a *Api) BindRoutes() {
	a.Router.Use(middleware.RequestID, middleware.Recoverer, middleware.Logger, a.Sessions.LoadAndSave)

	/*csrfMiddleware := csrf.Protect(
		[]byte(os.Getenv("GOBID_CSRF_KEY")),
		csrf.Secure(false), // DEV ONLY

	)
	a.Router.Use(csrfMiddleware)*/

	a.Router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Get("/csrftoken", a.HandleCSRFToken)
			r.Route("/users", func(r chi.Router) {
				r.Post("/signup", a.handleSignupUser)
				r.Post("/login", a.handleLoginUser)
				r.Group(func(r chi.Router) {
					r.Use(a.AuthMiddleware)
					r.Post("/logout", a.handleLogOutUser)
				})
			})
			r.Route("/products", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(a.AuthMiddleware)
					r.Post("/", a.handleCreateProduct)
					r.Get("/ws/subscribe/{product_id}", a.handleSubscribeUserToAuction)
				})
			})
		})
	})
}
