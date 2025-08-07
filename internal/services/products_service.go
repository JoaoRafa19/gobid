package services

import (
	"context"
	"errors"
	"github.com/JoaoRafa19/gobid/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type ProductsService struct {
	pool    *pgxpool.Pool
	queries *pgstore.Queries
}

var ErrProductNotFound = errors.New("product not found")

func NewProductsService(pool *pgxpool.Pool) *ProductsService {
	return &ProductsService{
		pool:    pool,
		queries: pgstore.New(pool),
	}
}

func (p *ProductsService) CreateProduct(
	ctx context.Context,
	sellerId uuid.UUID,
	productName,
	description string,
	basePrice float64,
	auctionEnd time.Time,
) (uuid.UUID, error) {
	id, err := p.queries.CreateProduct(ctx, pgstore.CreateProductParams{
		SellerID:    sellerId,
		ProductName: productName,
		Description: description,
		BasePrice:   basePrice,
		AuctionEnd:  auctionEnd,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (p *ProductsService) GetProductById(ctx context.Context, id uuid.UUID) (pgstore.Product, error) {
	product, err := p.queries.GetProductById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Product{}, ErrProductNotFound
		}
		return pgstore.Product{}, err
	}

	return product, nil
}
