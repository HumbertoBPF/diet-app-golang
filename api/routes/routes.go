package routes

import (
	"diet-app-backend/api/handlers"
	"diet-app-backend/util/authentication"

	"github.com/gin-gonic/gin"
)

func Route() {
	router := gin.Default()

	router.POST("/login", handlers.Login)
	router.POST("/signup", handlers.Signup)

	router.GET("/user", authentication.Authenticate(handlers.GetUser))

	router.GET("/food", handlers.GetFoods)
	router.GET("/food/:id", handlers.GetFood)

	router.GET("/user/food", authentication.Authenticate(handlers.GetUserFoods))
	router.GET("/user/food/:id", authentication.Authenticate(handlers.GetUserFood))
	router.POST("/user/food", authentication.Authenticate(handlers.PostUserFood))
	router.PUT("/user/food/:id", authentication.Authenticate(handlers.PutUserFood))
	router.DELETE("/user/food/:id", authentication.Authenticate(handlers.DeleteUserFood))

	router.Run("localhost:8080")
}
