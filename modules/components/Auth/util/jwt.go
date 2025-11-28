package utitjwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var JwtSecret = []byte("3e42babaa52457013fbb78275b17d7a963ec6a8dde9c24dc1beeaaf149ca021f146cd18fe3f630db97e0498ef33bbe3fe4578b291ce38ffba1478a731fa9d99b")

func GenerateToken(userID uuid.UUID, username string, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  userID.String(),
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	})
	return token.SignedString(JwtSecret)
}

func ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return JwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return nil, jwt.ErrTokenExpired
			}
		}
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
