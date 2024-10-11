package models

import (
	"diet-app-backend/util/config"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID        uint       `json:"id" gorm:"primarykey"`
	Email     string     `json:"email" binding:"required" gorm:"not null;unique"`
	FirstName string     `json:"first_name" binding:"required" gorm:"not null"`
	LastName  string     `json:"last_name" binding:"required" gorm:"not null"`
	Password  string     `json:"password,omitempty" binding:"required" gorm:"not null"`
	FoodItems []FoodItem `json:"-"`
}

func (user User) IssueToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"id": user.ID,
	})

	key, error := jwt.ParseRSAPrivateKeyFromPEM([]byte(fmt.Sprintf(
		"-----BEGIN RSA PRIVATE KEY-----\n%s\n-----END RSA PRIVATE KEY-----",
		config.AppConfig.JwtPrivateKey,
	)))

	if error != nil {
		fmt.Println(error)
		return "", error
	}

	return token.SignedString(key)
}

type Food struct {
	ID        uint       `json:"id" gorm:"primarykey"`
	Name      string     `json:"name" binding:"required" gorm:"unique;not null"`
	Calories  int        `json:"calories" binding:"required" gorm:"not null"`
	Portion   int        `json:"portion" binding:"required" gorm:"not null"`
	FoodItems []FoodItem `json:"-"`
}

type FoodItem struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	FoodID    uint      `json:"food_id" binding:"required" gorm:"not null"`
	Quantity  uint      `json:"quantity" binding:"required" gorm:"not null"`
	Timestamp time.Time `json:"timestamp" binding:"required" gorm:"not null"`
}
