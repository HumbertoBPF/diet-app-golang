package tokens

import (
	"diet-app-backend/util/config"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func GetClaims(c *gin.Context) (jwt.MapClaims, error) {
	authorizationHeader := c.GetHeader("Authorization")

	splitAuthorizationHeader := strings.Split(authorizationHeader, " ")

	if len(splitAuthorizationHeader) != 2 {
		return make(jwt.MapClaims), fmt.Errorf("incorrectly formatted authorization header")
	}

	return validate(splitAuthorizationHeader[1])
}

func validate(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return jwt.ParseRSAPublicKeyFromPEM([]byte(fmt.Sprintf(
			"-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----",
			config.AppConfig.JwtPublicKey,
		)))
	})

	if err != nil {
		return make(jwt.MapClaims), err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}

	return make(jwt.MapClaims), nil
}
