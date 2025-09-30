package api

import (
	"errors"
	"net/http"

	"github.com/JoaoRafa19/gobid/internal/jsonutils"
	"github.com/JoaoRafa19/gobid/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// handleSubscribeUserToAuction godoc
// @Summary      Subscribe to a product auction via WebSocket
// @Description  Establishes a WebSocket connection for real-time auction bidding. Once connected, clients can send and receive messages with specific 'kind' types for bidding operations.
// @Description
// @Description  **Message Structure:**
// @Description  All messages follow this JSON format:
// @Description  ```json
// @Description  {
// @Description    "kind": 0,
// @Description    "message": "string",
// @Description    "amount": 0.0,
// @Description    "user_id": "uuid"
// @Description  }
// @Description  ```
// @Description
// @Description  **Message Types (Kind):**
// @Description  - `0` (PlaceBid): Client→Server - Place a new bid
// @Description  - `1` (SuccessfullyPlacedBid): Server→Client - Bid was accepted
// @Description  - `2` (NewBidPlaced): Server→Clients - New bid from another user
// @Description  - `3` (AuctionFinished): Server→Clients - Auction has ended
// @Description  - `4` (FailedToPlaceBid): Server→Client - Bid was rejected
// @Description  - `5` (InvalidBody): Server→Client - Invalid message format
// @Description
// @Description  **Example Messages:**
// @Description
// @Description  Place bid (Client→Server):
// @Description  ```json
// @Description  {"kind": 0, "amount": 150.50}
// @Description  ```
// @Description
// @Description  Bid success response (Server→Client):
// @Description  ```json
// @Description  {"kind": 1, "message": "Your bid has been placed!", "user_id": "uuid"}
// @Description  ```
// @Description
// @Description  New bid notification (Server→Others):
// @Description  ```json
// @Description  {"kind": 2, "message": "New bid has been placed!", "amount": 150.50, "user_id": "uuid"}
// @Description  ```
// @Tags         auctions
// @Param        product_id  path      string  true  "Product ID (UUID format)"
// @Success      101   {string}  string  "Switching Protocols - WebSocket connection established"
// @Failure      400   {object}  map[string]any  "Bad Request - Invalid product ID or auction ended"
// @Failure      401   {object}  map[string]any  "Unauthorized - Invalid or missing JWT token"
// @Failure      404   {object}  map[string]any  "Not Found - Product does not exist"
// @Failure      500   {object}  map[string]any  "Internal Server Error"
// @Security     ApiKeyAuth
// @Router       /products/ws/subscribe/{product_id} [get]
func (a *Api) handleSubscribeUserToAuction(w http.ResponseWriter, r *http.Request) {
	rawProductId := chi.URLParam(r, "product_id")

	productId, err := uuid.Parse(rawProductId)
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"message": "invalid product id must be a valid UUID",
		})
		return
	}

	_, err = a.ProductsService.GetProductById(r.Context(), productId)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			_ = jsonutils.EncodeJson(w, r, http.StatusNotFound, map[string]any{
				"message": "product not found",
			})
			return
		}
		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "unexpected error",
		})
		return
	}

	// CORRIGIDO: Obter o ID do usuário do contexto da requisição (injetado pelo middleware JWT).
	rawUserID := r.Context().Value(AuthUserIdKey)
	userIDString, ok := rawUserID.(string)
	if !ok {
		_ = jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{
			"error": "unauthorized: invalid token claims",
		})
		return
	}

	userid, err := uuid.Parse(userIDString)
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{
			"error": "unauthorized: user id in token is not a valid uuid",
		})
		return
	}

	a.AuctionLoby.Lock()
	room, ok := a.AuctionLoby.Rooms[productId]
	a.AuctionLoby.Unlock()

	if !ok {
		_ = jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"message": "the auction has ended",
		})
		return
	}

	conn, err := a.WsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		// O método Upgrade já trata a resposta de erro, então não podemos escrever JSON aqui.
		return
	}

	client := services.NewClient(room, conn, userid)

	room.Register <- client

	go client.ReadEventLoop()
	go client.WriteEventLoop()
}
