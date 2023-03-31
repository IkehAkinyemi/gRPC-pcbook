package service

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// JWTManager is a JSON web token manager.
type JWTManager struct {
	privateKey    string
	publicKey     string
	tokenDuration time.Duration
}

// UserClaims is a custom JWT claims that contains some user's information.
type UserClaims struct {
	jwt.StandardClaims
	Username string `json:"username"`
	Role     string `json:"role"`
}

// NewJWTManager returns an instance of JWT manager.
func NewJWTManager(privateKey, publicKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{privateKey, publicKey, tokenDuration}
}

// GenerateToken creates and signs a new token for a user.
func (manager *JWTManager) GenerateToken(user *User) (string, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(manager.privateKey))
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %v", err)
	}

	claims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(manager.tokenDuration).Unix(),
		},
		Username: user.Username,
		Role:     user.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return signedToken, nil
}

// Verify verifies the given JWT token and returns the claims if the token is valid.
func (manager *JWTManager) Verify(accessToken string) (*UserClaims, error) {
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(manager.publicKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	token, err := jwt.ParseWithClaims(
		accessToken,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodRSA)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}

			return publicKey, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
