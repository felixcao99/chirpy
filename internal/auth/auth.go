package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

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

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	currentUTC := jwt.NewNumericDate(time.Now().UTC())
	expireTime := jwt.NewNumericDate(time.Now().UTC().Add(expiresIn))

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
