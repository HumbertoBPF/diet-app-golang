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
