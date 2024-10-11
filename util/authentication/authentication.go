package authentication

import (
	"diet-app-backend/database/connection"
	"diet-app-backend/database/models"
	"diet-app-backend/util/tokens"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Authenticate(handler func(c *gin.Context)) func(c *gin.Context) {
	return func(c *gin.Context) {
		claims, err := tokens.GetClaims(c)

		if err != nil {
			fmt.Println(err)
			c.IndentedJSON(http.StatusForbidden, gin.H{
				"error": "Authentication failed",
			})
			return
		}

		id, ok := claims["id"]

		if !ok {
			c.IndentedJSON(http.StatusForbidden, gin.H{
				"error": "Authentication failed",
			})
			return
		}

		var user models.User

		if err := connection.Db.First(&user, id).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Println(err)
			c.IndentedJSON(http.StatusForbidden, gin.H{
				"error": "Authentication failed",
			})
			return
		}

		handler(c)
	}
}
