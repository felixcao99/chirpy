package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/felixcao99/chirpy/internal/database"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	Hashedpassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(Hashedpassword), err
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {
	currentUTC := jwt.NewNumericDate(time.Now().UTC())
	expireTime := jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(3600) * time.Second))

	claims := &jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  currentUTC,
		ExpiresAt: expireTime,
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(tokenSecret))
	return ss, err
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return uuid.Nil, err
		}
		return userID, nil
	} else {
		return uuid.Nil, err
	}
}

func GetBearerToken(headers http.Header) (string, error) {
	bearertoken := headers.Get("Authorization")
	if len(bearertoken) == 0 {
		err := errors.New("not authorrized")
		return "", err
	}
	token, found := strings.CutPrefix(bearertoken, "Bearer")
	if found {
		token = strings.TrimSpace(token)
		return token, nil
	} else {
		err := errors.New("not authorrized")
		return "", err
	}
}

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

func ValidateFreshToken(refreshtoken database.Refreshtoken) (uuid.UUID, error) {
	if refreshtoken.RevokedAt.Valid {
		return uuid.UUID{}, errors.New("refresh token revoked")
	}
	if time.Now().After(refreshtoken.ExpiresAt) {
		return uuid.UUID{}, errors.New("refresh token expired")
	}
	return refreshtoken.UserID, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	apikey := headers.Get("Authorization")
	if len(apikey) == 0 {
		err := errors.New("not authorrized")
		return "", err
	}

	key, found := strings.CutPrefix(apikey, "ApiKey")
	if found {
		key = strings.TrimSpace(key)
		return key, nil
	} else {
		err := errors.New("not authorrized")
		return "", err
	}
}
