package api

import "github.com/go-chi/chi/v5"

func (a *Api) BindRoutes() {
	a.router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Route("/users", func(r chi.Router) {
				r.Post("/signup", a.handleSignupUser)
				r.Post("/signup", a.handleSignupUser)
				r.Post("/logout", a.handleSignupUser)
			})
		})
	})
}
