package foodservice

import (
	"diet-app-backend/database/connection"
	"diet-app-backend/database/models"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

	c.IndentedJSON(http.StatusOK, food)
}
