package loginHandler

import (
	"backend/internal/domain"
	"backend/internal/resources"
	"backend/internal/utils"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLoginHandler(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)

	tests := []struct {
		name      string
		method    string
		body      string
		setupMock func()
		wantCode  int
		wantBody  string
	}{
		{
			name:   "Successful login",
			method: http.MethodPost,
			body:   `{"email": "example@example.com","password": "securepassword123"}`,
			setupMock: func() {
				mock.ExpectQuery("SELECT id, email, password FROM users WHERE email = \\$1").
					WithArgs("example@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password"}).AddRow(
						1, "example@example.com", "$argon2id$v=19$m=65536,t=3,p=2$/U4AzOF11CnxuCR8/fVdmA$u4xw+"+
							"7qRvex7t9BUSENhGDyiNsdb+inrgo3r4tpaflY"))

				mock.ExpectExec("INSERT INTO sessions \\(user_id, session_token, expires_at, ip_address\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
					WithArgs(1, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantCode: http.StatusOK,
			wantBody: `{"expires_at":"`,
		},
		{
			name:      "Method not allowed",
			method:    http.MethodGet,
			body:      "",
			setupMock: func() {},
			wantCode:  http.StatusMethodNotAllowed,
			wantBody:  resources.MethodNotAllowed,
		},
		{
			name:      "Invalid request body",
			method:    http.MethodPost,
			body:      `{"email":`,
			setupMock: func() {},
			wantCode:  http.StatusBadRequest,
			wantBody:  "Invalid request data",
		},
		{
			name:   "Invalid email or password",
			method: http.MethodPost,
			body:   `{"email":"test@example.com","password":"securepassword123"}`,
			setupMock: func() {
				mock.ExpectQuery("SELECT id, email, password FROM users WHERE email = \\$1").
					WithArgs("test@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			wantCode: http.StatusUnauthorized,
			wantBody: "Invalid email or password",
		},
		{
			name:   "Password validation failed",
			method: http.MethodPost,
			body:   `{"email":"test@example.com","password":"securepassword123"}`,
			setupMock: func() {
				mock.ExpectQuery("SELECT id, email, password FROM users WHERE email = \\$1").
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password"}).AddRow(
						1, "test@example.com", "$argon2")) // Invalid hash for testing
			},
			wantCode: http.StatusUnauthorized,
			wantBody: "Invalid email or password",
		},
		{
			name:   "Could not create session",
			method: http.MethodPost,
			body:   `{"email": "example@example.com","password": "securepassword123"}`,
			setupMock: func() {
				mock.ExpectQuery("SELECT id, email, password FROM users WHERE email = \\$1").
					WithArgs("example@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password"}).AddRow(
						1, "example@example.com", "$argon2id$v=19$m=65536,t=3,p=2$/U4AzOF11CnxuCR8/fVdmA$u4xw+"+
							"7qRvex7t9BUSENhGDyiNsdb+inrgo3r4tpaflY"))
				mock.ExpectExec(
					"INSERT INTO sessions \\(user_id, session_token, expires_at\\) VALUES \\(\\$1, \\$2, \\$3\\)").
					WithArgs(1, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantCode: http.StatusInternalServerError,
			wantBody: "Could not create session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(tt.method, "/login", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler := LoginHandler(mockDB)
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

func Test_generateSecureLoginToken(t *testing.T) {
	type args struct {
		tokenLength int
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"Test invalid token length ", args{0}, 0, true},
		{"Test valid token length", args{32}, 44, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateSecureLoginToken(tt.args.tokenLength)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateSecureLoginToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.want {
				t.Errorf("generateSecureLoginToken() got = %v, want length %v", len(got), tt.want)
			}
		})
	}
}

func TestCheckUserExists(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name      string
		email     string
		setupMock func()
		want      domain.User
		wantErr   bool
	}{
		{
			name:  "User exists",
			email: "example@example.com",
			setupMock: func() {
				mock.ExpectQuery("SELECT id, email, password FROM users WHERE email = \\$1").
					WithArgs("example@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password"}).AddRow(
						1, "example@example.com", "$argon2id$v=19$m=65536,t=3,p=2$/U4AzOF11CnxuCR8/fVdmA$u4xw+"+
							"7qRvex7t9BUSENhGDyiNsdb+inrgo3r4tpaflY"))
			},
			want: domain.User{
				ID:    1,
				Email: "example@example.com",
				Password: "$argon2id$v=19$m=65536,t=3,p=2$/U4AzOF11CnxuCR8/fVdmA$u4xw+" +
					"7qRvex7t9BUSENhGDyiNsdb+inrgo3r4tpaflY",
			},
			wantErr: false,
		},
		{
			name:  "User does not exist",
			email: "",
			setupMock: func() {
				mock.ExpectQuery("SELECT id, email, password FROM users WHERE email = \\$1").
					WithArgs("").
					WillReturnError(sql.ErrNoRows)
			},
			want:    domain.User{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodPost, "/login", nil)
			req.Header.Set("Content-Type", "application/json")

			got, err := CheckUserExists(mockDB, tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckUserExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckUserExists() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateSession(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name      string
		userID    int
		token     string
		ip        string
		setupMock func()
		wantErr   bool
	}{
		{
			name:   "Successful session creation",
			userID: 1,
			token:  "mockToken",
			ip:     "127.0.0.1",
			setupMock: func() {
				mock.ExpectExec("INSERT INTO sessions \\(user_id, session_token, expires_at, ip_address\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
					WithArgs(1, "mockToken", sqlmock.AnyArg(), "127.0.0.1").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:   "Failed session creation",
			userID: 1,
			ip:     "127.0.0.1",
			setupMock: func() {

			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodPost, "/products", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tt.token)

			if err := CreateSession(mockDB, tt.userID, tt.token, time.Now().Add(24*time.Hour), tt.ip); (err != nil) != tt.wantErr {
				t.Errorf("CreateSession() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	tests := []struct {
		name             string
		providedPassword string
		storedHash       string
		wantErr          bool
	}{
		{
			name:             "Correct password",
			providedPassword: "securepassword123",
			storedHash: "$argon2id$v=19$m=65536,t=3,p=2$/U4AzOF11CnxuCR8/fVdmA$u4xw+" +
				"7qRvex7t9BUSENhGDyiNsdb+inrgo3r4tpaflY",
			wantErr: false,
		},
		{
			name:             "Incorrect password",
			providedPassword: "wrongpassword",
			storedHash:       "wrongpassword",
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := VerifyPassword(tt.providedPassword, tt.storedHash); (err != nil) != tt.wantErr {
				t.Errorf("VerifyPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
