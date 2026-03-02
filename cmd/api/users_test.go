package main

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/O-Nikitin/Social/internal/store"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
)

func TestUsers_GetUser(t *testing.T) {
	app, mocks := newTestApp(t)
	mux := app.mount()
	testToken := "abc123"
	t.Run("Should_not_allow_unauthentificated_requests",
		func(t *testing.T) {
			req, err := http.NewRequest(
				http.MethodGet,
				"/v1/users/1",
				nil)
			if err != nil {
				t.Fatal("Request not created: ", err)
			}

			rr := executeRequest(req, mux)

			checkResponseCode(rr.Code, http.StatusUnauthorized, t)
		})

	t.Run("DB_call_without_cache_No_errors",
		func(t *testing.T) {
			userID := int64(42)
			req, err := http.NewRequest(
				http.MethodGet,
				"/v1/users/42",
				nil)
			if err != nil {
				t.Fatal("Request not created: ", err)
			}

			req.Header.Set("Authorization", "Bearer "+testToken)
			mockJwtToken := &jwt.Token{
				Valid: true,
				Claims: jwt.MapClaims{
					"sub": float64(1), // jwt.MapClaims decodes numbers as float64
				},
			}
			mocks.Auth.EXPECT().ValidateToken(testToken).Return(mockJwtToken, nil)

			expectedUser := &store.User{
				ID:       userID,
				Username: "john_doe",
				Email:    "john@example.com",
			}
			mocks.Users.EXPECT().GetByID(gomock.Any(), int64(1)).Return(expectedUser, nil)
			mocks.Users.EXPECT().GetByID(gomock.Any(), userID).Return(expectedUser, nil)

			rr := executeRequest(req, mux)

			checkResponseCode(rr.Code, http.StatusOK, t)

			checkContentType("application/json", rr, t)

			var response struct {
				Data *store.User `json:"data"`
			}
			json.Unmarshal(rr.Body.Bytes(), &response)

			if !reflect.DeepEqual(expectedUser, response.Data) {
				t.Errorf("expected user to be %v got %v", expectedUser, response.Data)
			}
		})
}
