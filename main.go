package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/viren-vii/rss-aggregator/internal/database"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	godotenv.Load()

	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT not found in the environment variables")
	}

	dbURL := os.Getenv("DB_URL")
	if portString == "" {
		log.Fatal("DB_URL not found in the environment variables")
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Can't connect to database", err)
	}

	db := database.New(conn)
	apiCfg := apiConfig{
		DB: db,
	}

	go startScraping(db, 10, time.Minute)

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	v1Router := chi.NewRouter()
	v1Router.Get("/healthz", handlerReadiness)
	v1Router.Get("/error", handlerError)
	v1Router.Get("/users", apiCfg.authMiddleware(apiCfg.handlerGetUser))
	v1Router.Post("/users", apiCfg.handlerCreateUser)

	v1Router.Get("/feeds", apiCfg.handlerGetFeeds)
	v1Router.Post("/feeds", apiCfg.authMiddleware(apiCfg.handlerCreateFeed))

	v1Router.Get("/posts", apiCfg.authMiddleware(apiCfg.handleGetUserPosts))

	v1Router.Post("/feeds/follows", apiCfg.authMiddleware(apiCfg.handlerCreateFeedFollow))
	v1Router.Get("/feeds/follows", apiCfg.authMiddleware(apiCfg.handlerGetFeedFollow))
	v1Router.Delete("/feeds/follows/{feedFollowID}", apiCfg.authMiddleware(apiCfg.handlerDeleteFeedFollow))

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler: router,
		Addr:    ":" + portString,
	}

	log.Printf("Server starting on port %s", portString)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}
