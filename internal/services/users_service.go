package services

import (
	"context"
	"errors"
	"github.com/JoaoRafa19/gobid/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type UsersService struct {
	queries *pgstore.Queries
	pool    *pgxpool.Pool
}

func NewUsersService(pool *pgxpool.Pool) *UsersService {
	return &UsersService{
		pool:    pool,
		queries: pgstore.New(pool),
	}
}

var (
	ErrDuplicatedEmailOrUsername = errors.New("email or username already in use")
	ErrInvalidCredentials        = errors.New("invalid credentials")
)

func (us *UsersService) CreateUser(ctx context.Context, userName, email, password, bio string) (uuid.UUID, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return uuid.Nil, err
	}

	args := pgstore.CreateUserParams{
		UserName:     userName,
		Email:        email,
		PasswordHash: hash,
		Bio:          bio,
	}
	id, err := us.queries.CreateUser(ctx, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.Nil, ErrDuplicatedEmailOrUsername
		}
		return uuid.Nil, err
	}

	return id, nil
}

func (us *UsersService) AuthenticateUser(ctx context.Context, email, password string) (uuid.UUID, error) {
	user, err := us.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrInvalidCredentials
		}
		return uuid.Nil, err
	}
	err = bcrypt.CompareHashAndPassword(
		user.PasswordHash,
		[]byte(password),
	)
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return uuid.Nil, ErrInvalidCredentials
		}
		return uuid.Nil, err
	}

	return user.ID, nil
}
