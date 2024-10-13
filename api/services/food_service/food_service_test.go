package foodservice_test

import (
	"database/sql"
	"diet-app-backend/api/routes"
	"diet-app-backend/database/connection"
	"diet-app-backend/database/models"
	"diet-app-backend/util/config"
	"diet-app-backend/util/tests"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
)

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

func (suite *TestSuite) TestGetFoodsSuccessful() {
	suite.mock.ExpectQuery("^SELECT `id`,`name`,`calories`,`portion` FROM `foods` WHERE name LIKE \\?").
		WithArgs(fmt.Sprintf("%%%s%%", "")).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "calories", "portion"}).
				AddRow(1, "Pasta", 193, 80).
				AddRow(2, "Grated Cheese", 492, 100).
				AddRow(3, "Tomato Sauce", 34, 100),
		)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/food", nil)

	router.ServeHTTP(w, req)

	var responseBody []models.Food
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 200, w.Code)

	assert.Len(suite.T(), responseBody, 3)

	assert.Equal(suite.T(), uint(1), responseBody[0].ID)
	assert.Equal(suite.T(), "Pasta", responseBody[0].Name)
	assert.Equal(suite.T(), 193, responseBody[0].Calories)
	assert.Equal(suite.T(), 80, responseBody[0].Portion)

	assert.Equal(suite.T(), uint(2), responseBody[1].ID)
	assert.Equal(suite.T(), "Grated Cheese", responseBody[1].Name)
	assert.Equal(suite.T(), 492, responseBody[1].Calories)
	assert.Equal(suite.T(), 100, responseBody[1].Portion)

	assert.Equal(suite.T(), uint(3), responseBody[2].ID)
	assert.Equal(suite.T(), "Tomato Sauce", responseBody[2].Name)
	assert.Equal(suite.T(), 34, responseBody[2].Calories)
	assert.Equal(suite.T(), 100, responseBody[2].Portion)
}

func (suite *TestSuite) TestGetFoodsSuccessfulWithQueryString() {
	suite.mock.ExpectQuery("^SELECT `id`,`name`,`calories`,`portion` FROM `foods` WHERE name LIKE \\?").
		WithArgs(fmt.Sprintf("%%%s%%", "Pas")).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "calories", "portion"}).
				AddRow(1, "Pasta", 193, 80),
		)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/food?name=Pas", nil)

	router.ServeHTTP(w, req)

	var responseBody []models.Food
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 200, w.Code)

	assert.Len(suite.T(), responseBody, 1)

	assert.Equal(suite.T(), uint(1), responseBody[0].ID)
	assert.Equal(suite.T(), "Pasta", responseBody[0].Name)
	assert.Equal(suite.T(), 193, responseBody[0].Calories)
	assert.Equal(suite.T(), 80, responseBody[0].Portion)
}

func (suite *TestSuite) TestGetFoodSuccessful() {
	suite.mock.ExpectQuery("^SELECT \\* FROM `foods` WHERE `foods`.`id` = \\? ORDER BY `foods`.`id` LIMIT \\?").
		WithArgs("1", 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "calories", "portion"}).
				AddRow(1, "Pasta", 193, 80),
		)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/food/1", nil)

	router.ServeHTTP(w, req)

	var responseBody models.Food
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 200, w.Code)

	assert.Equal(suite.T(), uint(1), responseBody.ID)
	assert.Equal(suite.T(), "Pasta", responseBody.Name)
	assert.Equal(suite.T(), 193, responseBody.Calories)
	assert.Equal(suite.T(), 80, responseBody.Portion)
}

func (suite *TestSuite) TestGetFoodNotFound() {
	suite.mock.ExpectQuery("^SELECT \\* FROM `foods` WHERE `foods`.`id` = \\? ORDER BY `foods`.`id` LIMIT \\?").
		WithArgs("1", 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "calories", "portion"}),
		)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/food/1", nil)

	router.ServeHTTP(w, req)

	var responseBody tests.GenericErrorResponseBody
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 404, w.Code)
	assert.Equal(suite.T(), "Not Found", responseBody.Error)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
