package fooditemservice_test

import (
	"database/sql"
	"diet-app-backend/api/routes"
	"diet-app-backend/database/connection"
	"diet-app-backend/database/models"
	"diet-app-backend/schemas"
	"diet-app-backend/util/config"
	"diet-app-backend/util/hashing"
	"diet-app-backend/util/tests"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
)

const email = "test.user@test.com"
const firstName = "Joe"
const lastName = "Doe"
const password = "Str0ng-P@ssw0rd"

var hashedPassword, _ = hashing.HashPassword(password)

type TestSuite struct {
	suite.Suite
	db   *sql.DB
	mock sqlmock.Sqlmock
}

func (suite *TestSuite) SetupTest() {
	db, mock, err := sqlmock.New()

	if err != nil {
		suite.T().Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	dialector := mysql.New(mysql.Config{
		DSN:                       "sqlmock_db_0",
		DriverName:                "mysql",
		Conn:                      db,
		SkipInitializeWithVersion: true,
	})
	connection.Connect(dialector)

	suite.db = db
	suite.mock = mock

	config.LoadEnv("../../../.")
}

func (suite *TestSuite) TearDownTest() {
	suite.db.Close()

	if err := suite.mock.ExpectationsWereMet(); err != nil {
		suite.T().Errorf("there were unfulfilled expectations: %s", err)
	}
}

func (suite *TestSuite) TestGetUserFoodsSuccessful() {
	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(email, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE `users`.`id` = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery(
		"^SELECT food_items.id, food_items.user_id, foods.id as food_id, foods.name, foods.calories, foods.portion, food_items.quantity, food_items.timestamp"+
			" FROM `food_items` "+
			"JOIN foods ON food_items.food_id = foods.id WHERE user_id = \\? AND timestamp BETWEEN \\? AND \\?",
	).
		WithArgs(float64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "user_id", "food_id", "name", "calories", "portion", "quantity", "timestamp"}).
				AddRow(1, 1, 1, "Pasta", 193, 80, 100, time.Now()).
				AddRow(1, 1, 2, "Grated Cheese", 492, 100, 150, time.Now()).
				AddRow(1, 1, 3, "Tomato Sauce", 34, 100, 50, time.Now()),
		)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	token := tests.GetToken(email, password)

	req, _ := http.NewRequest("GET", "/user/food", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, req)

	var responseBody []schemas.JoinedFoodItem
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 200, w.Code)

	assert.Len(suite.T(), responseBody, 3)

	assert.Equal(suite.T(), uint(1), responseBody[0].ID)
	assert.Equal(suite.T(), uint(1), responseBody[0].UserID)
	assert.Equal(suite.T(), uint(1), responseBody[0].FoodID)
	assert.Equal(suite.T(), "Pasta", responseBody[0].Name)
	assert.Equal(suite.T(), 193, responseBody[0].Calories)
	assert.Equal(suite.T(), 80, responseBody[0].Portion)
	assert.Equal(suite.T(), uint(100), responseBody[0].Quantity)

	assert.Equal(suite.T(), uint(1), responseBody[1].ID)
	assert.Equal(suite.T(), uint(1), responseBody[1].UserID)
	assert.Equal(suite.T(), uint(2), responseBody[1].FoodID)
	assert.Equal(suite.T(), "Grated Cheese", responseBody[1].Name)
	assert.Equal(suite.T(), 492, responseBody[1].Calories)
	assert.Equal(suite.T(), 100, responseBody[1].Portion)
	assert.Equal(suite.T(), uint(150), responseBody[1].Quantity)

	assert.Equal(suite.T(), uint(1), responseBody[2].ID)
	assert.Equal(suite.T(), uint(1), responseBody[2].UserID)
	assert.Equal(suite.T(), uint(3), responseBody[2].FoodID)
	assert.Equal(suite.T(), "Tomato Sauce", responseBody[2].Name)
	assert.Equal(suite.T(), 34, responseBody[2].Calories)
	assert.Equal(suite.T(), 100, responseBody[2].Portion)
	assert.Equal(suite.T(), uint(50), responseBody[2].Quantity)
}

func (suite *TestSuite) TestGetUserFoodsWithoutAuthorization() {
	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/user/food", nil)

	router.ServeHTTP(w, req)

	var responseBody tests.GenericErrorResponseBody
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 403, w.Code)
	assert.Equal(suite.T(), "Authentication failed", responseBody.Error)
}

