package middleware

import (
	"backend/internal/utils"
	"bytes"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthMiddleware(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name       string
		token      string
		setupMocks func()
		wantStatus int
	}{
		{
			name:  "Successful authentication",
			token: "Bearer mockToken",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT s\.user_id, s\.session_token, s\.expires_at FROM sessions s LEFT JOIN public\.users u ON s\.user_id = u\.id WHERE s\.session_token = \$1 AND s\.expires_at > NOW\(\)`).
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id", "session_token", "expires_at"}).
						AddRow(1, "mockToken", time.Now().Add(1*time.Hour)))
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "No token provided",
			setupMocks: func() {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:  "Expired token",
			token: "Bearer mockToken",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT s\.user_id, s\.session_token, s\.expires_at FROM sessions s LEFT JOIN public\.users u ON s\.user_id = u\.id WHERE s\.session_token = \$1 AND s\.expires_at > NOW\(\)`).
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id", "session_token", "expires_at"})) // Using empty rows to simulate expired token
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:  "DB error",
			token: "Bearer mockToken",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT s\.user_id, s\.session_token, s\.expires_at FROM sessions s LEFT JOIN public\.users u ON s\.user_id = u\.id WHERE s\.session_token = \$1 AND s\.expires_at > NOW\(\)`).
					WithArgs("mockToken").
					WillReturnError(sql.ErrConnDone) // Simulating a database connection error with sql.ErrConnDone
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock expectations
			tt.setupMocks()

			// Create a mock HTTP request and response recorder
			req := httptest.NewRequest(http.MethodGet, "/products", nil)
			req.Header.Set("Content-Type", "application/json")

			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}

			rr := httptest.NewRecorder()

			// Create a test handler
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			handler := NewAuthHandler(mockDB).Middleware(nextHandler)

			handler.ServeHTTP(rr, req)

			// Check the response status code
			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.wantStatus)
			}

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestLoggingMiddleware(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)

	tests := []struct {
		name       string
		token      string
		path       string
		setupMocks func()
		wantReq    string
		wantRes    string
	}{
		{
			name:       "OK",
			token:      "validToken",
			path:       "/products",
			setupMocks: func() {},
			wantReq:    "Request: POST /products, On: ",
			wantRes:    "Response: [INFO] 200 OK, On: ",
		},
		{
			name:       "/login log without userID",
			token:      "validToken",
			path:       "/login",
			setupMocks: func() {},
			wantReq:    "Request: POST /login, On: ",
			wantRes:    "Response: [INFO] 200 OK, On: ",
		},
		{
			name:       "Bad request",
			token:      "validToken",
			path:       "/products",
			setupMocks: func() {},
			wantReq:    "Request: POST /products, On: ",
			wantRes:    "Response: [WARN] 400 Bad Request, On: ",
		},
		{
			name:       "Status internal server error",
			token:      "validToken",
			path:       "/products",
			setupMocks: func() {},
			wantReq:    "Request: POST /products, On: ",
			wantRes:    "Response: [ERROR] 500 Internal Server Error, On: ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock expectations
			tt.setupMocks()

			// Capture log output
			var logBuffer bytes.Buffer
			log.SetOutput(&logBuffer)
			defer log.SetOutput(nil) // Restore log output after test

			// Create a mock HTTP request and response recorder
			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tt.token)

			rr := httptest.NewRecorder()

			// Create a test handler
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch tt.name {
				case "Status internal server error":
					w.WriteHeader(http.StatusInternalServerError)
				case "Bad request":
					w.WriteHeader(http.StatusBadRequest)
				default:
					w.WriteHeader(http.StatusOK)
				}
			})

			handler := NewLoggingHandler(mockDB).LoggingMiddleware(nextHandler)

			handler.ServeHTTP(rr, req)

			/// Read captured log output
			logOutput := logBuffer.String()

			// Check the request log
			if !bytes.Contains([]byte(logOutput), []byte(tt.wantReq)) {
				t.Errorf("expected request log to contain %q, but got %q", tt.wantReq, logOutput)
			}

			// Check the response log
			if !bytes.Contains([]byte(logOutput), []byte(tt.wantRes)) {
				t.Errorf("expected response log to contain %q, but got %q", tt.wantRes, logOutput)
			}

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func Test_statusResponseWriter_WriteHeader(t *testing.T) {
	_, mock := utils.InitMockDB(t)

	tests := []struct {
		name       string
		status     int
		wantStatus int
	}{
		{
			name:       "Set status code 200",
			status:     http.StatusOK,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Set status code 500",
			status:     http.StatusInternalServerError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Set status code 404",
			status:     http.StatusNotFound,
			wantStatus: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			srw := &statusResponseWriter{
				ResponseWriter: rr,
			}
			srw.WriteHeader(tt.status)

			// Check the returned status code
			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.wantStatus)
			}

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name       string
		token      string
		setupMocks func()
		want       int
	}{
		{
			name:  "Valid user ID",
			token: "validToken",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("validToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))
			},
			want: 1,
		},
		{
			name:       "No token provided",
			token:      "",
			setupMocks: func() {},
			want:       0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock expectations
			tt.setupMocks()
			// Create a mock HTTP request and response recorder
			req := httptest.NewRequest(http.MethodGet, "/products", nil)
			req.Header.Set("Content-Type", "application/json")
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			rr := httptest.NewRecorder()

			if got := GetUserID(rr, req, mockDB); got != tt.want {
				t.Errorf("GetUserID() = %v, want %v", got, tt.want)
			}

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestGetUserTeamRole(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name       string
		token      string
		setupMocks func()
		want       int
		status     int
	}{
		{
			name:  "Valid user role and team",
			token: "validToken",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("validToken").WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				mock.ExpectQuery("SELECT team_id FROM users WHERE id = \\$1;").
					WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(1))

				mock.ExpectQuery("SELECT team_role FROM team WHERE id = \\$1;").
					WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"team_role"}).AddRow(1))
			},
			want:   1,
			status: http.StatusOK,
		},
		{
			name:  "No team role found",
			token: "validToken",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("validToken").WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				mock.ExpectQuery("SELECT team_id FROM users WHERE id = \\$1;").
					WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(1))

				mock.ExpectQuery("SELECT team_role FROM team WHERE id = \\$1;").
					WithArgs(1).WillReturnError(sql.ErrNoRows)
			},
			want:   0,
			status: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock expectations
			tt.setupMocks()
			// Create a mock HTTP request and response recorder
			req := httptest.NewRequest(http.MethodGet, "/products", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tt.token)
			rr := httptest.NewRecorder()

			got := GetUserTeamRole(rr, req, mockDB)
			if got != tt.want {
				t.Errorf("GetUserTeamRole() got = %v, want %v", got, tt.want)
			}

			assert.Equal(t, rr.Code, tt.status)

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestGetUserRole(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name       string
		token      string
		setupMocks func()
		want       string
		status     int
	}{
		{
			name:  "Valid user role",
			token: "validToken",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("validToken").WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				mock.ExpectQuery("SELECT user_role FROM users WHERE id = \\$1;").
					WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"user_role"}).AddRow("admin"))
			},
			want:   "admin",
			status: http.StatusOK,
		},
		{
			name:  "Invalid user role",
			token: "validToken",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("validToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				mock.ExpectQuery("SELECT user_role FROM users WHERE id = \\$1;").
					WithArgs(1).WillReturnError(sql.ErrNoRows)
			},
			want:   "",
			status: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			// Create a mock HTTP request and response recorder
			req := httptest.NewRequest(http.MethodGet, "/products", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tt.token)
			rr := httptest.NewRecorder()

			if got := GetUserRole(rr, req, mockDB); got != tt.want {
				t.Errorf("GetUserRole() = %v, want %v", got, tt.want)
			}

			assert.Equal(t, rr.Code, tt.status)

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestGetUserTeamID(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name       string
		token      string
		setupMocks func()
		want       int
		status     int
	}{
		{
			name:  "Valid team ID",
			token: "validToken",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("validToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				mock.ExpectQuery("SELECT team_id FROM users WHERE id = \\$1;").
					WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(1))
			},
			want:   1,
			status: http.StatusOK,
		},
		{
			name:  "No team ID found",
			token: "validToken",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("validToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				mock.ExpectQuery("SELECT team_id FROM users WHERE id = \\$1;").
					WithArgs(1).WillReturnError(sql.ErrNoRows)
			},
			want:   0,
			status: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock expectations
			tt.setupMocks()

			// Create a mock HTTP request and response recorder
			req := httptest.NewRequest(http.MethodGet, "/products", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tt.token)
			rr := httptest.NewRecorder()

			if got := GetUserTeamID(rr, req, mockDB); got != tt.want {
				t.Errorf("GetUserTeamID() = %v, want %v", got, tt.want)
			}

			assert.Equal(t, rr.Code, tt.status)

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}
