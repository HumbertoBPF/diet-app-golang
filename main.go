package main

import (
	"diet-app-backend/api/routes"
	"diet-app-backend/database/connection"
	"diet-app-backend/util/config"
	"fmt"

	"gorm.io/driver/mysql"
)

func main() {
	config.LoadEnv(".")

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.AppConfig.DbUsername,
		config.AppConfig.DbPassword,
		config.AppConfig.DbHost,
		config.AppConfig.DbPort,
		config.AppConfig.DbDatabase,
	)
	dialector := mysql.Open(dsn)

	connection.Connect(dialector)
	routes.Route()
}
