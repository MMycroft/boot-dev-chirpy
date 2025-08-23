// Package api holds api stuff
package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/mmycroft/boot-dev-chirpy/auth"
	"github.com/mmycroft/boot-dev-chirpy/database"
)

type APIConfig struct {
	FileServerHits atomic.Int32
	DBQueries      *database.Queries
	Templates      *template.Template
	Platform       string
	Secret         string
}

func (cfg *APIConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *APIConfig) MiddlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		log.Printf("%s %s", req.Method, req.URL.Path)
		next.ServeHTTP(wr, req)
	})
}

// HandlerNumRequests GET /admin/metrics
func (cfg *APIConfig) HandlerNumRequests(wr http.ResponseWriter, req *http.Request) {
	name := "admin.html"
	data := struct {
		Hits int32
	}{
		Hits: cfg.FileServerHits.Load(),
	}
	err := cfg.Templates.ExecuteTemplate(wr, name, data)
	if err != nil {
		log.Printf("error executing template: %v", err)
		http.Error(wr, "Internal Server Error", http.StatusInternalServerError)
	}
}

// HandlerResetNumRequests POST /admin/reset
func (cfg *APIConfig) HandlerResetNumRequests(wr http.ResponseWriter, req *http.Request) {
	if cfg.Platform != "dev" {
		log.Printf("access denied, platform is not dev")
		respondWithError(wr, fmt.Errorf("platform is not dev"), http.StatusForbidden)
	}
	if err := cfg.DBQueries.DeleteUsers(req.Context()); err != nil {
		log.Printf("error deleting users from database: %v", err)
		respondWithError(wr, err, http.StatusInternalServerError)
	}
	cfg.FileServerHits.Store(0)
	cfg.HandlerNumRequests(wr, req)
}

// HandlerReadiness GET /api/healthz
func (cfg *APIConfig) HandlerReadiness(wr http.ResponseWriter, req *http.Request) {
	wr.Header().Set("Content-Type", "text/plain; charset=utf-8")
	wr.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintf(wr, "OK\n")
	if err != nil {
		log.Printf("Error writing response body: %v", err)
	}
}

// HandlerRefresh POST /api/refresh
func (cfg *APIConfig) HandlerRefresh(wr http.ResponseWriter, req *http.Request) {
	tokenString, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("error pulling refresh token from authorization header: %v", err)
		respondWithError(wr, err, http.StatusUnauthorized)
	}

	dbRefreshToken, err := cfg.DBQueries.GetRefreshToken(req.Context(), tokenString)
	if err != nil {
		log.Printf("error getting refresh token from database: %v", err)
		respondWithError(wr, err, http.StatusUnauthorized)
	}

	accessToken, err := auth.MakeJWT(dbRefreshToken.UserID, cfg.Secret)
	if err != nil {
		log.Printf("error making JWT token: %v", err)
		respondWithError(wr, err, http.StatusInternalServerError)
	}

	apiAccessToken := NewAPIToken(accessToken)

	respondWithJSON(wr, apiAccessToken, http.StatusOK)
}

// HandlerRevoke POST /api/refresh
func (cfg *APIConfig) HandlerRevoke(wr http.ResponseWriter, req *http.Request) {
	tokenString, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("error pulling refresh token from authorization header: %v", err)
		respondWithError(wr, err, http.StatusUnauthorized)
	}

	revokedAt, err := cfg.DBQueries.RevokeRefreshToken(req.Context(), tokenString)
	if err != nil {
		log.Printf("error revoking refresh token: %v", err)
		respondWithError(wr, err, http.StatusUnauthorized)
	}
	log.Printf("token %s revoked at %v", tokenString, revokedAt)

	respondWithJSON(wr, struct{}{}, http.StatusNoContent)
}

func respondWithError(wr http.ResponseWriter, err error, code int) {
	log.Printf("%d error: %v\n", code, err)

	payload := struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	}
	respondWithJSON(wr, payload, code)
}

func respondWithJSON(wr http.ResponseWriter, payload any, code int) {
	if code == http.StatusNoContent || code == http.StatusNotModified || (100 <= code && code < 200) {
		wr.WriteHeader(code)
		return
	}

	wr.Header().Set("Content-Type", "application/json")

	b, err := json.Marshal(payload)
	if err != nil {
		log.Printf("error marshalling JSON: %v\n", err)
		wr.WriteHeader(http.StatusInternalServerError)
		return
	}

	wr.WriteHeader(code)
	if _, err = wr.Write(b); err != nil {
		log.Printf("error writing JSON body: %v\n", err)
		return
	}
}
