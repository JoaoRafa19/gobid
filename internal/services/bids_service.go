package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/JoaoRafa19/gobid/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BidsService struct {
	pool    *pgxpool.Pool
	queries *pgstore.Queries
}

var ErrBidIsTooLow = errors.New("the bid value is too low")

func NewBidsService(pool *pgxpool.Pool) *BidsService {
	return &BidsService{
		pool:    pool,
		queries: pgstore.New(pool),
	}
}

func (b *BidsService) PlaceBid(ctx context.Context, productId, bidder uuid.UUID, amount float64) (pgstore.Bid, error) {
	product, err := b.queries.GetProductById(ctx, productId)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Bid{}, err
		}
	}

	bids, err := b.queries.GetHighestBidByProductId(ctx, productId)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Bid{}, nil
		}
	}
	var highestBid pgstore.Bid
	if len(bids) == 0 {
		highestBid = pgstore.Bid{}
	} else {
		highestBid = bids[0]
	}

	if product.BasePrice >= amount || highestBid.Amount >= amount {
		fmt.Println("HIGHEST BID", highestBid)
		fmt.Println("BID ", amount)
		return pgstore.Bid{}, ErrBidIsTooLow
	}

	highestBid, err = b.queries.CreateBid(ctx, pgstore.CreateBidParams{
		ProductID: productId,
		BidderID:  bidder,
		Amount:    amount,
	})

	if err != nil {
		return pgstore.Bid{}, err
	}

	return highestBid, nil
}
