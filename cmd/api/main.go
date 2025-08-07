package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"github.com/JoaoRafa19/gobid/internal/api"
	"github.com/JoaoRafa19/gobid/internal/services"
	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"time"
)

func init() {
	gob.Register(uuid.UUID{})
}

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		os.Getenv("GOBID_DATABASE_USER"),
		os.Getenv("GOBID_DATABASE_PASSWORD"),
		os.Getenv("GOBID_DATABASE_NAME"),
		os.Getenv("GOBID_DATABASE_HOST"),
		os.Getenv("GOBID_DATABASE_PORT"),
	))

	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Could not ping database: %v", err)
	}

	s := scs.New()
	s.Store = pgxstore.New(pool)
	s.Lifetime = 24 * time.Hour
	s.Cookie.HttpOnly = true
	s.Cookie.SameSite = http.SameSiteLaxMode

	a := api.Api{
		UserService:     services.NewUsersService(pool),
		ProductsService: services.NewProductsService(pool),
		BidsService:     services.NewBidsService(pool),
		Router:          chi.NewMux(),
		WsUpgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		Sessions: s,
		AuctionLoby: services.AuctionLobby{
			Rooms: make(map[uuid.UUID]*services.AuctionRoom),
		},
	}

	a.BindRoutes()
	fmt.Println("Server is running on port 3080")
	if err := http.ListenAndServe("0.0.0.0:3080", a.Router); err != nil {
		panic(err)
	}
}
