package usersHandler

import (
	"backend/internal/domain"
	"backend/internal/resources"
	"backend/internal/utils"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func authenticateAdminUser(mock sqlmock.Sqlmock) {
	// Auth token lookup
	mock.ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
		WithArgs("mockToken").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

	// Check if user role is admin
	mock.ExpectQuery(`SELECT user_role FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"user_role"}).AddRow(domain.Admin))
}

func TestUsersHandler(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name         string
		method       string
		path         string
		body         string
		setupMocks   func()
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Method = GET",
			method:       http.MethodGet,
			path:         "/users/1",
			body:         "",
			expectedCode: http.StatusOK,
			expectedBody: `{"email":"example@example.com","team_name":"Test team","user_role":"admin"}`,
			setupMocks: func() {
				authenticateAdminUser(mock)

				// First the code queries user attributes
				mock.ExpectQuery(`SELECT email, user_role, team_id FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"email", "user_role", "team_id"}).
						AddRow("example@example.com", domain.Admin, 1))

				// Then begins a transaction
				mock.ExpectBegin()

				// Team name query within transaction
				mock.ExpectQuery(`SELECT name FROM team WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Test team"))
			},
		},
		{
			name:         "Method = PATCH (Status OK)",
			method:       http.MethodPatch,
			path:         "/users/password",
			body:         `{"current_password":"securepassword123","new_password":"newpassword123"}`,
			expectedCode: http.StatusOK,
			expectedBody: ``,
			setupMocks: func() {
				// Auth token lookup
				mock.ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				// Transaction for password change
				mock.ExpectBegin()

				// Get current password hash
				mock.ExpectQuery(`SELECT password FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"password"}).
						AddRow("$argon2id$v=19$m=65536,t=3,p=2$/U4AzOF11CnxuCR8/fVdmA$u4xw+" +
							"7qRvex7t9BUSENhGDyiNsdb+inrgo3r4tpaflY"))

				// Update password
				mock.ExpectExec(`UPDATE users SET password = \$1 WHERE id = \$2`).
					WithArgs(sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Commit transaction
				mock.ExpectCommit()
			},
		},
		{
			name:         "Method = DELETE (Tested with removal later)",
			method:       http.MethodDelete,
			path:         "/users/2",
			body:         "",
			expectedCode: http.StatusUnauthorized,
			expectedBody: "",
			setupMocks: func() {
				// Auth checks
				authenticateAdminUser(mock)
			},
		},
		{
			name:         "Method = PUT (Status method not allowed)",
			method:       http.MethodPut,
			path:         "/users/1",
			body:         "",
			expectedCode: http.StatusNotImplemented,
			expectedBody: resources.MethodNotAllowed,
			setupMocks:   func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks for this test
			tt.setupMocks()

			// Create request
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			// Add auth token for all except PUT which should fail anyway
			if tt.method != http.MethodPut {
				req.Header.Set("Authorization", "Bearer mockToken")
			}

			rr := httptest.NewRecorder()

			// Call handler
			handler := UsersHandler(mockDB)
			handler.ServeHTTP(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedCode, rr.Code)

			// Check response body if expected
			if tt.expectedBody != "" {
				if tt.method == http.MethodPut {
					assert.Contains(t, rr.Body.String(), tt.expectedBody)
				} else {
					assert.JSONEq(t, tt.expectedBody, strings.TrimSpace(rr.Body.String()))
				}
			}

			// Verify all mocks were called
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestUsersRequestDELETE(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name         string
		path         string
		setupMocks   func()
		expectedCode int
		expectedBody string
	}{
		{
			name: "Invalid request URL",
			path: "/users0",
			setupMocks: func() {
				// Get user_id from middleware
				mock.ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				// Get user role from middleware
				mock.ExpectQuery(`SELECT user_role FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"user_role"}).AddRow(domain.Admin))
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: "Invalid request URL, no user_id, session_id found.",
		},
		{
			name: "The user is not an admin",
			path: "/users/2",
			setupMocks: func() {
				// Auth token lookup
				mock.ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				// Check if user role is admin
				mock.ExpectQuery(`SELECT user_role FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"user_role"}).AddRow(domain.Member))
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: resources.AuthenticationError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks for this test
			tt.setupMocks()

			// Create request
			req := httptest.NewRequest(http.MethodDelete, tt.path, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")
			rr := httptest.NewRecorder()

			UsersRequestDELETE(rr, req, mockDB)

			assert.Equal(t, tt.expectedCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestUsersRequestGET(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name         string
		path         string
		setupMocks   func()
		expectedCode int
		expectedBody string
	}{
		{
			name: "The user is not an admin",
			path: "/users/2",
			setupMocks: func() {
				// Auth token lookup
				mock.ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				// Check if user role is admin
				mock.ExpectQuery(`SELECT user_role FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"user_role"}).AddRow(domain.Member))
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: resources.AuthenticationError,
		},
		{
			name: "Get user information",
			path: "/users/1",
			setupMocks: func() {
				authenticateAdminUser(mock)

				// User attributes query
				mock.ExpectQuery(`SELECT email, user_role, team_id FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"email", "user_role", "team_id"}).
						AddRow("example@example.com", domain.Admin, 1))

				// Transaction for getUserInformation
				mock.ExpectBegin()

				// Team name query
				mock.ExpectQuery(`SELECT name FROM team WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Test team"))
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"email":"example@example.com","team_name":"Test team","user_role":"admin"}`,
		},
		{
			name: "Get list of all active sessions",
			path: "/users/1/sessions?status=active",
			setupMocks: func() {
				authenticateAdminUser(mock)

				// Active sessions query
				mock.ExpectQuery(`SELECT created_at, ip_address FROM sessions WHERE user_id = \$1 AND status = \$2`).
					WithArgs(1, "active").
					WillReturnRows(sqlmock.NewRows([]string{"created_at", "ip_address"}).
						AddRow(time.Date(2023, 6, 15, 14, 30, 45, 0, time.UTC), "192.168.1.1").
						AddRow(time.Date(2023, 6, 16, 9, 15, 30, 0, time.UTC), "192.168.1.2"))
			},
			expectedCode: http.StatusOK,
			expectedBody: `[{"id":1,"ip":"192.168.1.1","last_active":"2023-06-15T14:30:45Z"},{"id":2,"ip":"192.168.1.2","last_active":"2023-06-15T14:30:45Z"},{"id":3,"ip":"192.168.1.1","last_active":"2023-06-16T09:15:30Z"},{"id":4,"ip":"192.168.1.2","last_active":"2023-06-16T09:15:30Z"}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks for this test
			tt.setupMocks()

			// Create request
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")
			rr := httptest.NewRecorder()

			UsersRequestGET(rr, req, mockDB)

			assert.Equal(t, tt.expectedCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestUsersRequestPATCH(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name         string
		path         string
		body         string
		setupMocks   func()
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Invalid request URL",
			path:         "///users////1",
			body:         "",
			setupMocks:   func() {},
			expectedCode: http.StatusBadRequest,
			expectedBody: "Invalid request URL, use '/users/password' to change password.\n",
		},
		{
			name: "Transaction start failed",
			path: "/users/password",
			body: `{"current_password":"securepassword123","new_password":"newpassword123"}`,
			setupMocks: func() {
				// Auth token lookup
				mock.ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				// Transaction for password change
				mock.ExpectBegin().WillReturnError(sql.ErrConnDone)
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: resources.TransactionStartFailed,
		},
		{
			name: "Invalid request body",
			path: "/users/password",
			body: `{"current_password":"securepassword123","new_password":"hhhh"}`, //to short password
			setupMocks: func() {
				// Auth token lookup
				mock.ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				// Transaction for password change
				mock.ExpectBegin()
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "Invalid request body.\n",
		},
		{
			name: "Invalid PATCH request",
			path: "/users/password",
			body: `{"current_password":"securepassword123","new_password":243657699}`, //to short password
			setupMocks: func() {
				// Auth token lookup
				mock.ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				// Transaction for password change
				mock.ExpectBegin()
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: resources.InvalidPATCHRequest,
		},
		{
			name: "Could not retrieve the current password from the database",
			path: "/users/password",
			body: `{"current_password":"securepassword123","new_password":"newpassword123"}`,
			setupMocks: func() {
				// Auth token lookup
				mock.ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				// Transaction for password change
				mock.ExpectBegin()

				// Get current password hash
				mock.ExpectQuery(`SELECT password FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnError(sql.ErrNoRows)
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Could not change the current password.\n",
		},
		/*{
			name: "Could not hash the new password",
			path: "/users/password",
			body: `{"current_password":"securepassword123","new_password":"newpassword123"}`,
			setupMocks: func() {
				// Auth token lookup
				mock.ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				// Transaction for password change
				mock.ExpectBegin()

				// Get current password hash
				mock.ExpectQuery(`SELECT password FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"password"}).
						AddRow("$argon2id$v=19$m=65536,t=3,p=2$/U4AzOF11CnxuCR8/fVdmA$u4xw+" +
							"7qRvex7t9BUSENhGDyiNsdb+inrgo3r4tpaflY"))

			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Could not hash the new password.",
		},*/
		{
			name: "Failed to commit the transaction",
			path: "/users/password",
			body: `{"current_password":"securepassword123","new_password":"newpassword123"}`,
			setupMocks: func() {
				// Auth token lookup
				mock.ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				// Transaction for password change
				mock.ExpectBegin()

				// Get current password hash
				mock.ExpectQuery(`SELECT password FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"password"}).
						AddRow("$argon2id$v=19$m=65536,t=3,p=2$/U4AzOF11CnxuCR8/fVdmA$u4xw+" +
							"7qRvex7t9BUSENhGDyiNsdb+inrgo3r4tpaflY"))

				// Update password
				mock.ExpectExec(`UPDATE users SET password = \$1 WHERE id = \$2`).
					WithArgs(sqlmock.AnyArg(), 1).WillReturnError(sql.ErrTxDone)

			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Could not update the password",
		},
		{
			name: "Failed to commit the transaction",
			path: "/users/password",
			body: `{"current_password":"securepassword123","new_password":"newpassword123"}`,
			setupMocks: func() {
				// Auth token lookup
				mock.ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				// Transaction for password change
				mock.ExpectBegin()

				// Get current password hash
				mock.ExpectQuery(`SELECT password FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"password"}).
						AddRow("$argon2id$v=19$m=65536,t=3,p=2$/U4AzOF11CnxuCR8/fVdmA$u4xw+" +
							"7qRvex7t9BUSENhGDyiNsdb+inrgo3r4tpaflY"))

				// Update password
				mock.ExpectExec(`UPDATE users SET password = \$1 WHERE id = \$2`).
					WithArgs(sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Commit transaction
				mock.ExpectCommit().WillReturnError(sql.ErrNoRows)
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: resources.TransactionCommitFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks for this test
			tt.setupMocks()

			// Create request
			req := httptest.NewRequest(http.MethodPatch, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")
			rr := httptest.NewRecorder()

			UsersRequestPATCH(rr, req, mockDB)

			assert.Equal(t, tt.expectedCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func Test_deleteUserSession(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name        string
		sessionID   int
		userID      int
		adminTeamID int
		setupMocks  func(*sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name:        "Could no retrieve team ID",
			sessionID:   1,
			userID:      1,
			adminTeamID: 1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				(*mock).ExpectBegin()

				(*mock).ExpectQuery("SELECT team_id FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name:        "Valid session deletion",
			sessionID:   1,
			userID:      1,
			adminTeamID: 1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				(*mock).ExpectBegin()

				// Get team ID
				(*mock).ExpectQuery("SELECT team_id FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(1))

				// Delete session
				(*mock).ExpectExec("DELETE FROM sessions WHERE id = \\$1 AND user_id = \\$2").
					WithArgs(1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Commit transaction
				(*mock).ExpectCommit()
			},
			wantErr: false,
		}, /*
				{
					name:        "User is not part of the same team as the admin",
					sessionID:   1,
					userID:      1,
					adminTeamID: 1,
					setupMocks: func(mock *sqlmock.Sqlmock) {
						(*mock).ExpectBegin()

						// Get team ID
						(*mock).ExpectQuery("SELECT team_id FROM users WHERE id = \\$1").
							WithArgs(1).
							WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(2))

					},
					wantErr: true,
				},
			{
				name:        "Failed to delete session",
				sessionID:   1,
				userID:      1,
				adminTeamID: 1,
				setupMocks: func(mock *sqlmock.Sqlmock) {
					(*mock).ExpectBegin()

					// Get team ID
					(*mock).ExpectQuery("SELECT team_id FROM users WHERE id = \\$1").
						WithArgs(1).
						WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(1))

					(*mock).ExpectExec("DELETE FROM sessions WHERE id = \\$1 AND user_id = \\$2").
						WithArgs(1, 1).
						WillReturnError(sql.ErrNoRows)
					(*mock).ExpectRollback().WillReturnError(sql.ErrTxDone)
				},
				wantErr: true,
			},*/
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks for this test before beginning the transaction
			tt.setupMocks(&mock)

			// Now begin the transaction
			mockTX, err := mockDB.Begin()
			if err != nil {
				t.Errorf("failed to begin transaction: %v", err)
			}

			if err = deleteUserSession(mockTX, tt.sessionID, tt.userID, tt.adminTeamID); (err != nil) != tt.wantErr {
				t.Errorf("deleteUserSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getAndCheckTeamID(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name       string
		adminID    int
		setupMocks func(*sqlmock.Sqlmock)
		want       int
		wantErr    bool
	}{
		{
			name:    "Valid team ID",
			adminID: 1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				(*mock).ExpectBegin()

				(*mock).ExpectQuery("SELECT team_id FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(2))
			},
			want:    2,
			wantErr: false,
		},
		{
			name:    "Invalid team ID",
			adminID: 1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				(*mock).ExpectBegin()

				(*mock).ExpectQuery("SELECT team_id FROM users WHERE id = \\$1").
					WithArgs(1).WillReturnError(sql.ErrNoRows)
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks for this test before beginning the transaction
			tt.setupMocks(&mock)

			// Now begin the transaction
			mockTX, err := mockDB.Begin()
			if err != nil {
				t.Fatalf("failed to begin transaction: %v", err)
			}

			var got int
			got, err = getAndCheckTeamID(mockTX, tt.adminID)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAndCheckTeamID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("getAndCheckTeamID() got = %v, want %v", got, tt.want)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func Test_getListOfAllActiveSessions(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name       string
		path       string
		setupMocks func()
		wantBody   string
	}{
		{
			name:       "Could not find any active sessions",
			path:       "/users/1",
			setupMocks: func() {},
			wantBody:   "Could not find any active sessions for user with id: 1\n",
		},
		{
			name: "Could not scan session row",
			path: "/users/1/sessions?status=active",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT created_at, ip_address FROM sessions WHERE user_id = \$1 AND status = \$2`).
					WithArgs(1, "active").
					WillReturnRows(sqlmock.NewRows([]string{"created_at", "ip_address"}).
						AddRow("invalid-time-format", "192.168.1.1"))
			},
			wantBody: "Could not scan session row.\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks for this test
			tt.setupMocks()

			// Create request
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")
			rr := httptest.NewRecorder()

			getListOfAllActiveSessions(rr, req, mockDB)

			assert.Contains(t, rr.Body.String(), tt.wantBody)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func Test_getTeamName(t *testing.T) {
	type args struct {
		w         http.ResponseWriter
		tx        *sql.Tx
		userIdStr string
		teamID    int
		err       error
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTeamName(tt.args.w, tt.args.tx, tt.args.userIdStr, tt.args.teamID, tt.args.err); got != tt.want {
				t.Errorf("getTeamName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getUserInformation(t *testing.T) {
	type args struct {
		w  http.ResponseWriter
		r  *http.Request
		db *sql.DB
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getUserInformation(tt.args.w, tt.args.r, tt.args.db)
		})
	}
}

func Test_removeActiveSession(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name       string
		teamID     int
		validPath  []string
		setupMocks func()
		wantBody   string
		wantCode   int
	}{
		{
			name:      "Failed to commit the transaction",
			teamID:    1,
			validPath: []string{"users", "1", "sessions", "2"},
			setupMocks: func() {
				mock.ExpectBegin()

				// Get team ID
				mock.ExpectQuery("SELECT team_id FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(1))

				mock.ExpectExec("DELETE FROM sessions WHERE id = \\$1 AND user_id = \\$2").
					WithArgs(2, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit().WillReturnError(sql.ErrConnDone)
			},
			wantBody: resources.TransactionCommitFailed + "\n",
			wantCode: http.StatusInternalServerError,
		},
		{
			name:       "Invalid request URL",
			teamID:     1,
			validPath:  []string{},
			setupMocks: func() {},
			wantBody:   "Invalid request URL, no user_id or session_id found.\n",
			wantCode:   http.StatusBadRequest,
		},
		{
			name:      "Transaction start failed",
			teamID:    1,
			validPath: []string{"users", "1", "sessions", "2"},
			setupMocks: func() {
				mock.ExpectBegin().WillReturnError(errors.New("transaction failed"))
			},
			wantBody: resources.TransactionStartFailed + "\n",
			wantCode: http.StatusInternalServerError,
		},
		{
			name:      "Failed to delete session",
			teamID:    1,
			validPath: []string{"users", "1", "sessions", "2"},
			setupMocks: func() {
				mock.ExpectBegin()
				// Get team ID
				mock.ExpectQuery("SELECT team_id FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(1))

				mock.ExpectExec("DELETE FROM sessions WHERE id = \\$1 AND user_id = \\$2").
					WithArgs(2, 1).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectRollback().WillReturnError(sql.ErrTxDone)
			},
			wantBody: "Failed to delete session:",
			wantCode: http.StatusInternalServerError,
		},
		/*
			{
				name:      "Valid session removal",
				teamID:    1,
				validPath: []string{"users", "1", "sessions", "2"},
				setupMocks: func() {
					mock.ExpectBegin()

					// Get team ID
					mock.ExpectQuery("SELECT team_id FROM users WHERE id = \\$1").
						WithArgs(1).
						WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(1))

					// Delete session
					mock.ExpectExec("DELETE FROM sessions WHERE id = \\$1 AND user_id = \\$2").
						WithArgs(2, 1).
						WillReturnResult(sqlmock.NewResult(1, 1))

					// Commit transaction
					mock.ExpectCommit()
				},
				wantBody: "",
				wantCode: http.StatusNoContent,
			},*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks for this test
			tt.setupMocks()

			// Create request
			var req *http.Request
			if tt.name != "Invalid request URL" {
				req = httptest.NewRequest(http.MethodDelete,
					"/users/1/sessions/2", nil)
			} else {
				req = httptest.NewRequest(http.MethodDelete, "/users", nil)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")
			rr := httptest.NewRecorder()

			removeActiveSession(rr, req, mockDB, tt.teamID, tt.validPath)

			assert.Equal(t, tt.wantCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.wantBody)
		})
	}
}

func Test_removeUserFromTeam(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name        string
		adminTeamID int
		validPath   []string
		setupMocks  func()
		wantBody    string
	}{
		{
			name:        "Invalid request URL",
			adminTeamID: 1,
			validPath:   []string{},
			setupMocks:  func() {},
			wantBody:    "Invalid request URL, no user_id found.\n",
		},
		{
			name:        "Temporary team creation failed",
			adminTeamID: 1,
			validPath:   []string{"users", "1"},
			setupMocks: func() {
				mock.ExpectBegin()

				// Get team ID
				mock.ExpectQuery("SELECT team_id FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(1))

				// Create temporary team
				mock.ExpectQuery("INSERT INTO team \\(name, team_role\\) VALUES \\(\\$1, \\$2\\) RETURNING id").
					WithArgs("Temporary team", domain.Researcher).
					WillReturnError(sql.ErrNoRows)
			},
			wantBody: "Failed to create temporary team: ",
		},
		{
			name:        "Failed to remove user from team",
			adminTeamID: 1,
			validPath:   []string{"users", "1"},
			setupMocks: func() {
				mock.ExpectBegin()

				// Get team ID
				mock.ExpectQuery("SELECT team_id FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(1))

				// Create temporary team
				mock.ExpectQuery("INSERT INTO team \\(name, team_role\\) VALUES \\(\\$1, \\$2\\) RETURNING id").
					WithArgs("Temporary team", domain.Researcher).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(999))

				// Remove user from team
				mock.ExpectExec("UPDATE users SET team_id = \\$1 WHERE id = \\$2 AND team_id = \\$3").
					WithArgs(999, 1, 1).
					WillReturnError(sql.ErrNoRows)
			},
			wantBody: "Failed to remove user from team",
		},
		{
			name:        "Failed to commit the transaction",
			adminTeamID: 1,
			validPath:   []string{"users", "1"},
			setupMocks: func() {
				mock.ExpectBegin()

				// Get team ID
				mock.ExpectQuery("SELECT team_id FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(1))

				// Create temporary team
				mock.ExpectQuery("INSERT INTO team \\(name, team_role\\) VALUES \\(\\$1, \\$2\\) RETURNING id").
					WithArgs("Temporary team", domain.Researcher).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(999))

				// Remove user from team
				mock.ExpectExec("UPDATE users SET team_id = \\$1 WHERE id = \\$2 AND team_id = \\$3").
					WithArgs(999, 1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit().WillReturnError(sql.ErrTxDone)
			},
			wantBody: resources.TransactionCommitFailed,
		},
		{
			name:        "The user is not part of the same team as the admin",
			adminTeamID: 1,
			validPath:   []string{"users", "1"},
			setupMocks: func() {
				mock.ExpectBegin()
			},
			wantBody: "The user is not part of your team.",
		},
		{
			name:        "Valid user removal from team",
			adminTeamID: 1,
			validPath:   []string{"users", "1"},
			setupMocks: func() {
				mock.ExpectBegin()

				// Get team ID
				mock.ExpectQuery("SELECT team_id FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(1))

				// Create temporary team
				mock.ExpectQuery("INSERT INTO team \\(name, team_role\\) VALUES \\(\\$1, \\$2\\) RETURNING id").
					WithArgs("Temporary team", domain.Researcher).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(999))

				// Remove user from team
				mock.ExpectExec("UPDATE users SET team_id = \\$1 WHERE id = \\$2 AND team_id = \\$3").
					WithArgs(999, 1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
			wantBody: "{\"message\":\"User removed from team successfully.\",\"temp_team_id\":999}\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks for this test
			tt.setupMocks()

			// Create request
			var req *http.Request
			if tt.name != "Invalid request URL" {
				req = httptest.NewRequest(http.MethodDelete, "/users/1", nil)
			} else {
				req = httptest.NewRequest(http.MethodDelete, "/users", nil)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")
			rr := httptest.NewRecorder()

			removeUserFromTeam(rr, req, mockDB, tt.adminTeamID, tt.validPath)

			assert.Contains(t, rr.Body.String(), tt.wantBody)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func Test_writeUsersResponse(t *testing.T) {
	type args struct {
		w        http.ResponseWriter
		response any
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeUsersResponse(tt.args.w, tt.args.response)
		})
	}
}
