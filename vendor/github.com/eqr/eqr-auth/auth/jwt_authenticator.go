package auth

import (
	"fmt"
	"github.com/eqr/eqr-auth/config"
	"github.com/golang-jwt/jwt"
	"time"
)

type JWTService interface {
	GenerateToken(email string, userId uint64) (string, error)
	ValidateToken(token string) (*jwt.Token, error)
}

type authCustomClaims struct {
	Name string `json:"name"`
	Id   uint64 `json:"userId"`
	jwt.StandardClaims
}

type jwtServices struct {
	secretKey string
	issuer    string
}

func JWTAuthService(cfg *config.Config) JWTService {
	return &jwtServices{
		secretKey: cfg.JWT.Secret,
		issuer:    "eqr",
	}
}

func (service jwtServices) GenerateToken(email string, userId uint64) (string, error) {
	claims := &authCustomClaims{
		email,
		userId,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 48).Unix(),
			Issuer:    service.issuer,
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(service.secretKey))
	if err != nil {
		return "", err
	}
	return t, nil
}

func (service jwtServices) ValidateToken(encodedToken string) (*jwt.Token, error) {
	return jwt.Parse(encodedToken, func(token *jwt.Token) (interface{}, error) {
		if _, isValid := token.Method.(*jwt.SigningMethodHMAC); !isValid {
			return nil, fmt.Errorf("Invalid token: %v", token.Header["alg"])
		}
		return []byte(service.secretKey), nil
	})
}
