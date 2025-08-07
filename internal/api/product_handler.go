package api

import (
	"context"
	"github.com/JoaoRafa19/gobid/internal/jsonutils"
	"github.com/JoaoRafa19/gobid/internal/services"
	"github.com/JoaoRafa19/gobid/internal/usecase/product"
	"github.com/google/uuid"
	"net/http"
)

func (a *Api) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[product.CreateProductRequest](r)

	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, problems)
		return
	}
	userId, ok := a.Sessions.Get(r.Context(), "authUserId").(uuid.UUID)
	if !ok {
		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "unexpected error",
		})
		return
	}

	productId, err := a.ProductsService.CreateProduct(
		r.Context(),
		userId,
		data.ProductName,
		data.Description,
		data.BasePrice,
		data.AuctionEnd,
	)

	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "failed to create product auction",
		})
		return
	}

	ctx, _ := context.WithDeadline(context.Background(), data.AuctionEnd)

	auctionRoom := services.NewAuctionRoom(ctx, productId, a.BidsService)

	go auctionRoom.Start()

	a.AuctionLoby.Lock()
	a.AuctionLoby.Rooms[productId] = auctionRoom
	a.AuctionLoby.Unlock()

	_ = jsonutils.EncodeJson(w, r, http.StatusCreated, map[string]any{
		"message": "auction has started successfully",
		"product": productId,
	})

}
