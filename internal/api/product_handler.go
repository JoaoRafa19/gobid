package api

import (
	"context"
	"net/http"

	"github.com/JoaoRafa19/gobid/internal/jsonutils"
	"github.com/JoaoRafa19/gobid/internal/services"
	"github.com/JoaoRafa19/gobid/internal/usecase/product"
)

// handleCreateProduct godoc
// @Summary      Create a new product auction
// @Description  Creates a new product and starts an auction for it. This endpoint requires authentication.
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        product  body      product.CreateProductRequest  true  "Product creation information"
// @Success      201   {object}  map[string]any
// @Failure      422   {object}  map[string]any
// @Failure      500   {object}  map[string]any
// @Security     ApiKeyAuth
// @Router       /products [post]
func (a *Api) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[product.CreateProductRequest](r)

	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, problems)
		return
	}

	// Obter o ID do usuário do contexto (funciona para JWT e sessão)
	userId, err := a.GetUserIdFromContext(r)
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{
			"error": err.Error(),
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

	ctx, cancel := context.WithDeadline(context.Background(), data.AuctionEnd)

	auctionRoom := services.NewAuctionRoom(ctx, productId, a.BidsService, cancel)

	go auctionRoom.Start()

	a.AuctionLoby.Lock()
	a.AuctionLoby.Rooms[productId] = auctionRoom
	a.AuctionLoby.Unlock()

	_ = jsonutils.EncodeJson(w, r, http.StatusCreated, map[string]any{
		"message": "auction has started successfully",
		"product": productId,
	})

}
