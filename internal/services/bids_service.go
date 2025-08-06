package services

import (
	"context"
	"errors"
	"github.com/JoaoRafa19/gobid/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
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
		if errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Bid{}, err
		}
		return pgstore.Bid{}, err
	}

	highestBid, err := b.queries.GetHighestBidByProductId(ctx, productId)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Bid{}, err
		}
	}

	if product.BasePrice >= amount || highestBid.Amount >= amount {
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
