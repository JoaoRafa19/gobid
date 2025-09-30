// @title           GoBid API
// @version         1.0
// @description     This is a sample server for a bidding application.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:3080
// @BasePath  /api/v1

// @securityDefinitions.apiKey  ApiKeyAuth
// @in header
// @name Authorization
package api

import (
	"github.com/JoaoRafa19/gobid/internal/services"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

type Api struct {
	Router     *chi.Mux
	WsUpgrader *websocket.Upgrader
	JwtSecret  string              // Chave para assinar os JWTs (v1)
	Sessions   *scs.SessionManager // Session manager para cookies (v2)

	UserService     *services.UsersService
	BidsService     *services.BidsService
	ProductsService *services.ProductsService
	AuctionLoby     services.AuctionLobby
}
