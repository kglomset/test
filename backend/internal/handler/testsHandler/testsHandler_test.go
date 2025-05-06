package testsHandler

import (
	"backend/internal/domain"
	"backend/internal/utils"
	"bytes"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func AuthenticationMock(mock sqlmock.Sqlmock) {
	// Mock the user id query
	mock.ExpectQuery("SELECT user_id FROM sessions WHERE \\(session_token = \\$1 AND expires_at > NOW\\(\\)\\)").
		WithArgs("mockToken").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

	// Mock the user team id query
	mock.ExpectQuery("SELECT team_id FROM users WHERE id = \\$1;").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(1))

	// Mock the user team role query
	mock.ExpectQuery("SELECT team_role FROM team WHERE id = \\$1;").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"team_role"}).AddRow(1))
}

var testsColumns = []string{
	"id", "test_date", "location", "comment", "sc_id", "tc_id", "ac_id", "version", "is_public", "testing_team",
}

// Mock rows for the sql.Rows interface
type mockRows struct {
	rows   [][]driver.Value
	rowPos int
	cols   []string
}

func (m *mockRows) Next() bool {
	m.rowPos++
	return m.rowPos <= len(m.rows)
}

func (m *mockRows) Scan(dest ...interface{}) error {
	if m.rowPos <= 0 || m.rowPos > len(m.rows) {
		return errors.New("scan called without next")
	}
	row := m.rows[m.rowPos-1]
	for i, val := range row {
		d := dest[i].(*interface{})
		*d = val
	}
	return nil
}

func (m *mockRows) Close() error {
	return nil
}

func (m *mockRows) Columns() ([]string, error) {
	return m.cols, nil
}

func (m *mockRows) Err() error {
	return nil
}

