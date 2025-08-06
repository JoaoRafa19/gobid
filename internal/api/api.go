package api

import (
	"github.com/JoaoRafa19/gobid/internal/services"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

type Api struct {
	Router   *chi.Mux
	Sessions *scs.SessionManager

	UserService     *services.UsersService
	ProductsService *services.ProductsService
}
