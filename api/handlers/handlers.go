package handlers

import (
	"diet-app-backend/database/connection"
	"diet-app-backend/database/models"
	"diet-app-backend/schemas"
	"diet-app-backend/util/hashing"
	"diet-app-backend/util/tokens"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

func GetFoods(c *gin.Context) {
	name := c.DefaultQuery("name", "")

	var foods []models.Food

	connection.Db.Select(
		[]string{"id", "name", "calories", "portion"},
	).Where(
		"name LIKE ?", fmt.Sprintf("%%%s%%", name),
	).Find(&foods)

	c.IndentedJSON(http.StatusOK, foods)
}

func GetFood(c *gin.Context) {
	id := c.Param("id")

	var food models.Food

	err := connection.Db.First(&food, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.IndentedJSON(http.StatusNotFound, gin.H{
			"error": "Not Found",
		})
		return
	}

	c.IndentedJSON(http.StatusNotFound, food)
}

func GetUserFoods(c *gin.Context) {
	claims, _ := tokens.GetClaims(c)

	now := time.Now()
	defaultDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	timestamp := c.DefaultQuery("timestamp", defaultDate.Format(time.RFC3339))

	timestampDayAfter, err := time.Parse(time.RFC3339, timestamp)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": "The timestamp query string is formatted badly",
		})
		return
	}

	timestampDayAfter = timestampDayAfter.AddDate(0, 0, 1)

	userId := claims["id"]

	var foodItems []models.FoodItem
	connection.Db.Where("user_id = ? AND timestamp BETWEEN ? AND ?", userId, timestamp, timestampDayAfter).Find(&foodItems)

	c.IndentedJSON(http.StatusOK, foodItems)
}

func GetUserFood(c *gin.Context) {
	claims, _ := tokens.GetClaims(c)

	id := c.Param("id")
	userId := claims["id"]

	var foodItem models.FoodItem
	result := connection.Db.Where("id = ? AND user_id = ?", id, userId).First(&foodItem)

	if result.Error != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{
			"error": "Not found",
		})
		return
	}

	c.IndentedJSON(http.StatusOK, foodItem)
}

func PostUserFood(c *gin.Context) {
	claims, _ := tokens.GetClaims(c)

	id := claims["id"]

	var foodItem models.FoodItem

	if err := c.BindJSON(&foodItem); err != nil {
		return
	}

	foodItem.UserID = uint(id.(float64))

	connection.Db.Create(&foodItem)
	c.IndentedJSON(http.StatusCreated, foodItem)
}

func PutUserFood(c *gin.Context) {
	claims, _ := tokens.GetClaims(c)

	id := c.Param("id")
	userId := claims["id"]

	var updateFoodItem schemas.UpdateFoodItem

	if err := c.BindJSON(&updateFoodItem); err != nil {
		fmt.Println(err)
		return
	}

	var foodItem models.FoodItem
	result := connection.Db.Where("id = ? AND user_id = ?", id, userId).First(&foodItem)

	if result.Error != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{
			"error": "Not found",
		})
		return
	}

	foodItem.Quantity = updateFoodItem.Quantity
	foodItem.Timestamp = updateFoodItem.Timestamp

	connection.Db.Save(&foodItem)
	c.IndentedJSON(http.StatusOK, foodItem)
}

func DeleteUserFood(c *gin.Context) {
	claims, _ := tokens.GetClaims(c)

	id := c.Param("id")
	userId := claims["id"]

	var foodItem models.FoodItem
	result := connection.Db.Where("id = ? AND user_id = ?", id, userId).First(&foodItem)

	if result.Error != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{
			"error": "Not found",
		})
		return
	}

	connection.Db.Delete(&foodItem)
	c.IndentedJSON(http.StatusNoContent, nil)
}
