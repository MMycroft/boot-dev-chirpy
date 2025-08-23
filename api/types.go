package api

import (
	"time"

	"github.com/google/uuid"

	"github.com/mmycroft/boot-dev-chirpy/database"
)

type APIChirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

type APIToken struct {
	Token string `json:"token"`
}

type APIUser struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func NewAPIUser(dbUser *database.User, token, refreshToken string) APIUser {
	return APIUser{
		ID:           dbUser.ID,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		Email:        dbUser.Email,
		Token:        token,
		RefreshToken: refreshToken,
	}
}

func NewAPIToken(token string) APIToken {
	return APIToken{
		Token: token,
	}
}

func NewAPIChirp(dbChirp *database.Chirp) APIChirp {
	return APIChirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}
}
