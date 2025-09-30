package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/JoaoRafa19/gobid/docs"
	"github.com/JoaoRafa19/gobid/internal/api"
	"github.com/JoaoRafa19/gobid/internal/services"
	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func init() {
	gob.Register(uuid.UUID{})
}

// @title GoBid API
// @version 1.0
// @description This is a sample server for a bid application.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3080
// @BasePath /
func main() {

	// Load .env file only in local development (optional)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found (normal in production)")
	}

	// Configure timezone based on TZ environment variable
	tz := os.Getenv("TZ")
	if tz != "" {
		location, err := time.LoadLocation(tz)
		if err != nil {
			log.Printf("Warning: Could not load timezone %s: %v. Using UTC.", tz, err)
		} else {
			time.Local = location
			log.Printf("Timezone set to: %s", tz)
		}
	}

	jwtSecret := os.Getenv("GOBID_JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("GOBID_JWT_SECRET environment variable not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=require",
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

	// Configurar session manager para API v2
	sessionManager := scs.New()
	sessionManager.Store = pgxstore.New(pool)
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode

	a := api.Api{
		JwtSecret:       jwtSecret,
		Sessions:        sessionManager,
		UserService:     services.NewUsersService(pool),
		ProductsService: services.NewProductsService(pool),
		BidsService:     services.NewBidsService(pool),
		Router:          chi.NewMux(),
		WsUpgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		AuctionLoby: services.AuctionLobby{
			Rooms: make(map[uuid.UUID]*services.AuctionRoom),
		},
	}

	port := os.Getenv("GOBID_APP_PORT")
	if port == "" {
		port = "3080"
	}

	a.BindRoutes()
	log.Printf("Server is running on port: %s", port)
	if err := http.ListenAndServe(":"+port, a.Router); err != nil {
		panic(err)
	}
}
