package product

import (
	"context"
	"github.com/JoaoRafa19/gobid/internal/validator"
	"github.com/google/uuid"
	"time"
)

type CreateProductRequest struct {
	SellerID    uuid.UUID `json:"seller_id"`
	ProductName string    `json:"product_name"`
	Description string    `json:"description"`
	BasePrice   float64   `json:"base_price"`
	AuctionEnd  time.Time `json:"auction_end"`
}

const minAuctionDuration = time.Hour * 2

func (c CreateProductRequest) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator = make(validator.Evaluator)

	eval.CheckField(validator.NotBlank(c.ProductName), "product_name", "product name can not be empty")
	eval.CheckField(validator.NotBlank(c.Description), "description", "description can not be empty")
	eval.CheckField(validator.MinChar(
		c.Description, 10) && validator.MaxChar(c.Description, 255),
		"description",
		"description requires length between 10 and 255 ",
	)
	eval.CheckField(c.BasePrice > 0, "base_price", "base price must be greater than 0")
	eval.CheckField(c.AuctionEnd.Sub(time.Now()) >= minAuctionDuration, "auction_end", "auction time must have at least 2 hours")

	return eval
}