func (suite *TestSuite) TestGetUserFoodsWithTimestampFilter() {
	timestamp := "2024-10-10T21:00:00.000Z"

	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(email, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE `users`.`id` = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery(
		"^SELECT food_items.id, food_items.user_id, foods.id as food_id, foods.name, foods.calories, foods.portion, food_items.quantity, food_items.timestamp"+
			" FROM `food_items` "+
			"JOIN foods ON food_items.food_id = foods.id WHERE user_id = \\? AND timestamp BETWEEN \\? AND \\?",
	).
		WithArgs(float64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "user_id", "food_id", "name", "calories", "portion", "quantity", "timestamp"}).
				AddRow(1, 1, 1, "Pasta", 193, 80, 100, time.Now()),
		)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	token := tests.GetToken(email, password)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/user/food?timestamp=%s", timestamp), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, req)

	var responseBody []schemas.JoinedFoodItem
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 200, w.Code)

	assert.Len(suite.T(), responseBody, 1)

	assert.Equal(suite.T(), uint(1), responseBody[0].ID)
	assert.Equal(suite.T(), uint(1), responseBody[0].UserID)
	assert.Equal(suite.T(), uint(1), responseBody[0].FoodID)
	assert.Equal(suite.T(), "Pasta", responseBody[0].Name)
	assert.Equal(suite.T(), 193, responseBody[0].Calories)
	assert.Equal(suite.T(), 80, responseBody[0].Portion)
	assert.Equal(suite.T(), uint(100), responseBody[0].Quantity)
}

func (suite *TestSuite) TestGetUserFoodSuccessful() {
	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(email, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE `users`.`id` = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery(
		"^SELECT food_items.id, food_items.user_id, foods.id as food_id, foods.name, foods.calories, foods.portion, food_items.quantity, food_items.timestamp"+
			" FROM `food_items` "+
			"JOIN foods ON food_items.food_id = foods.id WHERE food_items.id = \\? AND food_items.user_id = \\? ORDER BY `food_items`.`id` LIMIT \\?",
	).
		WithArgs("1", float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "user_id", "food_id", "name", "calories", "portion", "quantity", "timestamp"}).
				AddRow(1, 1, 1, "Pasta", 193, 80, 100, time.Now()),
		)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	token := tests.GetToken(email, password)

	req, _ := http.NewRequest("GET", "/user/food/1", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, req)

	var responseBody schemas.JoinedFoodItem
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 200, w.Code)

	assert.Equal(suite.T(), uint(1), responseBody.ID)
	assert.Equal(suite.T(), uint(1), responseBody.UserID)
	assert.Equal(suite.T(), uint(1), responseBody.FoodID)
	assert.Equal(suite.T(), "Pasta", responseBody.Name)
	assert.Equal(suite.T(), 193, responseBody.Calories)
	assert.Equal(suite.T(), 80, responseBody.Portion)
	assert.Equal(suite.T(), uint(100), responseBody.Quantity)
}

func (suite *TestSuite) TestGetUserFoodWithoutAuthorization() {
	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/user/food/1", nil)

	router.ServeHTTP(w, req)

	var responseBody tests.GenericErrorResponseBody
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 403, w.Code)
	assert.Equal(suite.T(), "Authentication failed", responseBody.Error)
}

func (suite *TestSuite) TestGetUserFoodNotFound() {
	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(email, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE `users`.`id` = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery(
		"^SELECT food_items.id, food_items.user_id, foods.id as food_id, foods.name, foods.calories, foods.portion, food_items.quantity, food_items.timestamp"+
			" FROM `food_items` "+
			"JOIN foods ON food_items.food_id = foods.id WHERE food_items.id = \\? AND food_items.user_id = \\? ORDER BY `food_items`.`id` LIMIT \\?",
	).
		WithArgs("1", float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "user_id", "food_id", "name", "calories", "portion", "quantity", "timestamp"}),
		)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	token := tests.GetToken(email, password)

	req, _ := http.NewRequest("GET", "/user/food/1", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, req)

	var responseBody tests.GenericErrorResponseBody
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 404, w.Code)
	assert.Equal(suite.T(), "Not found", responseBody.Error)
}

