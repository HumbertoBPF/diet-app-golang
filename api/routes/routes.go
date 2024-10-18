package routes

import (
	fooditemservice "diet-app-backend/api/services/food_item_service"
	foodservice "diet-app-backend/api/services/food_service"
	userservice "diet-app-backend/api/services/user_service"
	"diet-app-backend/util/authentication"
	"diet-app-backend/util/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{config.AppConfig.FrontEndUrl}
	corsConfig.AddAllowHeaders("Authorization")

	router.Use(cors.New(corsConfig))

	router.POST("login", userservice.Login)
	router.POST("signup", userservice.Signup)
	router.GET("user", authentication.Authenticate(userservice.GetUser))

	router.GET("food", foodservice.GetFoods)
	router.GET("food/:id", foodservice.GetFood)

	router.GET("user/food", authentication.Authenticate(fooditemservice.GetUserFoods))
	router.GET("user/food/:id", authentication.Authenticate(fooditemservice.GetUserFood))
	router.POST("user/food", authentication.Authenticate(fooditemservice.PostUserFood))
	router.PUT("user/food/:id", authentication.Authenticate(fooditemservice.PutUserFood))
	router.DELETE("user/food/:id", authentication.Authenticate(fooditemservice.DeleteUserFood))

	return router
}

func Route() {
	router := SetupRouter()

	router.Run("localhost:8080")
}
