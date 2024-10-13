package userservice

import (
	"diet-app-backend/database/connection"
	"diet-app-backend/database/models"
	"diet-app-backend/schemas"
	"diet-app-backend/util/hashing"
	"diet-app-backend/util/tokens"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	var credentials schemas.Credentials

	if error := c.BindJSON(&credentials); error != nil {
		fmt.Println(error.Error())
		// TODO return error message
		return
	}

	var user models.User
	result := connection.Db.First(&user, "email = ?", credentials.Email)

	if result.Error != nil {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "Invalid credentials"})
		return
	}

	if isValid := hashing.CheckPasswordHash(credentials.Password, user.Password); !isValid {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error": "Invalid credentials"})
		return
	}

	tokenString, error := user.IssueToken()

	if error != nil {
		fmt.Println(error)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "It was not possible to issue a token"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"token": tokenString})
}

func Signup(c *gin.Context) {
	var user models.User

	if error := c.BindJSON(&user); error != nil {
		fmt.Println(error.Error())
		// TODO return error message
		return
	}

	hashed_password, _ := hashing.HashPassword(user.Password)

	user = models.User{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Password:  hashed_password,
	}

	result := connection.Db.Create(&user)

	if error := result.Error; error != nil {
		if strings.Contains(error.Error(), "Duplicate entry") {
			c.IndentedJSON(http.StatusConflict, gin.H{"error": "This email is not available"})
			return
		}
	}
	// Omitting password from the output
	user.Password = ""
	c.IndentedJSON(http.StatusCreated, user)
}

func GetUser(c *gin.Context) {
	claims, _ := tokens.GetClaims(c)

	id := claims["id"]

	var user models.User
	connection.Db.First(&user, id)
	// Omitting password from the output
	user.Password = ""
	c.IndentedJSON(http.StatusOK, user)
}
