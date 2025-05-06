package registrationHandler

import (
	"backend/internal/resources"
	"backend/internal/utils"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegistrationRequestPOST(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name      string
		body      string
		setupMock func()
		wantCode  int
		wantBody  string
	}{
		{
			name: "Transaction start failed",
			body: `{"email": "test@example.com",
				  "password": "securepassword123'",
				  "team_name": "Supertesters",
				  "team_role": 1}`,
			setupMock: func() {},
			wantCode:  http.StatusInternalServerError,
			wantBody:  "Failed to start transaction\n",
		},
		{
			name: "Transaction rollback failed",
			body: `{"email": "test@example.com",
					  "password": "securepassword123",
					  "team_name": "Superstesters",
					  "team_role": 1}`,
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT email FROM users WHERE email = \\$1").
					WithArgs("test@example.com").
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery("SELECT id FROM team WHERE name = \\$1").
					WithArgs("Superstesters").
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback().WillReturnError(sql.ErrTxDone)
			},
			wantCode: http.StatusInternalServerError,
			wantBody: "database error checking team: sql: connection is already closed\nFailed to rollback transaction\n",
		},
		{
			name: "User already exists",
			body: `{"email": "test@example.com",
					"password": "securepassword123",
                    "team_name": "Supertesters",
				  	"team_role": 1}`,
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT email FROM users WHERE email = \\$1").
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow("test@example.com"))
			},
			wantCode: http.StatusConflict,
			wantBody: "user already exists\nFailed to rollback transaction\n",
		},
		{
			name: "Failed to register user",
			body: `{"email": "test@example.com",
					"password": "securepassword123",
					"team_name": "Supertesters",
					"team_role": 1}`,
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT email FROM users WHERE email = \\$1").
					WithArgs("test@example.com").WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			wantCode: http.StatusInternalServerError,
			wantBody: "database error checking user: sql: connection is already closed\n",
		},
		{
			name: "Transaction commit failed",
			body: `{"email": "test@example.com",
					 "password": "securepassword123",
					 "team_name": "Supertesters",
					 "team_role": 1}`,
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT email FROM users WHERE email = \\$1").
					WithArgs("test@example.com").WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery("SELECT id FROM team WHERE name = \\$1").
					WithArgs("Supertesters").WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery("INSERT INTO team \\( name, team_role\\) VALUES \\(\\$1, \\$2\\) RETURNING id").
					WithArgs("Supertesters", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectExec("INSERT INTO users \\(email, password, team_id\\) VALUES \\(\\$1, \\$2, \\$3\\)").
					WithArgs("test@example.com",
						sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit().WillReturnError(sql.ErrConnDone)
			},
			wantCode: http.StatusInternalServerError,
			wantBody: "Failed to commit the transaction.\nFailed to rollback transaction\n",
		},
		{
			name: "Successful registration",
			body: `{"email": "test@example.com",
				 "password": "securepassword123",
				 "team_name": "Supertesters",
				 "team_role": 1}`,
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT email FROM users WHERE email = \\$1").
					WithArgs("test@example.com").WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery("SELECT id FROM team WHERE name = \\$1").
					WithArgs("Supertesters").WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery("INSERT INTO team \\( name, team_role\\) VALUES \\(\\$1, \\$2\\) RETURNING id").
					WithArgs("Supertesters", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectExec("INSERT INTO users \\(email, password, team_id\\) VALUES \\(\\$1, \\$2, \\$3\\)").
					WithArgs("test@example.com",
						sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantCode: http.StatusCreated,
			wantBody: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			// Create a mock HTTP request and response recorder
			req := httptest.NewRequest(http.MethodPost, "/registration", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer validToken")
			rr := httptest.NewRecorder()

			RegistrationRequestPOST(rr, req, mockDB)

			assert.Equal(t, tt.wantCode, rr.Code)

			assert.Contains(t, tt.wantBody, rr.Body.String())

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func Test_checkExistingTeam(t *testing.T) {
	type args struct {
		tx          *sql.Tx
		credentials RegistrationPOSTRequest
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		want1   int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := checkExistingTeam(tt.args.tx, tt.args.credentials)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkExistingTeam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("checkExistingTeam() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("checkExistingTeam() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

/*
func Test_insertNewUser(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tx, _ := mockDB.Begin()

	tests := []struct {
		name        string
		credentials RegistrationPOSTRequest
		teamID      int
		setupMock   func()
		wantErr     bool
	}{

		{
			name: "Successful user insertion",
			credentials: RegistrationPOSTRequest{
				Email:    "test@example.com",
				Password: "securePassword123",
			},
			teamID: 1,
			setupMock: func() {
				mock.ExpectExec("INSERT INTO users \\(email, password, team_id\\) VALUES \\(\\$1, \\$2, \\$3\\)").
					WithArgs("test@example.com", sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "Database error during insertion",
			credentials: RegistrationPOSTRequest{
				Email:    "test@example.com",
				Password: "securePassword123",
			},
			teamID: 1,
			setupMock: func() {
				mock.ExpectExec("INSERT INTO users \\(email, password, team_id\\) VALUES \\(\\$1, \\$2, \\$3\\)").
					WithArgs("test@example.com", sqlmock.AnyArg(), 1).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations for this test case
			tt.setupMock()

			// Call the function we're testing
			err := insertNewUser(tx, tt.credentials, tt.teamID)

			// Check if error matches expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("insertNewUser() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify all expectations were met
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}*/

func Test_registerUserAndTeam(t *testing.T) {
	type args struct {
		tx          *sql.Tx
		credentials RegistrationPOSTRequest
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := registerUserAndTeam(tt.args.tx, tt.args.credentials)
			if (err != nil) != tt.wantErr {
				t.Errorf("registerUserAndTeam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("registerUserAndTeam() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistrationHandler(t *testing.T) {
	mockDB, _ := utils.InitMockDB(t)

	tests := []struct {
		name      string
		method    string
		body      string
		setupMock func()
		wantCode  int
		wantBody  string
	}{

		{
			name:      "Method not allowed",
			method:    http.MethodGet,
			body:      "",
			setupMock: func() {},
			wantCode:  http.StatusNotImplemented,
			wantBody:  resources.MethodNotAllowed,
		},
		{
			name:      "Invalid request body",
			method:    http.MethodPost,
			body:      `{"email":`,
			setupMock: func() {},
			wantCode:  http.StatusBadRequest,
			wantBody:  resources.InvalidPOSTRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(tt.method, "/register", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler := RegistrationHandler(mockDB)
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
