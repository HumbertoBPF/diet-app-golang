package connection

import (
	"diet-app-backend/database/models"

	"gorm.io/gorm"
)

var Db *gorm.DB

func Connect(dialector gorm.Dialector) {
	var err error

	Db, err = gorm.Open(dialector, &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	Db.AutoMigrate(&models.User{})
	Db.AutoMigrate(&models.Food{})
	Db.AutoMigrate(&models.FoodItem{})
}
