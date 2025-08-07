package api

import (
	"errors"
	"github.com/JoaoRafa19/gobid/internal/jsonutils"
	"github.com/JoaoRafa19/gobid/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
)

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

	userid, ok := a.Sessions.Get(r.Context(), "authUserId").(uuid.UUID)
	if !ok {
		_ = jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{
			"error": "unauthorized",
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
		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "connection error",
		})
		return
	}

	client := services.NewClient(room, conn, userid)

	room.Register <- client

	go client.ReadEventLoop()
	go client.WriteEventLoop()
}
