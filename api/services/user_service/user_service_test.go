package userservice_test

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
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

func (suite *TestSuite) TestLoginSuccessful() {
	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(email, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	credentials := schemas.Credentials{
		Email:    email,
		Password: password,
	}
	credentialsJson, _ := json.Marshal(credentials)

	req, _ := http.NewRequest("POST", "/login", strings.NewReader(string(credentialsJson)))

	router.ServeHTTP(w, req)

	var responseBody tests.LoginResponseBody
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 200, w.Code)
	assert.NotNil(suite.T(), responseBody.Token)
}

func (suite *TestSuite) TestLoginUserDoesNotExist() {
	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(email, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}),
		)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	credentials := schemas.Credentials{
		Email:    email,
		Password: password,
	}
	credentialsJson, _ := json.Marshal(credentials)

	req, _ := http.NewRequest("POST", "/login", strings.NewReader(string(credentialsJson)))

	router.ServeHTTP(w, req)

	var responseBody tests.GenericErrorResponseBody
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 403, w.Code)
	assert.Equal(suite.T(), "Invalid credentials", responseBody.Error)
}

func (suite *TestSuite) TestLoginWrongPassword() {
	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(email, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	credentials := schemas.Credentials{
		Email:    email,
		Password: "invalid",
	}
	credentialsJson, _ := json.Marshal(credentials)

	req, _ := http.NewRequest("POST", "/login", strings.NewReader(string(credentialsJson)))

	router.ServeHTTP(w, req)

	var responseBody tests.GenericErrorResponseBody
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 403, w.Code)
	assert.Equal(suite.T(), "Invalid credentials", responseBody.Error)
}

func (suite *TestSuite) TestSignupSuccessful() {
	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	userData := models.User{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Password:  password,
	}
	userDataJson, _ := json.Marshal(userData)

	req, _ := http.NewRequest("POST", "/signup", strings.NewReader(string(userDataJson)))

	router.ServeHTTP(w, req)

	var responseBody models.User
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 201, w.Code)
	assert.Equal(suite.T(), email, responseBody.Email)
	assert.Equal(suite.T(), firstName, responseBody.FirstName)
	assert.Equal(suite.T(), lastName, responseBody.LastName)
}

func (suite *TestSuite) TestSignupRequiresEmail() {
	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	userData := models.User{
		FirstName: firstName,
		LastName:  lastName,
		Password:  password,
	}
	userDataJson, _ := json.Marshal(userData)

	req, _ := http.NewRequest("POST", "/signup", strings.NewReader(string(userDataJson)))

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 400, w.Code)
}

func (suite *TestSuite) TestSignupRequiresFirstName() {
	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	userData := models.User{
		Email:    email,
		LastName: lastName,
		Password: password,
	}
	userDataJson, _ := json.Marshal(userData)

	req, _ := http.NewRequest("POST", "/signup", strings.NewReader(string(userDataJson)))

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 400, w.Code)
}

func (suite *TestSuite) TestSignupRequiresLastName() {
	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	userData := models.User{
		Email:     email,
		FirstName: firstName,
		Password:  password,
	}
	userDataJson, _ := json.Marshal(userData)

	req, _ := http.NewRequest("POST", "/signup", strings.NewReader(string(userDataJson)))

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 400, w.Code)
}

func (suite *TestSuite) TestSignupRequiresPassword() {
	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	userData := models.User{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}
	userDataJson, _ := json.Marshal(userData)

	req, _ := http.NewRequest("POST", "/signup", strings.NewReader(string(userDataJson)))

	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 400, w.Code)
}

func (suite *TestSuite) TestSignupRequiresUniqueEmail() {
	suite.mock.ExpectBegin()
	suite.mock.ExpectExec("INSERT INTO `users`").
		WithArgs(email, firstName, lastName, sqlmock.AnyArg()).
		WillReturnError(
			errors.New("Duplicate entry"),
		)
	suite.mock.ExpectRollback()

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	userData := models.User{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Password:  password,
	}
	userDataJson, _ := json.Marshal(userData)

	req, _ := http.NewRequest("POST", "/signup", strings.NewReader(string(userDataJson)))

	router.ServeHTTP(w, req)

	var responseBody tests.GenericErrorResponseBody
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 409, w.Code)
	assert.Equal(suite.T(), "This email is not available", responseBody.Error)
}

func (suite *TestSuite) TestGetUser() {
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

	suite.mock.ExpectQuery("^SELECT \\* FROM `users` WHERE `users`.`id` = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(float64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "password"}).
				AddRow(1, email, firstName, lastName, hashedPassword),
		)

	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	token := tests.GetToken(email, password)

	req, _ := http.NewRequest("GET", "/user", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	router.ServeHTTP(w, req)

	var responseBody models.User
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 200, w.Code)
	assert.Equal(suite.T(), email, responseBody.Email)
	assert.Equal(suite.T(), firstName, responseBody.FirstName)
	assert.Equal(suite.T(), lastName, responseBody.LastName)
}

func (suite *TestSuite) TestGetUserWithoutAuthorization() {
	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/user", nil)

	router.ServeHTTP(w, req)

	var responseBody tests.GenericErrorResponseBody
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Equal(suite.T(), 403, w.Code)
	assert.Equal(suite.T(), "Authentication failed", responseBody.Error)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
