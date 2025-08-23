package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/mmycroft/boot-dev-chirpy/auth"
	"github.com/mmycroft/boot-dev-chirpy/database"
)

// HandlerCreateChirp POST /api/chirps
func (cfg *APIConfig) HandlerCreateChirp(wr http.ResponseWriter, req *http.Request) {
	const _MAX_CHIRP_LENGTH = 140

	reqBody := struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}{}

	fmt.Printf("Empty reqBody: %s\n\n", reqBody)

	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		log.Printf("error decoding request body: %v\n", err)
		respondWithError(wr, err, http.StatusInternalServerError)
		return
	}

	fmt.Printf("Filled reqBody: %s\n\n", reqBody)

	tokenString, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("error getting token from header: %v\n", err)
		respondWithError(wr, err, http.StatusUnauthorized)
		return
	}

	fmt.Printf("tokenString: %s\n\n", tokenString)

	userID, err := auth.ValidateJWT(tokenString, cfg.Secret)
	if err != nil {
		log.Printf("error validating JWT token: %v\n", err)
		respondWithError(wr, err, http.StatusUnauthorized)
		return
	}

	if len(reqBody.Body) > _MAX_CHIRP_LENGTH {
		log.Printf("chirp must be 140 characters or less\n")
		respondWithError(wr, fmt.Errorf("chirp is too long"), http.StatusBadRequest)
		return
	}

	badWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	words := strings.Split(reqBody.Body, " ")
	for i, word := range words {
		if badWords[strings.ToLower(word)] {
			words[i] = "****"
		}
	}

	chirpParams := database.CreateChirpParams{
		Body:   strings.Join(words, " "),
		UserID: userID,
	}

	dbChirp, err := cfg.DBQueries.CreateChirp(req.Context(), chirpParams)
	if err != nil {
		log.Printf("error creating database chirp: %v\n", err)
		respondWithError(wr, err, http.StatusInternalServerError)
		return
	}

	apiChirp := NewAPIChirp(&dbChirp)

	respondWithJSON(wr, apiChirp, http.StatusCreated)
}

// HandlerGetChirps GET /api/chirps
func (cfg *APIConfig) HandlerGetChirps(wr http.ResponseWriter, req *http.Request) {
	dbChirps, err := cfg.DBQueries.GetChirps(req.Context())
	if err != nil {
		log.Printf("error retrieving chirps from database: %v\n", err)
		respondWithError(wr, err, http.StatusNotFound)
		return
	}

	apiChirps := make([]APIChirp, len(dbChirps))
	for i, dbChirp := range dbChirps {
		apiChirps[i] = NewAPIChirp(&dbChirp)
	}

	respondWithJSON(wr, apiChirps, http.StatusOK)
}

// HandlerGetChirp GET /api/chirps{chirpID}
func (cfg *APIConfig) HandlerGetChirp(wr http.ResponseWriter, req *http.Request) {
	chirpID, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		log.Printf("error parsing path {chirpID}: %v", err)
		respondWithError(wr, err, http.StatusInternalServerError)
		return
	}

	dbChirp, err := cfg.DBQueries.GetChirp(req.Context(), chirpID)
	if err != nil {
		log.Printf("error getting user from database: %v", err)
		respondWithError(wr, err, http.StatusNotFound)
		return
	}

	apiChirp := NewAPIChirp(&dbChirp)

	respondWithJSON(wr, apiChirp, http.StatusOK)
}

// HandlerDeleteChirp DELETE /api/chirps/{chirpID}
func (cfg *APIConfig) HandlerDeleteChirp(wr http.ResponseWriter, req *http.Request) {
	accessToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("error pulling access token from authorization header: %v\n", err)
		respondWithError(wr, err, http.StatusUnauthorized)
		return
	}

	reqUserID, err := auth.ValidateJWT(accessToken, cfg.Secret)
	if err != nil {
		log.Printf("error validating access token: %v\n", err)
		respondWithError(wr, err, http.StatusUnauthorized)
		return
	}

	chirpID, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		log.Printf("error parsing path {chirpID}: %v\n", err)
		respondWithError(wr, err, http.StatusNotFound)
		return
	}

	dbChirp, err := cfg.DBQueries.GetChirp(req.Context(), chirpID)
	if err != nil {
		log.Printf("error getting user from database: %v\n", err)
		respondWithError(wr, err, http.StatusNotFound)
		return
	}

	if reqUserID != dbChirp.UserID {
		err := fmt.Errorf("request user id does not match chirp user id")
		log.Println(err)
		respondWithError(wr, err, http.StatusForbidden)
		return
	}

	if err = cfg.DBQueries.DeleteChirp(req.Context(), chirpID); err != nil {
		log.Printf("error deleting chirp from database: %v\n", err)
		respondWithError(wr, err, http.StatusInternalServerError)
		return
	}

	respondWithJSON(wr, struct{}{}, http.StatusNoContent)
}
