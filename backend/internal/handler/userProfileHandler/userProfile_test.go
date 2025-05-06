package userProfileHandler

import (
	"backend/internal/domain"
	"backend/internal/resources"
	"backend/internal/utils"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var testUser = domain.User{
	ID:       1,
	Email:    "test@test.com",
	Password: "password",
	Team:     1,
	UserRole: "admin",
}

func TestGetUserAttributes(t *testing.T) {
	tests := []struct {
		name      string
		userID    int
		user      *domain.User
		setupMock func(*sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:   "successful retrieval",
			userID: 1,
			user:   &testUser,
			setupMock: func(mock *sqlmock.Sqlmock) {
				// Mock the database query
				(*mock).ExpectQuery("SELECT email, user_role, team_id FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"email", "user_role", "team_id"}).
						AddRow("test@test.com", "admin", 1))
			},
		},
		{
			name:   "user not found",
			userID: 2,
			user:   &domain.User{},
			setupMock: func(mock *sqlmock.Sqlmock) {
				// Begin a transaction
				(*mock).ExpectBegin()

				// Mock the database query
				(*mock).ExpectQuery("SELECT email, user_role, team_id FROM users WHERE id = \\$1").
					WithArgs(2).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock := utils.InitMockDB(t)
			tt.setupMock(&mock)

			if err := GetUserAttributes(mockDB, tt.userID, tt.user); (err != nil) != tt.wantErr {
				t.Errorf("GetUserAttributes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserProfileHandler(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		path         string
		setupMocks   func(*sqlmock.Sqlmock)
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Invalid request URL",
			method:       http.MethodGet,
			path:         "/use",
			setupMocks:   func(mock *sqlmock.Sqlmock) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `Invalid request URL`,
		},
		{
			name:         "Method not allowed",
			method:       http.MethodPut,
			path:         "/user/profile",
			setupMocks:   func(mock *sqlmock.Sqlmock) {},
			expectedCode: http.StatusNotImplemented,
			expectedBody: resources.MethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock := utils.InitMockDB(t)

			// Setup mocks for this test
			tt.setupMocks(&mock)

			// Create request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Content-Type", "application/json")

			// Add auth token for all except PUT which should fail anyway
			if tt.method != http.MethodPut {
				req.Header.Set("Authorization", "Bearer mockToken")
			}

			rr := httptest.NewRecorder()

			// Call handler
			handler := UserProfileHandler(mockDB)
			handler.ServeHTTP(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)

			// Verify all mocks were called
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestUserProfileRequestGET(t *testing.T) {
	tests := []struct {
		name         string
		setupMocks   func(*sqlmock.Sqlmock)
		expectedCode int
		expectedBody string
	}{
		{
			name: "User not found",
			setupMocks: func(mock *sqlmock.Sqlmock) {
				(*mock).ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				(*mock).ExpectQuery("SELECT email, user_role, team_id FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnError(sql.ErrNoRows)
			},
			expectedCode: http.StatusNotFound,
			expectedBody: resources.UserNotFound,
		},
		/*{
			name: "Invalid team role",
			setupMocks: func(mock *sqlmock.Sqlmock) {
				(*mock).ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				(*mock).ExpectQuery("SELECT email, user_role, team_id FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"email", "user_role", "team_id"}).
						AddRow("test@example.com", "admin", 1))

				(*mock).ExpectQuery("SELECT name, team_role FROM team WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"name", "team_role"}).
						AddRow("Test Team", 999)) // Using invalid team role value
			},
			expectedCode: http.StatusNotFound,
			expectedBody: "Invalid team role",
		},*/
		{
			name: "Successful retrieval (Official)",
			setupMocks: func(mock *sqlmock.Sqlmock) {
				(*mock).ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				(*mock).ExpectQuery("SELECT email, user_role, team_id FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"email", "user_role", "team_id"}).
						AddRow("test@example.com", "admin", 1))

				(*mock).ExpectQuery("SELECT name, team_role FROM team WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"name", "team_role"}).
						AddRow("Test Team", 1))
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"email":"test@example.com","user_role":"admin","team":{"name":"Test Team","team_role":"Official"}}`,
		},
		{
			name: "Invalid team role",
			setupMocks: func(mock *sqlmock.Sqlmock) {
				(*mock).ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				(*mock).ExpectQuery("SELECT email, user_role, team_id FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"email", "user_role", "team_id"}).
						AddRow("test@example.com", "admin", 1))

				(*mock).ExpectQuery("SELECT name, team_role FROM team WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"name", "team_role"}).
						AddRow("Test Team", 999)) // Using invalid team role value
			},
			expectedCode: http.StatusNotFound,
			expectedBody: "Invalid team role\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock := utils.InitMockDB(t)

			// Setup mocks for this test
			tt.setupMocks(&mock)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/user/profile", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")
			rr := httptest.NewRecorder()

			UserProfileRequestGET(rr, req, mockDB)

			assert.Equal(t, tt.expectedCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)

			// Verify all mocks were called
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func Test_getTeamAttributes(t *testing.T) {
	tests := []struct {
		name       string
		user       *domain.User
		team       *domain.Team
		setupMocks func(*sqlmock.Sqlmock)
		err        error
	}{
		{
			name: "Could not find team name and role",
			user: &domain.User{ID: 1, Team: 1},
			team: &domain.Team{},
			setupMocks: func(mock *sqlmock.Sqlmock) {
				(*mock).ExpectQuery("SELECT name, team_role FROM team WHERE id = \\$1").
					WithArgs(1).
					WillReturnError(sql.ErrNoRows)
			},
			err: sql.ErrNoRows,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock := utils.InitMockDB(t)

			// Setup mocks for this test
			tt.setupMocks(&mock)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/user/profile", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")
			rr := httptest.NewRecorder()

			getTeamAttributes(rr, mockDB, tt.user, tt.team, tt.err)
		})
	}
}
