package main

import (
	"diet-app-backend/api/routes"
	"diet-app-backend/database/connection"
)

func main() {
	connection.Connect()
	routes.Route()
}
