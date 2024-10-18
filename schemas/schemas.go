package schemas

import (
	"time"
)

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateFoodItem struct {
	Quantity  uint      `json:"quantity" binding:"required"`
	Timestamp time.Time `json:"timestamp" binding:"required"`
}

type JoinedFoodItem struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	FoodID    uint      `json:"food_id"`
	Name      string    `json:"name"`
	Calories  int       `json:"calories"`
	Portion   int       `json:"portion"`
	Quantity  uint      `json:"quantity"`
	Timestamp time.Time `json:"timestamp"`
}
