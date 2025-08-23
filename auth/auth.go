// Package auth holds authorization
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword uses bcrypt to hash provided password
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("error hashing password with bcrypt: %v\n", err)
		return "", nil
	}
	return string(hashedBytes), nil
}

// CheckPasswordHash checks the provided password agains the hashed password using bcrypt
func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// MakeJWT creates and returns JWT
func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {
	now := time.Now()

	registeredClaims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(1) * time.Hour)),
		Subject:   userID.String(),
		ID:        userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, registeredClaims)

	signedString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		log.Printf("error signing jwt token: %v\n", err)
		return "", fmt.Errorf("error signing jwt token: %w", err)
	}

	return signedString, nil
}

// ValidateJWT validates a token string agains the secret
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	keyFunc := func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("the token does not have signing method HMAC: %v\n", t.Header["alg"])
			return nil, fmt.Errorf("incorrect signing method: %v", t.Header["alg"])
		}
		return []byte(tokenSecret), nil
	}
	fmt.Printf("\nbefore parse with claims \n\ntokenString: %s\n\n", tokenString)
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil || !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid token: %w", err)
	}
	fmt.Println("after parse with claims")

	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		log.Printf("error parsing id from string: %v\n", err)
		return uuid.Nil, fmt.Errorf("invalid subject in token: %w", err)
	}

	return id, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		log.Printf("authorization header is empty")
		return "", fmt.Errorf("authorization header is empty")
	}

	if !strings.HasPrefix(authHeader, "Bearer") {
		log.Printf("authorization header does not have prefix 'Bearer '")
		return "", fmt.Errorf("authorization header does not have prefix 'Bearer '")
	}

	tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))
	if tokenString == "" {
		log.Printf("token string is empty")
		return "", fmt.Errorf("error: token string is empty")
	}

	return tokenString, nil
}

func MakeRefreshToken() (string, error) {
	b := make([]byte, 256)
	if _, err := rand.Read(b); err != nil {
		log.Printf("error making refresh token: %v", err)
		return "", fmt.Errorf("error making refresh token: %w", err)
	}

	return hex.EncodeToString(b), nil
}
