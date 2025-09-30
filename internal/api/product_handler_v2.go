package api

import (
	"net/http"
)

// handleSubscribeUserToAuctionV2 godoc
// @Summary      Subscribe to auction updates via WebSocket (v2 - Cookie Auth)
// @Description  Establishes a WebSocket connection for real-time auction updates using cookie-based authentication.
// @Tags         products-v2
// @Param        product_id  path  string  true  "Product ID"
// @Router       /api/v2/products/ws/subscribe/{product_id} [get]
func (a *Api) handleSubscribeUserToAuctionV2(w http.ResponseWriter, r *http.Request) {
	// Obter user ID da sessão através do contexto (adicionado pelo middleware)
	_, ok := r.Context().Value(AuthUserIdKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Reutilizar a lógica existente do WebSocket
	a.handleSubscribeUserToAuction(w, r)
}
