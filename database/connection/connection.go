package connection

import (
	"diet-app-backend/database/models"
	"diet-app-backend/util/config"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Db *gorm.DB

func Connect() {
	var err error

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.AppConfig.DbUsername,
		config.AppConfig.DbPassword,
		config.AppConfig.DbHost,
		config.AppConfig.DbPort,
		config.AppConfig.DbDatabase,
	)

	Db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	Db.AutoMigrate(&models.User{})
	Db.AutoMigrate(&models.Food{})
	Db.AutoMigrate(&models.FoodItem{})
}
