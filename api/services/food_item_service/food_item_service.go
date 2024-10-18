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

	timestampStr := c.DefaultQuery("timestamp", defaultDate.Format(time.RFC3339))
	timestamp, err := time.Parse(time.RFC3339, timestampStr)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": "The timestamp query string is formatted badly",
		})
		return
	}

	timestampDayAfter := timestamp.AddDate(0, 0, 1)

	userId := claims["id"]

	var foodItems []schemas.JoinedFoodItem
	connection.Db.Model(&models.FoodItem{}).
		Select("food_items.id, food_items.user_id, foods.id as food_id, foods.name, foods.calories, foods.portion, food_items.quantity, food_items.timestamp").
		Joins("JOIN foods ON food_items.food_id = foods.id").
		Where("user_id = ? AND timestamp BETWEEN ? AND ?", userId, timestamp, timestampDayAfter).
		Find(&foodItems)

	c.IndentedJSON(http.StatusOK, foodItems)
}

func GetUserFood(c *gin.Context) {
	claims, _ := tokens.GetClaims(c)

	id := c.Param("id")
	userId := claims["id"]

	var foodItem schemas.JoinedFoodItem
	result := connection.Db.Model(&models.FoodItem{}).
		Select("food_items.id, food_items.user_id, foods.id as food_id, foods.name, foods.calories, foods.portion, food_items.quantity, food_items.timestamp").
		Joins("JOIN foods ON food_items.food_id = foods.id").
		Where("food_items.id = ? AND food_items.user_id = ?", id, userId).
		First(&foodItem)

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

	userId := claims["id"]

	var foodItem models.FoodItem

	if err := c.BindJSON(&foodItem); err != nil {
		return
	}

	foodItem.UserID = uint(userId.(float64))

	result := connection.Db.Create(&foodItem)

	if result.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": "A food item entry could not be created",
		})
		return
	}

	var joinedFoodItem schemas.JoinedFoodItem

	connection.Db.Model(&models.FoodItem{}).
		Select("food_items.id, food_items.user_id, foods.id as food_id, foods.name, foods.calories, foods.portion, food_items.quantity, food_items.timestamp").
		Joins("JOIN foods ON food_items.food_id = foods.id").
		Where("food_items.id = ? AND food_items.user_id = ?", foodItem.ID, userId).
		First(&joinedFoodItem)

	c.IndentedJSON(http.StatusCreated, joinedFoodItem)
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

	var joinedFoodItem schemas.JoinedFoodItem

	connection.Db.Model(&models.FoodItem{}).
		Select("food_items.id, food_items.user_id, foods.id as food_id, foods.name, foods.calories, foods.portion, food_items.quantity, food_items.timestamp").
		Joins("JOIN foods ON food_items.food_id = foods.id").
		Where("food_items.id = ? AND food_items.user_id = ?", id, userId).
		First(&joinedFoodItem)

	c.IndentedJSON(http.StatusOK, joinedFoodItem)
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
