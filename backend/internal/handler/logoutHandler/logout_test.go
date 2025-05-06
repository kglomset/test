package logoutHandler

import (
	"backend/internal/resources"
	"backend/internal/utils"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogoutHandler(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)

	tests := []struct {
		name       string
		method     string
		mockExpect func()
		wantCode   int
		wantBody   string
	}{
		{
			name:       "Method not allowed",
			method:     http.MethodGet,
			mockExpect: func() {},
			wantCode:   http.StatusMethodNotAllowed,
			wantBody:   resources.MethodNotAllowed,
		},
		{
			name:       "Authentication header is missing",
			method:     http.MethodPost,
			mockExpect: func() {},
			wantCode:   http.StatusUnauthorized,
			wantBody:   resources.AuthenticationError,
		},
		{
			name:   "Successfully logged out",
			method: http.MethodPost,
			mockExpect: func() {
				mock.
					ExpectExec("UPDATE sessions SET status = 'expired' WHERE session_token = \\$1").
					WithArgs("token").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantCode: http.StatusOK,
			wantBody: "You have successfully logged out!",
		},
		{
			name:   "Could not log out",
			method: http.MethodPost,
			mockExpect: func() {
				mock.
					ExpectExec("DELETE FROM sessions WHERE session_token = $1").
					WithArgs("token").
					WillReturnError(sql.ErrNoRows)
			},
			wantCode: http.StatusInternalServerError,
			wantBody: "Could not log out",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()

			req := httptest.NewRequest(tt.method, "/logout", nil)
			req.Header.Set("Content-Type", "application/json")
			if tt.name != "Authentication header is missing" {
				req.Header.Set("Authorization", "Bearer token")
			}

			rr := httptest.NewRecorder()

			handler := LogoutHandler(mockDB)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.wantCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.wantCode)
			}

			if !strings.Contains(rr.Body.String(), tt.wantBody) {
				t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), tt.wantBody)
			}
		})
	}
}