func (suite *TestSuite) TestPostUserFoodSuccessful() {
	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(email, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE `users`.`id` = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec("INSERT INTO `food_items`").
		WithArgs(1, 1, 100, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	suite.mock.ExpectCommit()

	suite.mock.ExpectQuery(
		"^SELECT food_items.id, food_items.user_id, foods.id as food_id, foods.name, foods.calories, foods.portion, food_items.quantity, food_items.timestamp"+
			" FROM `food_items` "+
			"JOIN foods ON food_items.food_id = foods.id WHERE food_items.id = \\? AND food_items.user_id = \\? ORDER BY `food_items`.`id` LIMIT \\?",
	).
		WithArgs(1, float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "user_id", "food_id", "name", "calories", "portion", "quantity", "timestamp"}).
				AddRow(1, 1, 1, "Pasta", 193, 80, 100, time.Now()),
		)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	token := tests.GetToken(email, password)

	userFoodData := models.FoodItem{
		FoodID:    1,
		Quantity:  100,
		Timestamp: time.Now(),
	}
	userFoodDataJson, _ := json.Marshal(userFoodData)

	req, _ := http.NewRequest("POST", "/user/food", strings.NewReader(string(userFoodDataJson)))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, req)

	var responseBody schemas.JoinedFoodItem
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 201, w.Code)

	assert.Equal(suite.T(), uint(1), responseBody.ID)
	assert.Equal(suite.T(), uint(1), responseBody.UserID)
	assert.Equal(suite.T(), uint(1), responseBody.FoodID)
	assert.Equal(suite.T(), "Pasta", responseBody.Name)
	assert.Equal(suite.T(), 193, responseBody.Calories)
	assert.Equal(suite.T(), 80, responseBody.Portion)
	assert.Equal(suite.T(), uint(100), responseBody.Quantity)
}

func (suite *TestSuite) TestPostUserFoodWithoutAuthorization() {
	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	userFoodData := models.FoodItem{
		FoodID:    1,
		Quantity:  100,
		Timestamp: time.Now(),
	}
	userFoodDataJson, _ := json.Marshal(userFoodData)

	req, _ := http.NewRequest("POST", "/user/food", strings.NewReader(string(userFoodDataJson)))

	router.ServeHTTP(w, req)

	var responseBody tests.GenericErrorResponseBody
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 403, w.Code)
	assert.Equal(suite.T(), "Authentication failed", responseBody.Error)
}

func (suite *TestSuite) TestPostUserFoodRequiresFoodID() {
	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(email, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE `users`.`id` = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	token := tests.GetToken(email, password)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	userFoodData := models.FoodItem{
		Quantity:  100,
		Timestamp: time.Now(),
	}
	userFoodDataJson, _ := json.Marshal(userFoodData)

	req, _ := http.NewRequest("POST", "/user/food", strings.NewReader(string(userFoodDataJson)))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 400, w.Code)
}

func (suite *TestSuite) TestPostUserFoodRequiresQuantity() {
	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(email, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE `users`.`id` = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	token := tests.GetToken(email, password)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	userFoodData := models.FoodItem{
		FoodID:    1,
		Timestamp: time.Now(),
	}
	userFoodDataJson, _ := json.Marshal(userFoodData)

	req, _ := http.NewRequest("POST", "/user/food", strings.NewReader(string(userFoodDataJson)))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 400, w.Code)
}

func (suite *TestSuite) TestPostUserFoodRequiresTimestamp() {
	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(email, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE `users`.`id` = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	token := tests.GetToken(email, password)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	userFoodData := models.FoodItem{
		FoodID:   1,
		Quantity: 100,
	}
	userFoodDataJson, _ := json.Marshal(userFoodData)

	req, _ := http.NewRequest("POST", "/user/food", strings.NewReader(string(userFoodDataJson)))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 400, w.Code)
}

func (suite *TestSuite) TestPutUserFoodSuccessful() {
	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(email, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE `users`.`id` = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery("^SELECT \\* FROM `food_items` WHERE id = \\? AND user_id = \\? ORDER BY `food_items`.`id` LIMIT \\?").
		WithArgs("1", float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "user_id", "food_id", "quantity", "timestamp"}).
				AddRow(1, 1, 1, 100, time.Now()),
		)

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec("UPDATE `food_items`").WillReturnResult(sqlmock.NewResult(1, 1))
	suite.mock.ExpectCommit()

	suite.mock.ExpectQuery(
		"^SELECT food_items.id, food_items.user_id, foods.id as food_id, foods.name, foods.calories, foods.portion, food_items.quantity, food_items.timestamp"+
			" FROM `food_items` "+
			"JOIN foods ON food_items.food_id = foods.id WHERE food_items.id = \\? AND food_items.user_id = \\? ORDER BY `food_items`.`id` LIMIT \\?",
	).
		WithArgs("1", float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "user_id", "food_id", "name", "calories", "portion", "quantity", "timestamp"}).
				AddRow(1, 1, 1, "Pasta", 193, 80, 120, time.Now()),
		)

	token := tests.GetToken(email, password)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	userFoodData := schemas.UpdateFoodItem{
		Quantity:  120,
		Timestamp: time.Now(),
	}
	userFoodDataJson, _ := json.Marshal(userFoodData)

	req, _ := http.NewRequest("PUT", "/user/food/1", strings.NewReader(string(userFoodDataJson)))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, req)

	var responseBody schemas.JoinedFoodItem
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 200, w.Code)

	assert.Equal(suite.T(), uint(1), responseBody.ID)
	assert.Equal(suite.T(), uint(1), responseBody.UserID)
	assert.Equal(suite.T(), uint(1), responseBody.FoodID)
	assert.Equal(suite.T(), "Pasta", responseBody.Name)
	assert.Equal(suite.T(), 193, responseBody.Calories)
	assert.Equal(suite.T(), 80, responseBody.Portion)
	assert.Equal(suite.T(), uint(120), responseBody.Quantity)
}

