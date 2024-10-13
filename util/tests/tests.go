package tests

import (
	"diet-app-backend/api/routes"
	"diet-app-backend/schemas"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
)

type LoginResponseBody struct {
	Token string
}

type GenericErrorResponseBody struct {
	Error string
}

func GetToken(email string, password string) string {
	router := routes.SetupRouter()
	w := httptest.NewRecorder()

	credentials := schemas.Credentials{
		Email:    email,
		Password: password,
	}
	credentialsJson, _ := json.Marshal(credentials)

	req, _ := http.NewRequest("POST", "/login", strings.NewReader(string(credentialsJson)))

	router.ServeHTTP(w, req)

	var responseBody LoginResponseBody
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	return responseBody.Token
}
