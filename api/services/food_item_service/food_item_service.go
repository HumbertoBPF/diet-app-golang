package fooditemservice

import (
	"diet-app-backend/database/connection"
	"diet-app-backend/database/models"
	"diet-app-backend/schemas"
	"diet-app-backend/util/tokens"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

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

	result := connection.Db.Create(&foodItem)

	if result.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": "A food item entry could not be created",
		})
		return
	}

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

	updateResult := connection.Db.Save(&foodItem)

	if updateResult.Error != nil {
		fmt.Println("updateResult.Error", updateResult.Error)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update record",
		})
		return
	}

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
