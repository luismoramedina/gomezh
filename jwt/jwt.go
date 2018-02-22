package jwt

import (
	"strings"
	"github.com/dgrijalva/jwt-go"
	"log"
	"crypto/rsa"
)

type JwtValidator struct {
	PublicKey *rsa.PublicKey
}

func (j JwtValidator) IsValidCredential(authorization string) (bool, map[string]interface{}) {
	if ((len(authorization) == 0)) {
		log.Println("no authorization header")
		return false, nil
	}

	authorization = strings.Replace(authorization, "Bearer ", "", 1)
	token, err := jwt.Parse(authorization, func(token *jwt.Token) (interface{}, error) {
		return j.PublicKey, nil
	})

	if err != nil {
		log.Println(err)
		return false, nil
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		log.Printf("user: %s", claims["sub"])
		return true, map[string]interface{}(claims)
	}
	return false, nil
}