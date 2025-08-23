// Package main holds the main app logic
package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mmycroft/boot-dev-chirpy/api"
	"github.com/mmycroft/boot-dev-chirpy/database"
)

const (
	_ROOT = "./"
	_PORT = 8080
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("SECRET")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)

	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal(err)
	}

	cfg := &api.APIConfig{
		FileServerHits: atomic.Int32{},
		DBQueries:      dbQueries,
		Templates:      templates,
		Platform:       platform,
		Secret:         secret,
	}

	mux := http.NewServeMux()

	mux.Handle("/app/", cfg.MiddlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(_ROOT)))))

	mux.HandleFunc("GET /admin/metrics", cfg.HandlerNumRequests)
	mux.HandleFunc("POST /admin/reset", cfg.HandlerResetNumRequests)

	mux.HandleFunc("GET /api/healthz", cfg.HandlerReadiness)

	mux.HandleFunc("POST /api/login", cfg.HandlerLogin)
	mux.HandleFunc("POST /api/refresh", cfg.HandlerRefresh)
	mux.HandleFunc("POST /api/revoke", cfg.HandlerRevoke)

	mux.HandleFunc("POST /api/users", cfg.HandlerCreateUser)
	mux.HandleFunc("GET /api/users", cfg.HandlerGetUsers)
	mux.HandleFunc("GET /api/users/{userID}", cfg.HandlerGetUser)
	mux.HandleFunc("PUT /api/users", cfg.HandlerUpdateUser)

	mux.HandleFunc("POST /api/chirps", cfg.HandlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", cfg.HandlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.HandlerGetChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.HandlerDeleteChirp)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", _PORT),
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %d\n", _ROOT, _PORT)
	log.Fatal(server.ListenAndServe())
}