func (suite *TestSuite) TestPutUserFoodWithoutAuthorization() {
	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	userFoodData := schemas.UpdateFoodItem{
		Quantity:  120,
		Timestamp: time.Now(),
	}
	userFoodDataJson, _ := json.Marshal(userFoodData)

	req, _ := http.NewRequest("PUT", "/user/food/1", strings.NewReader(string(userFoodDataJson)))

	router.ServeHTTP(w, req)

	var responseBody tests.GenericErrorResponseBody
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 403, w.Code)
	assert.Equal(suite.T(), "Authentication failed", responseBody.Error)
}

func (suite *TestSuite) TestPutUserFoodRequiresQuantity() {
	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(email, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE `users`.`id` = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	token := tests.GetToken(email, password)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	userFoodData := schemas.UpdateFoodItem{
		Timestamp: time.Now(),
	}
	userFoodDataJson, _ := json.Marshal(userFoodData)

	req, _ := http.NewRequest("PUT", "/user/food/1", strings.NewReader(string(userFoodDataJson)))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 400, w.Code)
}

func (suite *TestSuite) TestPutUserFoodRequiresTimestamp() {
	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(email, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE `users`.`id` = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	token := tests.GetToken(email, password)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	userFoodData := schemas.UpdateFoodItem{
		Quantity: 120,
	}
	userFoodDataJson, _ := json.Marshal(userFoodData)

	req, _ := http.NewRequest("PUT", "/user/food/1", strings.NewReader(string(userFoodDataJson)))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 400, w.Code)
}

func (suite *TestSuite) TestPutUserFoodItemDoesNotBelongToUser() {
	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(email, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE `users`.`id` = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery("^SELECT \\* FROM `food_items` WHERE id = \\? AND user_id = \\? ORDER BY `food_items`.`id` LIMIT \\?").
		WithArgs("1", float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "user_id", "food_id", "quantity", "timestamp"}),
		)

	token := tests.GetToken(email, password)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	userFoodData := schemas.UpdateFoodItem{
		Quantity:  120,
		Timestamp: time.Now(),
	}
	userFoodDataJson, _ := json.Marshal(userFoodData)

	req, _ := http.NewRequest("PUT", "/user/food/1", strings.NewReader(string(userFoodDataJson)))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, req)

	var responseBody tests.GenericErrorResponseBody
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 404, w.Code)
	assert.Equal(suite.T(), "Not found", responseBody.Error)
}

func (suite *TestSuite) TestDeleteUserFoodSuccessful() {
	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(email, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE `users`.`id` = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery("^SELECT \\* FROM `food_items` WHERE id = \\? AND user_id = \\? ORDER BY `food_items`.`id` LIMIT \\?").
		WithArgs("1", float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "user_id", "food_id", "quantity", "timestamp"}).
				AddRow(1, 1, 1, 100, time.Now()),
		)

	suite.mock.ExpectBegin()
	suite.mock.ExpectExec("DELETE FROM `food_items`").WillReturnResult(sqlmock.NewResult(1, 1))
	suite.mock.ExpectCommit()

	token := tests.GetToken(email, password)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("DELETE", "/user/food/1", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 204, w.Code)
}

func (suite *TestSuite) TestDeleteUserFoodWithoutAuthorization() {
	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("DELETE", "/user/food/1", nil)

	router.ServeHTTP(w, req)

	var responseBody tests.GenericErrorResponseBody
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 403, w.Code)
	assert.Equal(suite.T(), "Authentication failed", responseBody.Error)
}

func (suite *TestSuite) TestDeleteUserFoodNotFound() {
	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(email, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE `users`.`id` = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	suite.mock.ExpectQuery("^SELECT \\* FROM `food_items` WHERE id = \\? AND user_id = \\? ORDER BY `food_items`.`id` LIMIT \\?").
		WithArgs("1", float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "user_id", "food_id", "quantity", "timestamp"}),
		)

	token := tests.GetToken(email, password)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("DELETE", "/user/food/1", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, req)

	var responseBody tests.GenericErrorResponseBody
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 404, w.Code)
	assert.Equal(suite.T(), "Not found", responseBody.Error)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
