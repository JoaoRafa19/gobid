package api

import (
	"github.com/JoaoRafa19/gobid/internal/services"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

type Api struct {
	Router     *chi.Mux
	Sessions   *scs.SessionManager
	WsUpgrader *websocket.Upgrader

	UserService     *services.UsersService
	BidsService     *services.BidsService
	ProductsService *services.ProductsService
	AuctionLoby     services.AuctionLobby
}