func TestFetchTests(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	tests := []struct {
		name           string
		setupMock      func(sqlmock.Sqlmock)
		err            error
		expectedStatus int
		expectedTests  []domain.Test
	}{
		{
			name: "Valid data return",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(testsColumns).
					AddRow(1, time.Now(), "location 1", "Test Comment", 1, 1, 1, time.Now(), true, 1).
					AddRow(2, time.Now(), "Location 2", "Test Comment 2", 2, 2, 2, time.Now(), true, 1)

				// Either return this directly or have your test function use it
				mock.ExpectQuery("SELECT .* FROM tests").WillReturnRows(rows)
			},
			err:            nil,
			expectedStatus: http.StatusOK,
			expectedTests:  []domain.Test{},
		},
		// Other test cases...
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			testList := []domain.Test{}

			if tc.setupMock != nil {
				tc.setupMock(mock)
			}

			// If FetchTests executes a query directly:
			var rows *sql.Rows
			rows, err = db.Query("SELECT * FROM tests")
			if err != nil {
				err = tc.err // Use the test case error
			}

			// Call the function with real sql.Rows
			FetchTests(w, testList, rows, err)

			// Verify response
			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestGetTestWithID(t *testing.T) {
	// Create a new mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Create test cases
	tests := []struct {
		name           string
		testID         int
		mockSetup      func()
		expectedTest   domain.Test
		expectedStatus int
	}{
		{
			name:   "Test found",
			testID: 2,
			mockSetup: func() {
				cols := []string{"id", "test_date", "location", "comment", "sc_id", "tc_id", "ac_id", "version", "publicly_available", "testing_team"}
				mock.ExpectQuery("SELECT \\* FROM tests WHERE id = \\$1").
					WithArgs(2).
					WillReturnRows(sqlmock.NewRows(cols).
						AddRow(2, time.Now(), "Location 2", "Test Comment 2", 2, 2, 2, time.Now(), true, 1))
			},
			expectedTest:   domain.Test{ID: 2},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Test not found",
			testID: 999,
			mockSetup: func() {
				mock.ExpectQuery("SELECT \\* FROM tests WHERE id = \\$1").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expectedTest:   domain.Test{},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()
			w := httptest.NewRecorder()

			// Call the function
			result := GetTestWithID(w, db, tc.testID)

			// If we expect to find the test, check the result has the expected ID
			if tc.expectedTest.ID > 0 {
				assert.Equal(t, tc.expectedTest.ID, result.ID)
			} else {
				assert.Equal(t, tc.expectedTest, result)
			}

			// Ensure all expectations were met
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestParseAndValidateTestRequest(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		expectedErr bool
	}{
		{
			name: "Valid request",
			requestBody: `{
				  "sc": {
						"temperature": -10,
						"snow_type": "FS",
						"snow_humidity": "W2"
					},
					"ac": {
						"temperature": -15,
						"humidity": 30,
						"wind": "L",
						"cloud": "1"
					},
					"tc": {
						"track_hardness": "H1",
						"track_type": "D1"
					},
				  "location": "Holmenkollen, Oslo",
				  "date": "2025-03-30T10:30:00Z",
				  "comment": "Test conducted under typical winter conditions. Excellent glide.",
				  "is_public": false,
				  "testing_team": 1,
				  "test_ranks": [
					{
					  "product_id": 1,
					  "rank": 1,
					  "distance_behind": 0,
					  "is_public": true
					},
					{
					  "product_id": 2,
					  "rank": 2,
					  "distance_behind": 5,
					  "is_public": true
					}
				  ]
				}`,
			expectedErr: false,
		},
		{
			name: "Invalid JSON format",
			requestBody: `{
				"date": "2023-01-01",
				"comment": "Test comment",
				"snow_conditions": {
					"temperature": -1,
					"snow_type_artificial": "",
					"snow_type_natural": "FS",
					"snow_humidity": "W1"
				},
				"track_conditions": {
					"track_hardness": "H2",
					"track_type": "T2"
				},
				"air_conditions": {
					"temperature": -5,
					"humidity": 40,
					"wind": "2",
					"cloud": "L"
				},
				"version": "2023-01-01T00:00:00Z",
				"publicly_available": false,
				"created_by": 1,
				"testing_team": 2,
				 "distances": [
					{
					  "winner": {
						"test_id": 1,
						"rank": 1,
						"wins": 5,
						"is_public": true,
						"version": "2023-01-01T00:00:00Z",
						"created_by": 1,
						"testing_team": 2,
						"test_product": {
						  "rank_id": 1,
						  "glider_id": 101,
						  "mid_layer_id": 201,
						  "top_layer_id": {
							"products": [301, 302, 303]
						  }
						}
					  },
					  "distance_between": 20
					}
				  ]
			}`,
			expectedErr: true,
		},
		{
			name: "Missing required fields",
			requestBody: `{
				"date": "2023-01-01",
				"comment": "Test comment"
			}`,
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/tests", bytes.NewBufferString(tc.requestBody))

			// Call the function
			_, err := utils.ParseAndValidateRequest[TestPOSTRequest](r)

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInsertSnowConditions(t *testing.T) {
	// Create a new mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Begin transaction
	mock.ExpectBegin()
	tx, _ := db.Begin()

	// Test snow conditions
	conditions := SnowConditionsPOST{
		Temperature:  -5.4,
		SnowType:     "A1",
		SnowHumidity: "W1",
	}

	// Expect query and return ID
	mock.ExpectQuery("INSERT INTO snow_conditions").
		WithArgs(conditions.Temperature,
			conditions.SnowType, conditions.SnowHumidity).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Call the function
	id, err := insertSnowConditions(tx, conditions)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 1, id)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestInsertAirConditions(t *testing.T) {
	// Create a new mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Begin transaction
	mock.ExpectBegin()
	tx, _ := db.Begin()

	// Test air conditions
	conditions := AirConditionsPOST{
		Temperature: -10,
		Humidity:    40,
		Wind:        "L",
		Cloud:       "2",
	}

	// Expect query and return ID
	mock.ExpectQuery("INSERT INTO air_conditions").
		WithArgs(conditions.Temperature, conditions.Humidity,
			conditions.Wind, conditions.Cloud).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	// Call the function
	id, err := insertAirConditions(tx, conditions)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 2, id)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestInsertTrackConditions(t *testing.T) {
	// Create a new mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Begin transaction
	mock.ExpectBegin()
	tx, _ := db.Begin()

	// Test track conditions
	conditions := TrackConditionsPOST{
		TrackHardness: "H1",
		TrackType:     "Downhill",
	}

	// Expect query and return ID
	mock.ExpectQuery("INSERT INTO track_conditions").
		WithArgs(conditions.TrackHardness, conditions.TrackType).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

	// Call the function
	id, err := insertTrackConditions(tx, conditions)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 3, id)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateTest(t *testing.T) {
	// Create a new mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Begin transaction
	mock.ExpectBegin()
	tx, _ := db.Begin()

	// Test data
	testData := TestPOSTRequest{
		Date:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Location: "Test Location",
		Comment:  "Test Comment",
		SnowConditions: SnowConditionsPOST{
			Temperature:  -5,
			SnowType:     "A1",
			SnowHumidity: "DS",
		},
		TrackConditions: TrackConditionsPOST{
			TrackHardness: "H1",
			TrackType:     "Downhill",
		},
		AirConditions: AirConditionsPOST{
			Temperature: -10,
			Humidity:    40,
			Wind:        "L",
			Cloud:       "2",
		},
		IsPublic:    false,
		TestingTeam: 1,
	}
	team := 2

	// Mock snow conditions insertion
	mock.ExpectQuery("INSERT INTO snow_conditions").
		WithArgs(testData.SnowConditions.Temperature, testData.SnowConditions.SnowType,
			testData.SnowConditions.SnowHumidity).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Mock air conditions insertion
	mock.ExpectQuery("INSERT INTO air_conditions").
		WithArgs(testData.AirConditions.Temperature, testData.AirConditions.Humidity,
			testData.AirConditions.Wind, testData.AirConditions.Cloud).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	// Mock track conditions insertion
	mock.ExpectQuery("INSERT INTO track_conditions").
		WithArgs(testData.TrackConditions.TrackHardness, testData.TrackConditions.TrackType).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

	// Using sqlmock.AnyArg() for the timestamp
	mock.ExpectQuery("INSERT INTO tests").
		WithArgs(
			testData.Date,
			testData.Location,
			testData.Comment,
			1,                // snow condition ID
			3,                // track condition ID
			2,                // air condition ID
			sqlmock.AnyArg(), // Use AnyArg() for time.Now()
			testData.IsPublic,
			team,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))

	// Call the function
	id, err := createTest(tx, testData, team)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 5, id)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_TestsHandler(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name         string
		method       string
		path         string
		body         string
		setupMocks   func()
		db           *sql.DB
		expectedCode int
	}{
		{
			name:         "Method = GET (Status OK - no tests found)",
			method:       http.MethodGet,
			path:         "/tests/1",
			body:         ``,
			expectedCode: http.StatusOK,
			setupMocks: func() {
				AuthenticationMock(mock)

				mock.ExpectQuery("SELECT \\* FROM tests").
					WillReturnRows(sqlmock.NewRows(testsColumns))
			},
		},
		{
			name:         "Method = POST (Status OK)",
			method:       http.MethodPost,
			path:         "/tests",
			body:         `{"sc":{"temperature":-10,"snow_type":"FS","snow_humidity":"W2"},"ac":{"temperature":-15,"humidity":30,"wind":"L","cloud":"1"},"tc":{"track_hardness":"H1","track_type":"D1"},"location":"Holmenkollen, Oslo","date":"2025-03-30T10:30:00Z","comment":"Test conducted under typical winter conditions. Excellent glide.","is_public":false,"testing_team":1,"test_ranks":[{"product_id":1,"rank":1,"distance_behind":0}]}`,
			expectedCode: http.StatusCreated,
			setupMocks: func() {
				AuthenticationMock(mock)

				// Begin transaction
				mock.ExpectBegin()

				// Mock snow conditions insertion
				mock.ExpectQuery("INSERT INTO snow_conditions").
					WithArgs(-10.0, "FS", "W2").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				// Mock air conditions insertion
				mock.ExpectQuery("INSERT INTO air_conditions").
					WithArgs(-15.0, 30, "L", "1").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

				// Mock track conditions insertion
				mock.ExpectQuery("INSERT INTO track_conditions").
					WithArgs("H1", "D1").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

				// Mock test insertion
				mock.ExpectQuery("INSERT INTO tests").
					WithArgs(
						sqlmock.AnyArg(), // date
						"Holmenkollen, Oslo",
						"Test conducted under typical winter conditions. Excellent glide.",
						1,                // sc_id
						3,                // tc_id
						2,                // ac_id
						sqlmock.AnyArg(), // version
						false,            // is_public
						1,                // testing_team
					).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				// Get product availability
				mock.ExpectQuery("SELECT is_public FROM products WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				// Mock test ranks insertion
				mock.ExpectExec("INSERT INTO test_ranks").
					WithArgs(1, 1, 1, 0, sqlmock.AnyArg(), true).
					WillReturnResult(sqlmock.NewResult(5, 1))

				// Commit transaction
				mock.ExpectCommit()
			},
		},
		{
			name:         "Method = POST (Status internal server error - failed to create test)",
			method:       http.MethodPost,
			path:         "/tests",
			body:         `{"sc":{"temperature":-10,"snow_type":"FS","snow_humidity":"W2"},"ac":{"temperature":-15,"humidity":30,"wind":"L","cloud":"1"},"tc":{"track_hardness":"H1","track_type":"D1"},"location":"Holmenkollen, Oslo","date":"2025-03-30T10:30:00Z","comment":"Test conducted under typical winter conditions. Excellent glide.","is_public":false,"testing_team":1,"test_ranks":[{"product_id":1,"rank":1,"distance_behind":0}]}`,
			expectedCode: http.StatusInternalServerError,
			setupMocks: func() {
				AuthenticationMock(mock)

				// Begin transaction
				mock.ExpectBegin()

				// Mock snow conditions insertion
				mock.ExpectQuery("INSERT INTO snow_conditions").
					WithArgs(-10.0, "FS", "W2").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				// Mock air conditions insertion
				mock.ExpectQuery("INSERT INTO air_conditions").
					WithArgs(-15.0, 30, "L", "1").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

				// Mock track conditions insertion
				mock.ExpectQuery("INSERT INTO track_conditions").
					WithArgs("H1", "D1").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

				// Mock test insertion
				mock.ExpectQuery("INSERT INTO tests").
					WithArgs(
						sqlmock.AnyArg(), // date
						"Holmenkollen, Oslo",
						"Test conducted under typical winter conditions. Excellent glide.",
						1,                // sc_id
						3,                // tc_id
						2,                // ac_id
						sqlmock.AnyArg(), // version
						false,            // is_public
						1,                // testing_team
					).
					WillReturnError(errors.New("test error"))
			},
		},
		{
			name:         "Method = POST  (Status internal server error - failed to create rankings)",
			method:       http.MethodPost,
			path:         "/tests",
			body:         `{"sc":{"temperature":-10,"snow_type":"FS","snow_humidity":"W2"},"ac":{"temperature":-15,"humidity":30,"wind":"L","cloud":"1"},"tc":{"track_hardness":"H1","track_type":"D1"},"location":"Holmenkollen, Oslo","date":"2025-03-30T10:30:00Z","comment":"Test conducted under typical winter conditions. Excellent glide.","is_public":false,"testing_team":1,"test_ranks":[{"product_id":1,"rank":1,"distance_behind":0}]}`,
			expectedCode: http.StatusInternalServerError,
			setupMocks: func() {
				AuthenticationMock(mock)

				// Begin transaction
				mock.ExpectBegin()

				// Mock snow conditions insertion
				mock.ExpectQuery("INSERT INTO snow_conditions").
					WithArgs(-10.0, "FS", "W2").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				// Mock air conditions insertion
				mock.ExpectQuery("INSERT INTO air_conditions").
					WithArgs(-15.0, 30, "L", "1").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

				// Mock track conditions insertion
				mock.ExpectQuery("INSERT INTO track_conditions").
					WithArgs("H1", "D1").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

				// Mock test insertion
				mock.ExpectQuery("INSERT INTO tests").
					WithArgs(
						sqlmock.AnyArg(), // date
						"Holmenkollen, Oslo",
						"Test conducted under typical winter conditions. Excellent glide.",
						1,                // sc_id
						3,                // tc_id
						2,                // ac_id
						sqlmock.AnyArg(), // version
						false,            // is_public
						1,                // testing_team
					).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				// Get product availability
				mock.ExpectQuery("SELECT is_public FROM products WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				// Mock test ranks insertion
				mock.ExpectExec("INSERT INTO test_ranks").
					WithArgs(1, 1, 1, 0, sqlmock.AnyArg(), true).
					WillReturnError(errors.New("mock error")) // Simulate an error
			},
		},
		{
			name:         "Method = POST (Status internal server error - failed to commit transaction)",
			method:       http.MethodPost,
			path:         "/tests",
			body:         `{"sc":{"temperature":-10,"snow_type":"FS","snow_humidity":"W2"},"ac":{"temperature":-15,"humidity":30,"wind":"L","cloud":"1"},"tc":{"track_hardness":"H1","track_type":"D1"},"location":"Holmenkollen, Oslo","date":"2025-03-30T10:30:00Z","comment":"Test conducted under typical winter conditions. Excellent glide.","is_public":false,"testing_team":1,"test_ranks":[{"product_id":1,"rank":1,"distance_behind":0}]}`,
			expectedCode: http.StatusInternalServerError,
			setupMocks: func() {
				AuthenticationMock(mock)

				// Begin transaction
				mock.ExpectBegin()

				// Mock snow conditions insertion
				mock.ExpectQuery("INSERT INTO snow_conditions").
					WithArgs(-10.0, "FS", "W2").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				// Mock air conditions insertion
				mock.ExpectQuery("INSERT INTO air_conditions").
					WithArgs(-15.0, 30, "L", "1").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

				// Mock track conditions insertion
				mock.ExpectQuery("INSERT INTO track_conditions").
					WithArgs("H1", "D1").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

				// Mock test insertion
				mock.ExpectQuery("INSERT INTO tests").
					WithArgs(
						sqlmock.AnyArg(), // date
						"Holmenkollen, Oslo",
						"Test conducted under typical winter conditions. Excellent glide.",
						1,                // sc_id
						3,                // tc_id
						2,                // ac_id
						sqlmock.AnyArg(), // version
						false,            // is_public
						1,                // testing_team
					).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				// Get product availability
				mock.ExpectQuery("SELECT is_public FROM products WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				// Mock test ranks insertion
				mock.ExpectExec("INSERT INTO test_ranks").
					WithArgs(1, 1, 1, 0, sqlmock.AnyArg(), true).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Commit transaction
				mock.ExpectCommit().WillReturnError(errors.New("mock error"))
			},
		},
		{
			name:         "Method = PATCH (Status OK)",
			method:       http.MethodPatch,
			path:         "/tests/1",
			body:         `{"updates":{"location":"Updated location"}, "version":"0001-01-01T00:00:00Z"}`,
			expectedCode: http.StatusOK,
			setupMocks: func() {
				// Get existing test
				mock.ExpectQuery("SELECT \\* FROM tests WHERE id = \\$1;").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(testsColumns).
						AddRow(1, time.Now(), "Old Location", "Comment", 1, 1, 1, time.Time{}, false, 1))

				// Mock authentication
				AuthenticationMock(mock)

				// Begin transaction
				mock.ExpectBegin()

				// Update test with new location
				mock.ExpectQuery("UPDATE tests SET location = \\$1, version = \\$2 WHERE id = \\$3 AND version = \\$4 RETURNING version").
					WithArgs("Updated location", sqlmock.AnyArg(), 1, time.Time{}).
					WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(time.Now()))

				// Commit transaction
				mock.ExpectCommit()
			},
		},
		{
			name:         "Method = PUT (Status method not allowed)",
			method:       http.MethodPut,
			path:         "/tests/1",
			body:         "",
			expectedCode: http.StatusNotImplemented,
			setupMocks:   func() {},
		},
		{
			name:         "Method = DELETE (Status method not allowed)",
			method:       http.MethodDelete,
			path:         "/tests/1",
			body:         "",
			expectedCode: http.StatusNotImplemented,
			setupMocks:   func() {},
		},
		{
			name:         "Method = GET (Status bad request - start_date)",
			method:       http.MethodGet,
			path:         "/tests?start_date=2022-02-5&end_date=2024-02-07",
			body:         "",
			expectedCode: http.StatusBadRequest,
			setupMocks: func() {
				AuthenticationMock(mock)
			},
		},
		{
			name:         "Method = GET (Status bad request - end_date)",
			method:       http.MethodGet,
			path:         "/tests?start_date=2022-02-05&end_date=2024-02-",
			body:         "",
			expectedCode: http.StatusBadRequest,
			setupMocks: func() {
				AuthenticationMock(mock)
			},
		},
		{
			name:         "Method = PATCH (validation error)",
			method:       http.MethodPatch,
			path:         "/tests/1",
			body:         `{"updates":{"test_location":"Updated location"}, "version":"0001-01-01T00:00:00Z"}`,
			expectedCode: http.StatusBadRequest,
			setupMocks: func() {
				// Get existing test
				mock.ExpectQuery(`SELECT \* FROM tests WHERE id = \$1;`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(testsColumns).
						AddRow(1, time.Now(), "Old Location", "Comment", 1, 1, 1, time.Time{}, false, 1))
			},
		},
		{
			name:         "Method = GET (Status OK - all tests in date range)",
			method:       http.MethodGet,
			path:         "/tests?start_date=2022-02-05&end_date=2030-07-19",
			body:         ``,
			expectedCode: http.StatusOK,
			setupMocks: func() {
				AuthenticationMock(mock)

				// Setup query result rows
				rows := sqlmock.NewRows(testsColumns).
					AddRow(1, time.Now(), "Location 1", "Test Comment", 1, 1, 1, time.Now(), true, 1).
					AddRow(2, time.Now(), "Location 2", "Test Comment 2", 1, 1, 1, time.Now(), true, 1)

				// The expected query should match what the handler actually executes
				// Use the correct date range parameters
				mock.ExpectQuery("SELECT \\* FROM tests WHERE testing_team = \\$1 AND test_date >= to_date\\(\\$2, 'YYYY-MM-DD'\\) AND test_date <= to_date\\(\\$3, 'YYYY-MM-DD'\\)").
					WithArgs(1, "2022-02-05", "2030-07-19").
					WillReturnRows(rows)
			},
		},
		{
			name:         "Method = GET (Status OK - all public tests)",
			method:       http.MethodGet,
			path:         "/tests?public=true",
			body:         ``,
			expectedCode: http.StatusOK,
			setupMocks: func() {
				// Setup query result rows
				rows := sqlmock.NewRows(testsColumns).
					AddRow(1, time.Now(), "Location 1", "Test Comment", 1, 1, 1, time.Now(), true, 1).
					AddRow(2, time.Now(), "Location 2", "Test Comment 2", 2, 2, 2, time.Now(), true, 1)

				// Expect the query with the correct parameter name (is_public not publicly_available)
				mock.ExpectQuery("SELECT \\* FROM tests WHERE is_public = \\$1").
					WithArgs(true).
					WillReturnRows(rows)
			},
		},
		{
			name:         "Method = POST (Status internal server error - failed to commit transaction)",
			method:       http.MethodPost,
			path:         "/tests",
			body:         `{"sc":{"temperature":-10,"snow_type":"FS","snow_humidity":"W2"},"ac":{"temperature":-15,"humidity":30,"wind":"L","cloud":"1"},"tc":{"track_hardness":"H1","track_type":"D1"},"location":"Holmenkollen, Oslo","date":"2025-03-30T10:30:00Z","comment":"Test conducted under typical winter conditions. Excellent glide.","is_public":false,"testing_team":1,"test_ranks":[{"product_id":1,"rank":1,"distance_behind":0}]}`,
			expectedCode: http.StatusInternalServerError,
			setupMocks: func() {
				AuthenticationMock(mock)

				// Begin transaction
				mock.ExpectBegin().WillReturnError(sql.ErrNoRows)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			// Create the request with appropriate body
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")
			rr := httptest.NewRecorder()

			// Get the handler function from TestsHandler
			handler := TestsHandler(mockDB)

			// Call the handler function with the request and response writer
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			// Verify expectations
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func Test_createTestRankings(t *testing.T) {
	tests := []struct {
		name         string
		testID       int
		setupMocks   func(mock *sqlmock.Sqlmock)
		testRankings []TestRanksPOST
		wantErr      bool
	}{
		{
			name:   "Valid test rankings",
			testID: 1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				(*mock).ExpectBegin()

				// Expectations for the first product (ID 1)
				(*mock).ExpectQuery("SELECT is_public FROM products WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"is_public"}).AddRow(true))

				(*mock).ExpectExec("INSERT INTO test_ranks").
					WithArgs(1, 1, 1, 0, sqlmock.AnyArg(), true).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Expectations for the second product (ID 2)
				(*mock).ExpectQuery("SELECT is_public FROM products WHERE id = \\$1").
					WithArgs(2).
					WillReturnRows(sqlmock.NewRows([]string{"is_public"}).AddRow(true))

				(*mock).ExpectExec("INSERT INTO test_ranks").
					WithArgs(1, 2, 2, 5, sqlmock.AnyArg(), true).
					WillReturnResult(sqlmock.NewResult(2, 1))

				(*mock).ExpectCommit()
			},
			testRankings: []TestRanksPOST{
				{
					ProductID:      1,
					Rank:           1,
					DistanceBehind: 0,
					IsRankPublic:   true,
				},
				{
					ProductID:      2,
					Rank:           2,
					DistanceBehind: 5,
					IsRankPublic:   true,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock := utils.InitMockDB(t)

			// Setup mocks for this test before beginning the transaction
			tt.setupMocks(&mock)

			// Now begin the transaction
			mockTX, err := mockDB.Begin()
			if err != nil {
				t.Errorf("failed to begin transaction: %v", err)
			}

			if err = createTestRankings(mockTX, tt.testID, tt.testRankings); (err != nil) != tt.wantErr {
				t.Errorf("createTestsRankings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_insertTestRanking(t *testing.T) {
	tests := []struct {
		name       string
		tx         *sql.Tx
		testID     int
		setupMocks func(mock *sqlmock.Sqlmock)
		testRank   TestRanksPOST
		wantErr    bool
	}{
		{
			name:   "Valid test ranking",
			testID: 1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				(*mock).ExpectBegin()

				(*mock).ExpectQuery("SELECT is_public FROM products WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"is_public"}).AddRow(true))

				(*mock).ExpectExec("INSERT INTO test_ranks").
					WithArgs(1, 1, 1, 0, sqlmock.AnyArg(), true).
					WillReturnResult(sqlmock.NewResult(1, 1))

				(*mock).ExpectCommit()
			},
			testRank: TestRanksPOST{
				ProductID:      1,
				Rank:           1,
				DistanceBehind: 0,
				IsRankPublic:   true,
			},
			wantErr: false,
		},
		{
			name:   "Invalid test ranking",
			testID: 1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				(*mock).ExpectBegin()

				(*mock).ExpectExec("INSERT INTO test_ranks").
					WithArgs(1, 1, 1, 0, sqlmock.AnyArg(), false).
					WillReturnResult(sqlmock.NewResult(1, 1))

				(*mock).ExpectCommit().WillReturnError(sql.ErrNoRows)
			},
			testRank: TestRanksPOST{
				ProductID:      1,
				Rank:           1,
				DistanceBehind: 0,
				IsRankPublic:   false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock := utils.InitMockDB(t)

			// Setup mocks for this test before beginning the transaction
			tt.setupMocks(&mock)

			// Now begin the transaction
			mockTX, err := mockDB.Begin()
			if err != nil {
				t.Errorf("failed to begin transaction: %v", err)
			}

			if err = insertTestRanking(mockTX, tt.testID, tt.testRank); (err != nil) != tt.wantErr {
				t.Errorf("insertTestRanking() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_createTestUpdateQueries(t *testing.T) {
	tests := []struct {
		name                string
		path                string
		testUpdateRequest   TestPATCHRequest
		existingTestVersion time.Time
		testID              int
		productID           int
		setupMocks          func(mock *sqlmock.Sqlmock)
		want                time.Time
	}{
		{
			name: "Contains AC updates",
			path: "/tests/1",
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{
					"ac_temperature": 10,
				},
			},
			existingTestVersion: time.Time{},
			testID:              1,
			productID:           1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				// Begin transaction
				(*mock).ExpectBegin()

				// Get the air conditions ID
				(*mock).ExpectQuery("SELECT ac_id FROM tests WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"ac_id"}).AddRow(1))

				(*mock).ExpectExec("UPDATE air_conditions SET temperature = \\$1 WHERE id = \\$2").
					WithArgs(10, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Commit transaction
				(*mock).ExpectCommit()
			},
			want: time.Now(),
		},
		{
			name: "Contains TC updates",
			path: "/tests/1",
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{
					"track_hardness": "H1",
				},
			},
			existingTestVersion: time.Time{},
			testID:              1,
			productID:           1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				// Begin transaction
				(*mock).ExpectBegin()

				// Get the track conditions ID
				(*mock).ExpectQuery("SELECT tc_id FROM tests WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"tc_id"}).AddRow(1))

				(*mock).ExpectExec("UPDATE track_conditions SET track_hardness = \\$1 WHERE id = \\$2").
					WithArgs("H1", 1).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Commit transaction
				(*mock).ExpectCommit()
			},
			want: time.Now(),
		},
		{
			name: "Contains SC updates",
			path: "/tests/1",
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{
					"sc_temperature": 10,
				},
			},
			existingTestVersion: time.Time{},
			testID:              1,
			productID:           1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				// Begin transaction
				(*mock).ExpectBegin()

				// Get the snow conditions ID
				(*mock).ExpectQuery("SELECT sc_id FROM tests WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"sc_id"}).AddRow(1))

				(*mock).ExpectExec("UPDATE snow_conditions SET temperature = \\$1 WHERE id = \\$2").
					WithArgs(10, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Commit transaction
				(*mock).ExpectCommit()
			},
			want: time.Now(),
		},
		{
			name: "Contains test rank updates",
			path: "/tests/1/products/1",
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{
					"rank": 10,
				},
			},
			existingTestVersion: time.Time{},
			testID:              1,
			productID:           1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				// Begin transaction
				(*mock).ExpectBegin()

				(*mock).ExpectQuery("UPDATE test_ranks SET rank = \\$1, version = \\$2 WHERE test_id = \\$3 AND product_id = \\$4 RETURNING version").
					WithArgs(10, sqlmock.AnyArg(), 1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(time.Now()))

				// Commit transaction
				(*mock).ExpectCommit()
			},
			want: time.Now(),
		},
		{
			name: "Contains test information updates",
			path: "/tests/1",
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{
					"location": "New Location",
				},
			},
			existingTestVersion: time.Time{},
			testID:              1,
			productID:           1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				// Begin transaction
				(*mock).ExpectBegin()

				(*mock).ExpectQuery("UPDATE tests SET location = \\$1, version = \\$2 WHERE id = \\$3 AND version = \\$4 RETURNING version").
					WithArgs("New Location", sqlmock.AnyArg(), 1, time.Time{}).
					WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(time.Now()))

				// Commit transaction
				(*mock).ExpectCommit()
			},
			want: time.Now(),
		},
		{
			name: "Contains update attributes from multiple tables",
			path: "/tests/1/products/1",
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{
					"rank":           10,
					"ac_temperature": 10,
					"sc_temperature": 10,
				},
			},
			existingTestVersion: time.Time{},
			testID:              1,
			productID:           1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				// Begin transaction
				(*mock).ExpectBegin()

				// Get the air_conditions ID
				(*mock).ExpectQuery("SELECT ac_id FROM tests WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"ac_id"}).AddRow(1))

				// Update air conditions
				(*mock).ExpectExec("UPDATE air_conditions SET temperature = \\$1 WHERE id = \\$2").
					WithArgs(10, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Get the snow_conditions ID
				(*mock).ExpectQuery("SELECT sc_id FROM tests WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"sc_id"}).AddRow(1))

				// Update snow conditions
				(*mock).ExpectExec("UPDATE snow_conditions SET temperature = \\$1 WHERE id = \\$2").
					WithArgs(10, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Update test_ranks
				(*mock).ExpectQuery("UPDATE test_ranks SET rank = \\$1, version = \\$2 WHERE test_id = \\$3 AND product_id = \\$4 RETURNING version").
					WithArgs(10, sqlmock.AnyArg(), 1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(time.Now()))

				// Commit transaction
				(*mock).ExpectCommit()
			},
			want: time.Now(),
		},
		{
			name: "Could not start transaction",
			path: "/tests/1",
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{
					"ac_temperature": 10,
				},
			},
			existingTestVersion: time.Time{},
			testID:              1,
			productID:           1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
			},
			want: time.Time{},
		},
		{
			name: "AC conflict",
			path: "/tests/1",
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{
					"ac_temperature": 10,
				},
			},
			existingTestVersion: time.Time{},
			testID:              1,
			productID:           1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				// Begin transaction
				(*mock).ExpectBegin()

				// Get the air conditions ID
				(*mock).ExpectQuery("SELECT ac_id FROM tests WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"ac_id"}).AddRow(1))

				(*mock).ExpectExec("UPDATE air_conditions SET temperature = \\$1 WHERE id = \\$2").
					WithArgs(10, 1).
					WillReturnError(sql.ErrNoRows)

				// Commit transaction
				(*mock).ExpectCommit()
			},
			want: time.Time{},
		},
		{
			name: "TC conflict",
			path: "/tests/1",
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{
					"track_hardness": "H1",
				},
			},
			existingTestVersion: time.Time{},
			testID:              1,
			productID:           1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				// Begin transaction
				(*mock).ExpectBegin()

				// Get the track conditions ID
				(*mock).ExpectQuery("SELECT tc_id FROM tests WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"tc_id"}).AddRow(1))

				(*mock).ExpectExec("UPDATE track_conditions SET track_hardness = \\$1 WHERE id = \\$2").
					WithArgs("H1", 1).
					WillReturnError(sql.ErrNoRows)
			},
			want: time.Time{},
		},
		{
			name: "SC conflict",
			path: "/tests/1",
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{
					"sc_temperature": 10,
				},
			},
			existingTestVersion: time.Time{},
			testID:              1,
			productID:           1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				// Begin transaction
				(*mock).ExpectBegin()

				// Get the snow conditions ID
				(*mock).ExpectQuery("SELECT sc_id FROM tests WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"sc_id"}).AddRow(1))

				(*mock).ExpectExec("UPDATE snow_conditions SET temperature = \\$1 WHERE id = \\$2").
					WithArgs(10, 1).
					WillReturnError(sql.ErrNoRows)
			},
			want: time.Time{},
		},
		{
			name: "Test rank conflict",
			path: "/tests/1/products/1",
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{
					"rank": 10,
				},
			},
			existingTestVersion: time.Time{},
			testID:              1,
			productID:           1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				// Begin transaction
				(*mock).ExpectBegin()

				(*mock).ExpectQuery("UPDATE test_ranks SET rank = \\$1, version = \\$2 WHERE test_id = \\$3 AND product_id = \\$4 RETURNING version").
					WithArgs(10, sqlmock.AnyArg(), 1, 1).
					WillReturnError(sql.ErrNoRows)
			},
			want: time.Time{},
		},
		{
			name: "Test information conflict",
			path: "/tests/1",
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{
					"location": "New Location",
				},
			},
			existingTestVersion: time.Time{},
			testID:              1,
			productID:           1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				// Begin transaction
				(*mock).ExpectBegin()

				(*mock).ExpectQuery("UPDATE tests SET location = \\$1, version = \\$2 WHERE id = \\$3 AND version = \\$4 RETURNING version").
					WithArgs("New Location", sqlmock.AnyArg(), 1, time.Time{}).
					WillReturnError(sql.ErrNoRows)
			},
			want: time.Time{},
		},
		{
			name: "Could not find AC id",
			path: "/tests/1",
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{
					"ac_temperature": 10,
				},
			},
			existingTestVersion: time.Time{},
			testID:              1,
			productID:           1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				// Begin transaction
				(*mock).ExpectBegin()

				// Get the air conditions ID
				(*mock).ExpectQuery("SELECT ac_id FROM tests WHERE id = \\$1").
					WithArgs(1).
					WillReturnError(sql.ErrNoRows)
			},
			want: time.Time{},
		},
		{
			name: "Could not find TC id",
			path: "/tests/1",
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{
					"track_hardness": "H1",
				},
			},
			existingTestVersion: time.Time{},
			testID:              1,
			productID:           1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				// Begin transaction
				(*mock).ExpectBegin()

				// Get the track conditions ID
				(*mock).ExpectQuery("SELECT tc_id FROM tests WHERE id = \\$1").
					WithArgs(1).
					WillReturnError(sql.ErrNoRows)
			},
			want: time.Time{},
		},
		{
			name: "Could not find SC id",
			path: "/tests/1",
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{},
			},
			existingTestVersion: time.Time{},
			testID:              1,
			productID:           1,
			setupMocks: func(mock *sqlmock.Sqlmock) {
				// Begin transaction
				(*mock).ExpectBegin()
			},
			want: time.Time{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock := utils.InitMockDB(t)
			// Setup mocks for this test before beginning the transaction
			tt.setupMocks(&mock)

			// Create a mock HTTP request and response recorder
			req := httptest.NewRequest(http.MethodPatch, tt.path, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")
			rr := httptest.NewRecorder()

			formatedWant := time.Date(tt.want.Year(), tt.want.Month(), tt.want.Day(), tt.want.Hour(), tt.want.Minute(), tt.want.Second(), tt.want.Nanosecond(), tt.want.Location())

			returnedTime := createTestUpdateQueries(rr, req, mockDB,
				tt.testUpdateRequest, tt.existingTestVersion, tt.testID, tt.productID)
			// Truncate the time to seconds precision for reliable comparison
			formatedReturnedTime := time.Date(returnedTime.Year(), returnedTime.Month(), returnedTime.Day(), returnedTime.Hour(), returnedTime.Minute(), returnedTime.Second(), tt.want.Nanosecond(), returnedTime.Location())
			assert.Equalf(t, formatedWant, formatedReturnedTime,
				"createTestUpdateQueries(%v, %v, %v, %v, %v, %v, %v)",
				rr, req, mockDB, tt.testUpdateRequest, tt.existingTestVersion, tt.testID, tt.productID)

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func Test_mapUpdatedTestAndRankingsFields(t *testing.T) {
	tests := []struct {
		name          string
		updatedFields []string
		newValues     []interface{}
		validFields   map[string]bool
		want          []string
		want1         int
	}{
		{
			name:          "Valid fields",
			updatedFields: []string{"rank = 10", "location = 'Lillehammer'"},
			newValues:     []interface{}{10, "Lillehammer"},
			validFields:   map[string]bool{"rank": true, "location": true},
			want:          []string{"rank = $1", "location = $2", "version = $3"},
			want1:         4,
		},
		{
			name:          "Invalid fields",
			updatedFields: []string{"air_temperature", "snow_humidity"},
			newValues:     []interface{}{10.0, "DS"},
			validFields:   map[string]bool{"air_temperature": true, "snow_humidity": true},
			want:          []string(nil),
			want1:         0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, _ := mapUpdatedTestAndRankingsFields(tt.updatedFields, tt.newValues, tt.validFields)
			assert.Equalf(t, tt.want, got, "mapUpdatedTestAndRankingsFields(%v, %v, %v)", tt.updatedFields, tt.newValues, tt.validFields)
			assert.Equalf(t, tt.want1, got1, "mapUpdatedTestAndRankingsFields(%v, %v, %v)", tt.updatedFields, tt.newValues, tt.validFields)
		})
	}
}

func Test_mapUpdatedTestConditionFields(t *testing.T) {
	tests := []struct {
		name          string
		updatedFields []string
		newValues     []interface{}
		validFields   map[string]bool
		actualNames   map[string]string
		want          []string
		want1         int
		want2         []interface{}
	}{
		{
			name:          "Valid fields",
			updatedFields: []string{"ac_temperature = 10", "snow_humidity = 'DS'"},
			newValues:     []interface{}{10, "DS"},
			validFields:   map[string]bool{"ac_temperature": true, "snow_humidity": true},
			actualNames:   map[string]string{"ac_temperature": "temperature", "snow_humidity": "humidity"},
			want:          []string{"temperature = $1", "humidity = $2"},
			want1:         3,
			want2:         []interface{}{10, "DS"},
		},
		{
			name:          "Invalid fields",
			updatedFields: []string{"air_temperature", "snow_humidity"},
			newValues:     []interface{}{10.0, "DS"},
			validFields:   map[string]bool{"air_temperature": true, "snow_humidity": true},
			actualNames:   map[string]string{"air_temperature": "temperature", "snow_humidity": "humidity"},
			want:          []string(nil),
			want1:         0,
			want2:         []interface{}(nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := mapUpdatedTestConditionFields(tt.updatedFields, tt.newValues, tt.validFields, tt.actualNames)
			assert.Equalf(t, tt.want, got, "mapUpdatedTestConditionFields(%v, %v, %v, %v)", tt.updatedFields, tt.newValues, tt.validFields, tt.actualNames)
			assert.Equalf(t, tt.want1, got1, "mapUpdatedTestConditionFields(%v, %v, %v, %v)", tt.updatedFields, tt.newValues, tt.validFields, tt.actualNames)
			assert.Equalf(t, tt.want2, got2, "mapUpdatedTestConditionFields(%v, %v, %v, %v)", tt.updatedFields, tt.newValues, tt.validFields, tt.actualNames)
		})
	}
}

func TestValidateTestPATCHRequestBody(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	testTime := time.Now()

	tests := []struct {
		name              string
		db                *sql.DB
		testUpdateRequest TestPATCHRequest
		existingTest      domain.Test
		setupMocks        func()
		want              error
		code              int
	}{
		{
			name: "Successful validation",
			db:   mockDB,
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{"location": "Updated location"},
			},
			existingTest: domain.Test{
				ID:              1,
				Date:            time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				Location:        "Test Location",
				Comment:         "Test Comment",
				SnowConditions:  1,
				TrackConditions: 1,
				AirConditions:   1,
				Version:         testTime,
				IsPublic:        false,
				TestingTeam:     1,
			},
			setupMocks: func() {
				// Mock the user id query
				mock.ExpectQuery("SELECT user_id FROM sessions WHERE \\(session_token = \\$1 AND expires_at > NOW\\(\\)\\)").
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				// Mock the user team id query
				mock.ExpectQuery("SELECT team_id FROM users WHERE id = \\$1;").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(1))

				// Mock the user team role query
				mock.ExpectQuery("SELECT team_role FROM team WHERE id = \\$1;").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"team_role"}).AddRow(1))
			},
			want: nil,
			code: 0,
		},
		{
			name: "Invalid PATCH request body keys",
			db:   mockDB,
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{"invalid_field": "value"},
			},
			existingTest: domain.Test{
				ID:          1,
				TestingTeam: 1,
				IsPublic:    false,
			},
			setupMocks: func() {
				// No mocks needed for this test case
			},
			want: fmt.Errorf("invalid PATCH request body keys, %d", http.StatusBadRequest),
			code: http.StatusBadRequest,
		},
		{
			name: "Invalid PATCH request body values",
			db:   mockDB,
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{"ac_temperature": 190},
			},
			existingTest: domain.Test{
				ID:          1,
				TestingTeam: 1,
				IsPublic:    false,
			},
			setupMocks: func() {},
			want:       fmt.Errorf("invalid PATCH request body values, %d", http.StatusBadRequest),
			code:       http.StatusBadRequest,
		},
		{
			name: "Could not decode PATCH request body",
			db:   mockDB,
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{"is_public": "not_a_boolean"},
			},
			existingTest: domain.Test{
				ID:          1,
				TestingTeam: 1,
				IsPublic:    false,
			},
			setupMocks: func() {
			},
			want: fmt.Errorf("could not decode request body, %d", http.StatusInternalServerError),
			code: http.StatusInternalServerError,
		},
		{
			name: "User cannot update this test",
			db:   mockDB,
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{"location": "Updated location"},
			},
			existingTest: domain.Test{
				ID:          1,
				TestingTeam: 2, // Different team than user's team
				IsPublic:    false,
			},
			setupMocks: func() {
				// Mock session token lookup
				mock.ExpectQuery(`SELECT user_id FROM sessions WHERE \(session_token = \$1 AND expires_at > NOW\(\)\)`).
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				// Mock getting user team
				mock.ExpectQuery(`SELECT team_id FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(1))
			},
			want: fmt.Errorf("user cannot update this test, %d", http.StatusUnauthorized),
			code: http.StatusUnauthorized,
		},
		{
			name: "Researcher cannot update public tests",
			db:   mockDB,
			testUpdateRequest: TestPATCHRequest{
				Updates: map[string]interface{}{"location": "Updated location"},
			},
			existingTest: domain.Test{
				ID:          2,
				TestingTeam: 2,
				IsPublic:    true,
			},
			setupMocks: func() {
				// Mock the user id query
				mock.ExpectQuery("SELECT user_id FROM sessions WHERE \\(session_token = \\$1 AND expires_at > NOW\\(\\)\\)").
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				// Mock the user team id query
				mock.ExpectQuery("SELECT team_id FROM users WHERE id = \\$1;").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(1))

				// Mock the user team role query
				mock.ExpectQuery("SELECT team_role FROM team WHERE id = \\$1;").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"team_role"}).AddRow(2))
			},
			want: fmt.Errorf("researcher cannot update public tests, %d", http.StatusUnauthorized),
			code: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock expectations
			tt.setupMocks()

			// Create a mock HTTP request and response recorder
			req := httptest.NewRequest(http.MethodPatch, "/tests/1", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")
			rr := httptest.NewRecorder()

			got, got1 := ValidateTestPATCHRequestBody(rr, req, tt.db, tt.testUpdateRequest, tt.existingTest)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.code, got1)

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}
