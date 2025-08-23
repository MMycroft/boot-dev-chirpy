package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	password := "mySecret123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if hash == "" {
		t.Fatal("expected non-empty hash, got empty string")
	}

	if hash == password {
		t.Fatal("hash should not equal the raw password")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "Correct password",
			password: password1,
			hash:     hash1,
			wantErr:  false,
		},
		{
			name:     "Incorrect password",
			password: "wrongPassword",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Password doesn't match different hash",
			password: password1,
			hash:     hash2,
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Invalid hash",
			password: password1,
			hash:     "invalidhash",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	secret := "supersecret"
	expiresIn := time.Minute

	// Call the function under test
	tokenStr, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tokenStr == "" {
		t.Fatal("expected a non-empty token string")
	}

	// Parse the token back to validate it
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		// ensure signing method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			t.Fatalf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	if !token.Valid {
		t.Fatal("expected token to be valid, got invalid")
	}

	// Check claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("expected MapClaims")
	}

	if claims["sub"] != userID.String() {
		t.Errorf("expected sub claim %q, got %q", userID.String(), claims["sub"])
	}

	if claims["jti"] != userID.String() {
		t.Errorf("expected jti claim %q, got %q", userID.String(), claims["jti"])
	}

	exp, ok := claims["exp"].(float64) // exp is usually float64 (unix timestamp)
	if !ok {
		t.Fatal("expected exp claim to be a number")
	}
	if int64(exp) <= time.Now().Unix() {
		t.Errorf("expected exp in the future, got %v", exp)
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	validToken, _ := MakeJWT(userID, "secret", time.Hour)

	tests := []struct {
		name        string
		tokenString string
		tokenSecret string
		wantUserID  uuid.UUID
		wantErr     bool
	}{
		{
			name:        "Valid token",
			tokenString: validToken,
			tokenSecret: "secret",
			wantUserID:  userID,
			wantErr:     false,
		},
		{
			name:        "Invalid token",
			tokenString: "invalid.token.string",
			tokenSecret: "secret",
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
		{
			name:        "Wrong secret",
			tokenString: validToken,
			tokenSecret: "wrong_secret",
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserID, err := ValidateJWT(tt.tokenString, tt.tokenSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUserID != tt.wantUserID {
				t.Errorf("ValidateJWT() gotUserID = %v, want %v", gotUserID, tt.wantUserID)
			}
		})
	}
}
