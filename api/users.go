package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mmycroft/boot-dev-chirpy/auth"

	"github.com/mmycroft/boot-dev-chirpy/database"

	"github.com/google/uuid"
)

// HandlerCreateUser POST /api/users
func (cfg *APIConfig) HandlerCreateUser(wr http.ResponseWriter, req *http.Request) {
	userData := struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		IsChirpyRed bool   `json:"is_chirpy_red"`
	}{}

	if err := json.NewDecoder(req.Body).Decode(&userData); err != nil {
		log.Printf("error decoding request body: %v\n", err)
		respondWithError(wr, err, http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(userData.Password)
	if err != nil {
		log.Printf("error hashing password: %v\n", err)
		respondWithError(wr, err, http.StatusBadRequest)
		return
	}

	userParams := database.CreateUserParams{
		Email:          userData.Email,
		HashedPassword: hashedPassword,
		IsChirpyRed:    userData.IsChirpyRed,
	}

	dbUser, err := cfg.DBQueries.CreateUser(req.Context(), userParams)
	if err != nil {
		log.Printf("error creating database user: %v\n", err)
		respondWithError(wr, err, http.StatusInternalServerError)
		return
	}

	apiUser := NewAPIUser(&dbUser, "", "")

	respondWithJSON(wr, apiUser, http.StatusCreated)
}

// HandlerGetUsers GET /api/users
func (cfg *APIConfig) HandlerGetUsers(wr http.ResponseWriter, req *http.Request) {
	dbUsers, err := cfg.DBQueries.GetUsers(req.Context())
	if err != nil {
		log.Printf("error retrieving chirps from database: %v\n", err)
		respondWithError(wr, err, http.StatusNotFound)
		return
	}
	apiUsers := make([]APIUser, len(dbUsers))
	for i, dbUser := range dbUsers {
		apiUsers[i] = NewAPIUser(&dbUser, "", "")
	}

	respondWithJSON(wr, apiUsers, http.StatusOK)
}

// HandlerGetUser GET /api/users/{userID}
func (cfg *APIConfig) HandlerGetUser(wr http.ResponseWriter, req *http.Request) {
	userID, err := uuid.Parse(req.PathValue("userID"))
	if err != nil {
		log.Printf("error parsing path {userID}: %v\n", err)
		respondWithError(wr, err, http.StatusInternalServerError)
		return
	}

	dbUser, err := cfg.DBQueries.GetUserByID(req.Context(), userID)
	if err != nil {
		log.Printf("error getting user from database: %v\n", err)
		respondWithError(wr, err, http.StatusNotFound)
		return
	}

	apiUser := NewAPIUser(&dbUser, "", "")

	respondWithJSON(wr, apiUser, http.StatusOK)
}

// HandlerUpdateUser PUT /api/users
func (cfg *APIConfig) HandlerUpdateUser(wr http.ResponseWriter, req *http.Request) {
	accessToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("error pulling access token from authorization header: %v\n", err)
		respondWithError(wr, err, http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.Secret)
	if err != nil {
		log.Printf("error validating access token: %v\n", err)
		respondWithError(wr, err, http.StatusUnauthorized)
		return
	}

	userData := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	if err := json.NewDecoder(req.Body).Decode(&userData); err != nil {
		log.Printf("error decoding request body: %v\n", err)
		respondWithError(wr, err, http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(userData.Password)
	if err != nil {
		log.Printf("error hashing password: %v\n", err)
		respondWithError(wr, err, http.StatusBadRequest)
		return
	}

	userParams := database.UpdateUserParams{
		ID:             userID,
		Email:          userData.Email,
		HashedPassword: hashedPassword,
	}

	dbUser, err := cfg.DBQueries.UpdateUser(req.Context(), userParams)
	if err != nil {
		log.Printf("error updating database user: %v\n", err)
		respondWithError(wr, err, http.StatusInternalServerError)
		return
	}

	apiUser := NewAPIUser(&dbUser, "", "")

	respondWithJSON(wr, apiUser, http.StatusOK)
}

// HandlerUpgradeUser POST /api/polka/webhooks
func (cfg *APIConfig) HandlerUpgradeUser(wr http.ResponseWriter, req *http.Request) {
	eventData := struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}{}

	if err := json.NewDecoder(req.Body).Decode(&eventData); err != nil {
		log.Printf("error decoding request body: %v\n", err)
		respondWithError(wr, err, http.StatusBadRequest)
		return
	}

	if eventData.Event != "user.upgraded" {
		log.Printf("incorrect event")
		respondWithJSON(wr, struct{}{}, http.StatusNoContent)
		return
	}

	dbUser, err := cfg.DBQueries.UpgradeUser(req.Context(), true)
	if err != nil {
		log.Printf("error upgrading user %v: %v", dbUser, err)
		respondWithError(wr, err, http.StatusNotFound)
		return
	}
	respondWithJSON(wr, struct{}{}, http.StatusNoContent)
	return
}

// HandlerLogin POST /api/login
func (cfg *APIConfig) HandlerLogin(wr http.ResponseWriter, req *http.Request) {
	userData := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	if err := json.NewDecoder(req.Body).Decode(&userData); err != nil {
		log.Printf("error decoding request body: %v\n", err)
		respondWithError(wr, err, http.StatusBadRequest)
		return
	}

	dbUser, err := cfg.DBQueries.GetUserByEmail(req.Context(), userData.Email)
	if err != nil {
		log.Printf("error getting user from database: %\n", err)
		respondWithError(wr, err, http.StatusUnauthorized)
		return
	}

	if err = auth.CheckPasswordHash(userData.Password, dbUser.HashedPassword); err != nil {
		log.Printf("error comparing hashed password: %v\n", err)
		respondWithError(wr, err, http.StatusUnauthorized)
		return
	}

	accessToken, err := auth.MakeJWT(dbUser.ID, cfg.Secret)
	if err != nil {
		log.Printf("error making JWT token: %v\n", err)
		respondWithError(wr, err, http.StatusInternalServerError)
		return
	}

	refreshTokenString, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("error making Refresh Token %v\n", err)
		respondWithError(wr, err, http.StatusInternalServerError)
		return
	}

	refreshTokenParams := database.CreateRefreshTokenParams{
		Token:  refreshTokenString,
		UserID: dbUser.ID,
	}

	refreshToken, err := cfg.DBQueries.CreateRefreshToken(req.Context(), refreshTokenParams)
	if err != nil {
		log.Printf("error creating refresh token: %v\n", err)
		respondWithError(wr, err, http.StatusInternalServerError)
		return
	}

	apiUser := NewAPIUser(&dbUser, accessToken, refreshToken.Token)

	respondWithJSON(wr, apiUser, http.StatusOK)
}
