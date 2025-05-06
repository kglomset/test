package productsHandler

import (
	"backend/internal/domain"
	"backend/internal/resources"
	"backend/internal/utils"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func resetTime(products []domain.Product) []domain.Product {
	for i := range products {
		products[i].Version = time.Time{}
	}
	return products
}

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

var productColumns = []string{"id", "name", "brand", "ean_code", "image_url", "comment", "is_public",
	"type", "high_temperature", "low_temperature", "testing_team", "version", "status"}

var productsArray = []domain.Product{
	{ID: 1, Name: "Product1", Brand: "Brand1", EANCode: "1234567890123", ImageURL: "", Type: "Type1",
		HighTemperature: 1.0, LowTemperature: -1.0, Comment: "Comment1", TestingTeam: 1, IsPublic: false,
		Version: time.Time{}, Status: "Status1"},
	{ID: 2, Name: "Product2", Brand: "Brand2", EANCode: "1234567890124", ImageURL: "", Type: "Type2",
		HighTemperature: 2.0, LowTemperature: -2.0, Comment: "Comment2", TestingTeam: 2, IsPublic: true,
		Version: time.Time{}, Status: "Status2"},
}

var testProduct = domain.Product{
	ID:              1,
	Name:            "Product1",
	Brand:           "Brand1",
	EANCode:         "1234567890123",
	ImageURL:        "image_url",
	Comment:         "Comment1",
	IsPublic:        false,
	Type:            "Type1",
	HighTemperature: 1.0,
	LowTemperature:  -1.0,
	TestingTeam:     1,
	Version:         time.Time{},
	Status:          "Status1",
}

var testProductInsert = ProductPOSTRequest{
	Name:            "Product1",
	Brand:           "Brand1",
	EANCode:         "1234567890123",
	ImageURL:        "image_url",
	Comment:         "Comment1",
	IsPublic:        false,
	Type:            "bundle",
	HighTemperature: 1.0,
	LowTemperature:  -1.0,
	TestingTeam:     1,
	Status:          "active",
}

func TestDirectReferenceUpdateProductAppearances(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)

	tests := []struct {
		name       string
		id         int
		eanCode    string
		setupMocks func()
		wantStatus int
		wantBody   string
	}{
		{
			name:    "Successful update",
			id:      1,
			eanCode: "1234567890123",
			setupMocks: func() {
				AuthenticationMock(mock)
				// Mock the private product query
				mock.ExpectQuery("SELECT \\* FROM products WHERE id = \\$1 AND ean_code = \\$2 AND is_public = \\$3 AND testing_team = \\$4;").
					WithArgs(1, "1234567890123", false, 1).
					WillReturnRows(sqlmock.NewRows(productColumns).
						AddRow(1, "Product1", "Brand1", "1234567890123", "", "Comment1", false,
							"Type1", 1.0, 1.0, 1, time.Time{}, "Status1"))

				// Mock the public product query
				mock.ExpectQuery("SELECT \\* FROM products WHERE ean_code = \\$1 AND is_public = \\$2;").
					WithArgs("1234567890123", true).
					WillReturnRows(sqlmock.NewRows(productColumns).
						AddRow(1, "Product1", "Brand1", "1234567890123", "", "Comment1", false,
							"Type1", 1.0, 1.0, 1, time.Time{}, "Status1"))

				// Mock the update rankings query
				mock.ExpectExec("UPDATE test_ranks SET product_id = \\$1 WHERE product_id = \\$2;").
					WithArgs(1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Mock the delete private product query
				mock.ExpectExec("DELETE FROM products WHERE id = \\$1 AND testing_team = \\$2;").
					WithArgs(1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantStatus: http.StatusOK,
			wantBody:   "",
		},
		{
			name:    "Could not retrieve the private product",
			id:      1,
			eanCode: "1234567890123",
			setupMocks: func() {
				AuthenticationMock(mock)

				// Mock the private product query
				mock.ExpectQuery("SELECT \\* FROM products WHERE id = \\$1 AND ean_code = \\$2 AND is_public = \\$3 AND testing_team = \\$4;").
					WithArgs(1, "1234567890123", false, 1).
					WillReturnError(sql.ErrNoRows)
			},
			wantStatus: http.StatusNotFound,
			wantBody:   "Could not retrieve the product",
		},
		{
			name:    "Could not retrieve the public product",
			id:      1,
			eanCode: "1234567890123",
			setupMocks: func() {
				AuthenticationMock(mock)

				// Mock the private product query
				mock.ExpectQuery("SELECT \\* FROM products WHERE id = \\$1 AND ean_code = \\$2 AND is_public = \\$3 AND testing_team = \\$4;").
					WithArgs(1, "1234567890123", false, 1).
					WillReturnRows(sqlmock.NewRows(productColumns).
						AddRow(1, "Product1", "Brand1", "1234567890123", "", "Comment1", false,
							"Type1", 1.0, 1.0, 1, time.Time{}, "Status1"))

				// Mock the public product query to return no rows
				mock.ExpectQuery("SELECT \\* FROM products WHERE ean_code = \\$1 AND is_public = \\$2;").
					WithArgs("1234567890123", true).
					WillReturnError(sql.ErrNoRows)
			},
			wantStatus: http.StatusNotFound,
			wantBody:   "Could not retrieve the product",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock expectations
			tt.setupMocks()

			// Create a mock HTTP request and response recorder
			req := httptest.NewRequest(http.MethodPatch, "/products", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")
			rr := httptest.NewRecorder()

			// Call the function with the mock objects
			DirectReferenceUpdateProductAppearances(rr, req, mockDB, tt.eanCode, tt.id)

			assert.Equal(t, rr.Code, tt.wantStatus)
			assert.Contains(t, rr.Body.String(), tt.wantBody)

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestFetchProducts(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)

	tests := []struct {
		name       string
		products   []domain.Product
		setupMocks func()
		wantStatus int
		wantBody   string
	}{
		{
			name:     "Successful product retrieval",
			products: productsArray,
			setupMocks: func() {
				// Set up expectation for the database query
				mock.ExpectQuery("SELECT \\* FROM products").
					WillReturnRows(sqlmock.NewRows(productColumns).
						AddRow(1, "Product1", "Brand1", "1234567890123", "", "Comment1", false,
							"Type1", 1.0, 1.0, 1, time.Time{}, "Status1").
						AddRow(2, "Product2", "Brand2", "1234567890124", "", "Comment2", true,
							"Type2", 2.0, 2.0, 2, time.Time{}, "Status2"))
			},
			wantStatus: http.StatusOK,
			wantBody:   "",
		},
		{
			name:     "No products found",
			products: []domain.Product{},
			setupMocks: func() {
				// Set up expectation for the database query
				mock.ExpectQuery("SELECT \\* FROM products").
					WillReturnRows(sqlmock.NewRows(productColumns))
			},
			wantStatus: http.StatusOK,
			wantBody:   "No products found.",
		},
		{
			name:     "Could not retrieve all products",
			products: []domain.Product{},
			setupMocks: func() {
				// Set up expectation for the database query
				mock.ExpectQuery("SELECT \\* FROM products").
					WillReturnRows(sqlmock.NewRows(productColumns).
						AddRow("", "", "", "", "", "", false, "", 0.0, 0.0, 0, time.Time{}, ""))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Could not retrieve all products.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock expectations
			tt.setupMocks()

			// Create a mock HTTP request and response recorder
			req := httptest.NewRequest(http.MethodGet, "/products", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")
			rr := httptest.NewRecorder()

			// Execute mock query
			mockedRows, err := mockDB.Query("SELECT * FROM products")
			assert.NoError(t, err)

			// Call the function with the mock objects
			FetchProducts(rr, tt.products, mockedRows, err)

			assert.Equal(t, rr.Code, tt.wantStatus)
			assert.Contains(t, rr.Body.String(), tt.wantBody)

			// Ensure all expectations were met
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestGetProductWithEANCode(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name        string
		eanCode     string
		productName string
		team        int
		setupMocks  func()
		want        domain.Product
		code        int
	}{
		{
			name:        "Successful product retrieval by EAN code",
			eanCode:     "1234567890123",
			productName: "Product1",
			team:        1,
			setupMocks: func() {
				// Mock the product retrieval query
				mock.ExpectQuery("SELECT \\* FROM products WHERE testing_team = \\$1 AND ean_code = \\$2;").
					WithArgs(1, "1234567890123").
					WillReturnRows(sqlmock.NewRows(productColumns).
						AddRow(1, "Product1", "Brand1", "1234567890123", "image_url", "Comment1", false,
							"Type1", 1.0, -1.0, 1, time.Time{}, "Status1"))
			},
			want: testProduct,
			code: http.StatusConflict,
		},
		{
			name:        "Product lookup failed ",
			eanCode:     "1234567890123",
			productName: "TestProduct",
			team:        1,
			setupMocks: func() {
				// Mock the product retrieval query to return no rows
				mock.ExpectQuery("SELECT \\* FROM products WHERE testing_team = \\$1 AND ean_code = \\$2;").
					WithArgs(1, "1234567890123").
					WillReturnError(sql.ErrNoRows)
			},
			want: domain.Product{},
			code: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock expectations
			tt.setupMocks()

			// Create a mock HTTP request and response recorder
			req := httptest.NewRequest(http.MethodGet, "/products", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")

			got, got1 := GetProductWithEANCode(mockDB, tt.eanCode, tt.productName, tt.team)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetProductWithEANCode() got = %v, want %v", got, tt.want)
			}

			assert.Equal(t, tt.code, got1)

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestGetProductWithID(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name       string
		productID  int
		setupMocks func()
		want       domain.Product
	}{
		{
			name:      "Successful product retrieval by ID",
			productID: 1,
			setupMocks: func() {
				mock.ExpectQuery("SELECT \\* FROM products WHERE id = \\$1;").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "brand", "ean_code", "image_url", "comment", "is_public", "type", "high_temperature", "low_temperature", "testing_team", "version", "status"}).
						AddRow(1, "Product1", "Brand1", "1234567890123", "image_url", "Comment1", false, "Type1", 1.0, 1.0, 1, time.Time{}, "Status1"))
			},
			want: domain.Product{
				ID:              1,
				Name:            "Product1",
				Brand:           "Brand1",
				EANCode:         "1234567890123",
				ImageURL:        "image_url",
				Comment:         "Comment1",
				IsPublic:        false,
				Type:            "Type1",
				HighTemperature: 1.0,
				LowTemperature:  1.0,
				TestingTeam:     1,
				Version:         time.Time{},
				Status:          "Status1",
			},
		},
		{
			name:      "Could not retrieve the product",
			productID: 1,
			setupMocks: func() {
				mock.ExpectQuery("SELECT \\* FROM products WHERE id = \\$1;").
					WithArgs(1).
					WillReturnError(sql.ErrNoRows)
			},
			want: domain.Product{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock expectations
			tt.setupMocks()

			// Create a mock HTTP request and response recorder
			req := httptest.NewRequest(http.MethodGet, "/products", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")
			rr := httptest.NewRecorder()

			if got := GetProductWithID(rr, mockDB, tt.productID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetProductWithID() = %v, want %v", got, tt.want)
			}

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestGetProductFields(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name       string
		path       string
		fields     []string
		productID  int
		team       int
		setupMocks func()
		wantBody   map[string]interface{}
		wantCode   int
	}{
		{
			name:      "Successful product field retrieval (one field)",
			path:      "/products/1/?fields=name",
			fields:    []string{"name"},
			productID: 1,
			team:      1,
			setupMocks: func() {
				// Mock the product retrieval query
				mock.ExpectQuery("SELECT name FROM products WHERE id = \\$1 AND testing_team = \\$2;").
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Product1"))
			},
			wantBody: map[string]interface{}{
				"name": "Product1",
			},
			wantCode: http.StatusOK,
		},
		{
			name:      "Successful product field retrieval (multiple fields)",
			path:      "/products/1/?fields=id,name,type",
			fields:    []string{"id", "name", "type"},
			productID: 1,
			team:      1,
			setupMocks: func() {
				// Mock the product retrieval query
				mock.ExpectQuery("SELECT id, name, type FROM products WHERE id = \\$1 AND testing_team = \\$2;").
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "type"}).AddRow(1, "Product1", "gel"))
			},
			wantBody: map[string]interface{}{
				"id":   1,
				"name": "Product1",
				"type": "gel",
			},
			wantCode: http.StatusOK,
		},
		{
			name:      "Invalid field name",
			path:      "/products/1/?fields=brand124123",
			fields:    []string{"brand124123"},
			productID: 1,
			team:      1,
			setupMocks: func() {

			},
			wantBody: nil,
			wantCode: http.StatusBadRequest,
		},
		{
			name:      "Could not retrieve product fields",
			path:      "/products/1/?fields=high_temperature,low_temperature",
			fields:    []string{"high_temperature", "low_temperature"},
			productID: 1,
			team:      1,
			setupMocks: func() {
				mock.ExpectQuery("SELECT high_temperature, low_temperature FROM products WHERE id = \\$1 AND testing_team = \\$2;").
					WithArgs(1, 1).
					WillReturnError(sql.ErrNoRows)
			},
			wantCode: http.StatusNotFound,
		},
		{
			name:      "Could not encode product fields",
			path:      "/products/1/?fields=name",
			fields:    []string{"name"},
			productID: 1,
			team:      1,
			setupMocks: func() {
				// Mock the product retrieval query
				mock.ExpectQuery("SELECT name FROM products WHERE id = \\$1 AND testing_team = \\$2;").
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow(true)) // Invalid type for name
			},
			wantBody: nil,
			wantCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock expectations
			tt.setupMocks()

			// Create a mock HTTP request and response recorder
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")
			rr := httptest.NewRecorder()

			query := fmt.Sprintf(`SELECT %s FROM products WHERE id = $1 AND testing_team = $2;`,
				strings.Join(tt.fields, ", "))

			GetProductFields(rr, mockDB, query, tt.fields, tt.productID, tt.team)

			/*
				// Check the response body contains the expected error message
				if !strings.Contains(rr.Body.String(), tt.wantBody) {
					t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), tt.wantBody)
				}*/

			if tt.name == "Successful product field retrieval (one field)" ||
				tt.name == "Successful product field retrieval (multiple fields)" {
				// Parse the response body
				var responseBody map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &responseBody); err != nil {
					t.Errorf("couldn't parse response body: %v", err)
					return
				}
			}

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func Test_isProductUnique(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name                 string
		db                   *sql.DB
		productUpdateRequest ProductPATCHRequest
		team                 int
		setupMocks           func()
		want                 bool
	}{
		{
			name: "Product is not unique",
			db:   mockDB,
			productUpdateRequest: ProductPATCHRequest{
				Updates: map[string]interface{}{"name": "Product1"},
				Version: time.Time{},
			},
			setupMocks: func() {
				// Add a mock product to the database
				mock.ExpectQuery("SELECT \\* FROM products").
					WillReturnRows(sqlmock.NewRows(productColumns).
						AddRow(1, "Product1", "Brand1", "1234567890123", "", "Comment1", false,
							"Type1", 1.0, 1.0, 1, time.Time{}, "Status1").
						AddRow(2, "Product2", "Brand2", "1234567890124", "", "Comment2", true,
							"Type2", 2.0, 2.0, 2, time.Time{}, "Status2"))
				// Mock that the product is not unique
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products WHERE name = \\$1 AND testing_team = \\$2;").
					WithArgs("Product1", 1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			team: 1,
			want: false,
		},
		{
			name: "Product is unique",
			db:   mockDB,
			productUpdateRequest: ProductPATCHRequest{
				Updates: map[string]interface{}{"name": "Product1"},
				Version: time.Time{},
			},
			setupMocks: func() {
				// Mock the product uniqueness check
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products WHERE name = \\$1 AND testing_team = \\$2;").
					WithArgs("Product1", 1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			team: 1,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			assert.Equalf(t, tt.want, isProductUnique(tt.db, tt.productUpdateRequest, tt.team),
				"isProductUnique(%v, %v, %v)", tt.db, tt.productUpdateRequest, tt.team)
		})
	}
}

func TestProductsHandler(t *testing.T) {
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
			name:         "Method = GET (Status OK)",
			method:       http.MethodGet,
			path:         "/products",
			body:         "",
			expectedCode: http.StatusOK,
			setupMocks: func() {
				AuthenticationMock(mock)

				mock.ExpectQuery("SELECT \\* FROM products").
					WillReturnRows(sqlmock.NewRows(productColumns))
			},
		},
		{
			name:         "Method = POST (Status OK)",
			method:       http.MethodPost,
			path:         "/products",
			body:         `{"name":"Product1","brand":"Brand1","ean_code":"1234567890123","image_url":"","comment":"Comment1","is_public":false,"type":"bundle","high_temperature":1.0,"low_temperature":-1.0,"testing_team":1,"status":"active"}`,
			expectedCode: http.StatusCreated,
			setupMocks: func() {
				AuthenticationMock(mock)

				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products;").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

				// Mock the product retrieval query to return no rows
				mock.ExpectQuery("SELECT \\* FROM products WHERE testing_team = \\$1 AND ean_code = \\$2;").
					WithArgs(1, "1234567890123").
					WillReturnError(sql.ErrNoRows)

				mock.ExpectExec("INSERT INTO products").
					WithArgs("Product1", "Brand1", "1234567890123", "", "Comment1", false,
						"bundle", 1.0, -1.0, 1, sqlmock.AnyArg(), "active").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name:         "Method = PATCH (Status OK)",
			method:       http.MethodPatch,
			path:         "/products/1",
			body:         `{"updates":{"name":"Updated product"}, "version":"0001-01-01T00:00:00Z"}`,
			expectedCode: http.StatusOK,
			setupMocks: func() {
				AuthenticationMock(mock)

				mock.ExpectQuery("SELECT \\* FROM products WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(productColumns).
						AddRow(1, "Product1", "Brand1", "1234567890123", "", "Comment1", false,
							"Type1", 1.0, -1.0, 1, time.Time{}, "Status1"))

				mock.ExpectQuery("SELECT \\* FROM products WHERE testing_team = \\$1 AND name = \\$2;").
					WithArgs(1, "Updated product").
					WillReturnRows(sqlmock.NewRows(productColumns))

				mock.ExpectQuery("UPDATE products SET name = \\$1, version = \\$2 WHERE id = \\$3 AND version = \\$4 RETURNING version").
					WithArgs("Updated product", sqlmock.AnyArg(), 1, time.Time{}).
					WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(time.Now()))
			},
		},
		{
			name:         "Method = PUT (Status method not allowed)",
			method:       http.MethodPut,
			path:         "/products/1",
			body:         "",
			expectedCode: http.StatusNotImplemented,
			setupMocks:   func() {},
		},
		{
			name:         "Method = DELETE (Status method not allowed)",
			method:       http.MethodDelete,
			path:         "/products/1",
			body:         "",
			expectedCode: http.StatusNotImplemented,
			setupMocks:   func() {},
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

			// Get the handler function from ProductsHandler
			handler := ProductsHandler(mockDB)

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

func TestProductsRequestGET(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)

	tests := []struct {
		name         string
		path         string
		setupMocks   func()
		expectedCode int
		expectedBody string
	}{
		{
			name: "Get all products",
			path: "/products",
			setupMocks: func() {
				AuthenticationMock(mock)

				mock.ExpectQuery("SELECT \\* FROM products").
					WillReturnRows(sqlmock.NewRows(productColumns).
						AddRow(1, "Product1", "Brand1", "1234567890123", "", "Comment1", false,
							"Type1", 1.0, 1.0, 1, time.Time{}, "Status1").
						AddRow(2, "Product2", "Brand2", "1234567890124", "", "Comment2", true,
							"Type2", 2.0, 2.0, 2, time.Time{}, "Status2"))
			},
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name: "Get product by ID",
			path: "/products/1",
			setupMocks: func() {
				AuthenticationMock(mock)

				mock.ExpectQuery("SELECT \\* FROM products WHERE id = \\$1 AND testing_team = \\$2;").
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows(productColumns).
						AddRow(1, "Product1", "Brand1", "1234567890123", "", "Comment1", false,
							"Type1", 1.0, 1.0, 1, time.Time{}, "Status1"))
			},
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name: "Get public products",
			path: "/products?public=true",
			setupMocks: func() {
				mock.ExpectQuery("SELECT \\* FROM products WHERE is_public = \\$1;").
					WithArgs(true).
					WillReturnRows(sqlmock.NewRows(productColumns).
						AddRow(1, "Product1", "Brand1", "1234567890123", "", "Comment1", true,
							"Type1", 1.0, 1.0, 1, time.Time{}, "Status1").
						AddRow(2, "Product2", "Brand2", "1234567890124", "", "Comment2", true,
							"Type2", 2.0, 2.0, 2, time.Time{}, "Status2"))
			},
			expectedCode: http.StatusOK,
			expectedBody: "Product2",
		},
		{
			name: "Get product fields",
			path: "/products/1?fields=high_temperature,low_temperature",
			setupMocks: func() {
				AuthenticationMock(mock)

				mock.ExpectQuery("SELECT high_temperature, low_temperature FROM products WHERE id = \\$1 AND testing_team = \\$2").
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"high_temperature", "low_temperature"}).
						AddRow(20, -10))
			},
			expectedCode: http.StatusOK,
			expectedBody: "{\"high_temperature\":20,\"low_temperature\":-10}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")
			rr := httptest.NewRecorder()

			ProductsRequestGET(rr, req, mockDB)

			assert.Equal(t, tt.expectedCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestProductsRequestPATCH(t *testing.T) {
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
			name: "Update product name",
			path: "/products/1",
			body: `{"updates":{"name":"Updated product"}, "version":"0001-01-01T00:00:00Z"}`,
			setupMocks: func() {
				AuthenticationMock(mock)

				mock.ExpectQuery("SELECT \\* FROM products WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(productColumns).
						AddRow(1, "Product1", "Brand1", "1234567890123", "", "Comment1", false,
							"Type1", 1.0, 1.0, 1, time.Time{}, "Status1"))

				mock.ExpectQuery("SELECT \\* FROM products WHERE testing_team = \\$1 AND name = \\$2;").
					WithArgs(1, "Updated product").
					WillReturnError(sql.ErrNoRows)

				mock.ExpectQuery("UPDATE products SET name = \\$1, version = \\$2 WHERE id = \\$3 AND version = \\$4 RETURNING version").
					WithArgs("Updated product", sqlmock.AnyArg(), 1, time.Time{}).
					WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(time.Now()))
			},
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name: "Invalid PATCH request body",
			path: "/products/1",
			body: `{"updates":{"name": "Invalid JSON"`,
			setupMocks: func() {
				AuthenticationMock(mock)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: resources.InvalidPATCHRequest,
		},
		{
			name: "Product not found",
			path: "/products/2000",
			body: `{"updates":{"name":"Updated Product"}}`,
			setupMocks: func() {
				AuthenticationMock(mock)

				// Mock product not found
				mock.ExpectQuery("SELECT \\* FROM products WHERE id = \\$1;").
					WithArgs(2000).
					WillReturnError(sql.ErrNoRows)
			},
			expectedCode: http.StatusNotFound,
			expectedBody: "Could not retrieve the product",
		},
		{
			name: "Could not update the product because of a conflict, please refresh",
			path: "/products/1",
			body: `{"updates":{"name":"Updated product"}}`,
			setupMocks: func() {
				AuthenticationMock(mock)

				mock.ExpectQuery("SELECT \\* FROM products WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(productColumns).
						AddRow(1, "Product1", "Brand1", "1234567890123", "", "Comment1", false,
							"Type1", 1.0, 1.0, 1, time.Time{}, "Status1"))
			},
			expectedCode: http.StatusConflict,
			expectedBody: "Could not update the product because of a conflict, please refresh",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			req := httptest.NewRequest(http.MethodPatch, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")
			rr := httptest.NewRecorder()

			ProductsRequestPATCH(rr, req, mockDB)

			assert.Equal(t, tt.expectedCode, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestProductsRequestPOST(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)

	tests := []struct {
		name         string
		body         string
		setupMocks   func()
		expectedCode int
		expectedBody string
	}{
		{
			name: "Create new product",
			body: `{"name":"Product1","brand":"Brand1","ean_code":"1234567890123","image_url":"","comment":"Comment1","is_public":false,"type":"bundle","high_temperature":1.0,"low_temperature":-1.0,"testing_team":1,"status":"active"}`,
			setupMocks: func() {
				AuthenticationMock(mock)

				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products;").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

				mock.ExpectQuery("SELECT \\* FROM products WHERE testing_team = \\$1 AND ean_code = \\$2;").
					WithArgs(1, "1234567890123").
					WillReturnError(sql.ErrNoRows)

				mock.ExpectExec("INSERT INTO products").
					WithArgs("Product1", "Brand1", "1234567890123", "", "Comment1", false,
						"bundle", 1.0, -1.0, 1, sqlmock.AnyArg(), "active").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedCode: http.StatusCreated,
			expectedBody: "",
		},
		{
			name: "Create product with invalid data",
			body: `{"name":1,"brand":45"}`,
			setupMocks: func() {
				AuthenticationMock(mock)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "Invalid POST request body",
		},
		{
			name: "Unable to add new product",
			body: `{"name":"Product1","brand":"Brand1","ean_code":"1234567890123","image_url":"","comment":"Comment1","is_public":false,"type":"bundle","high_temperature":1.0,"low_temperature":-1.0,"testing_team":1,"status":"active"}`,
			setupMocks: func() {
				AuthenticationMock(mock)

				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products;").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

				mock.ExpectQuery("SELECT \\* FROM products WHERE testing_team = \\$1 AND ean_code = \\$2;").
					WithArgs(1, "1234567890123").
					WillReturnError(sql.ErrNoRows)

				mock.ExpectExec("INSERT INTO products").
					WithArgs("Product1", "Brand1", "1234567890123", "", "Comment1", false,
						"bundle", 1.0, -1.0, 1, sqlmock.AnyArg(), "active").
					WillReturnError(sql.ErrNoRows)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "Unable to add new product",
		},
		{
			name: "A product with this EAN code already exists",
			body: `{"name":"Product1","brand":"Brand1","ean_code":"1234567890123","image_url":"","comment":"Comment1","is_public":false,"type":"bundle","high_temperature":1.0,"low_temperature":-1.0,"testing_team":1,"status":"active"}`,
			setupMocks: func() {
				AuthenticationMock(mock)

				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products;").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				mock.ExpectQuery("SELECT \\* FROM products WHERE testing_team = \\$1 AND ean_code = \\$2;").
					WithArgs(1, "1234567890123").
					WillReturnRows(sqlmock.NewRows(productColumns).
						AddRow(1, "Product1", "Brand1", "1234567890123", "", "Comment1", false,
							"bundle", 1.0, -1.0, 1, time.Time{}, "active"))
			},
			expectedCode: http.StatusConflict,
			expectedBody: "A product with this EAN code already exists\n",
		},
		{
			name: "Product creation failed (product count failed)",
			body: `{"name":"Product1","brand":"Brand1","ean_code":"1234567890123","image_url":"","comment":"Comment1","is_public":true,"type":"bundle","high_temperature":1.0,"low_temperature":-1.0,"testing_team":1,"status":"active"}`,
			setupMocks: func() {
				mock.ExpectQuery("SELECT user_id FROM sessions WHERE \\(session_token = \\$1 AND expires_at > NOW\\(\\)\\)").
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				mock.ExpectQuery("SELECT team_id FROM users WHERE id = \\$1;").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(1))

				mock.ExpectQuery("SELECT team_role FROM team WHERE id = \\$1;").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"team_role"}).AddRow(2))
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "Product creation failed\n",
		},
		{
			name: "Researcher cannot create public products",
			body: `{"name":"Product1","brand":"Brand1","ean_code":"1234567890123","image_url":"","comment":"Comment1","is_public":true,"type":"bundle","high_temperature":1.0,"low_temperature":-1.0,"testing_team":1,"status":"active"}`,
			setupMocks: func() {
				mock.ExpectQuery("SELECT user_id FROM sessions WHERE \\(session_token = \\$1 AND expires_at > NOW\\(\\)\\)").
					WithArgs("mockToken").
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

				mock.ExpectQuery("SELECT team_id FROM users WHERE id = \\$1;").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(1))

				mock.ExpectQuery("SELECT team_role FROM team WHERE id = \\$1;").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"team_role"}).AddRow(2))

				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products;").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			expectedCode: http.StatusForbidden,
			expectedBody: "Researcher cannot create public products\n",
		},
		{
			name: "Product name conflict with existing product",
			body: `{"name":"Product1","brand":"Brand1","ean_code":"","image_url":"","comment":"Comment1","is_public":false,"type":"bundle","high_temperature":1.0,"low_temperature":-1.0,"testing_team":1,"status":"active"}`,
			setupMocks: func() {
				AuthenticationMock(mock)

				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products;").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				// When EAN code is blank, we should check if a product with the same name already exists
				mock.ExpectQuery("SELECT \\* FROM products WHERE testing_team = \\$1 AND name = \\$2;").
					WithArgs(1, "Product1").
					WillReturnRows(sqlmock.NewRows(productColumns).
						AddRow(1, "Product1", "Brand1", "", "", "Comment1", false,
							"bundle", 1.0, -1.0, 1, time.Time{}, "active"))
			},
			expectedCode: http.StatusConflict,
			expectedBody: "This Product already exists, try another name if the EAN code is blank\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")
			rr := httptest.NewRecorder()

			ProductsRequestPOST(rr, req, mockDB)

			assert.Equal(t, tt.expectedCode, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestScanProducts(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)

	tests := []struct {
		name    string
		args    *sqlmock.Rows
		want    []domain.Product
		wantErr bool
	}{
		{
			name: "Successful scan",
			args: sqlmock.NewRows(productColumns).
				AddRow(1, "Product1", "Brand1", "1234567890123", "", "Comment1", false,
					"Type1", 1.0, -1.0, 1, time.Time{}, "Status1").
				AddRow(2, "Product2", "Brand2", "1234567890124", "", "Comment2", true,
					"Type2", 2.0, -2.0, 2, time.Time{}, "Status2"),
			want:    productsArray,
			wantErr: false,
		},
		{
			name:    "Could not query",
			args:    sqlmock.NewRows([]string{}),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock HTTP request and response recorder
			req := httptest.NewRequest(http.MethodGet, "/products", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")

			// Set up the mock expectations
			mock.ExpectQuery("SELECT \\* FROM products").WillReturnRows(tt.args)

			// Execute mock query
			mockedRows, err := mockDB.Query("SELECT * FROM products")
			assert.NoError(t, err)

			// Call the function with the mock objects
			products, actualErr := ScanProducts(mockedRows)

			// Reset the time for comparing the products
			resetGot := resetTime(products)
			resetWant := resetTime(tt.want)

			var gotErr bool
			if actualErr != nil {
				gotErr = true
				if gotErr != tt.wantErr {
					t.Errorf("queryAndScanProduct() gotErr = %v, wantErr %v", actualErr, tt.wantErr)
					return
				}
			}

			// Compare the products
			assert.Equal(t, resetGot, resetWant)

			// Ensure all expectations were met
			if err1 := mock.ExpectationsWereMet(); err1 != nil {
				t.Errorf("there were unfulfilled expectations: %v", err1)
			}
		})
	}
}

func TestValidateProductPATCHRequestBody(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)
	tests := []struct {
		name                 string
		db                   *sql.DB
		productUpdateRequest ProductPATCHRequest
		existingProduct      domain.Product
		team                 int
		path                 string
		setupMocks           func()
		want                 error
		code                 int
	}{
		{
			name: "Successful validation",
			db:   mockDB,
			productUpdateRequest: ProductPATCHRequest{
				Updates: map[string]interface{}{"name": "Updated product"},
			},
			existingProduct: testProduct,
			team:            1,
			path:            "/products/1",
			setupMocks: func() {
				// Mock the product lookup in GetProductWithEANCode
				mock.ExpectQuery("SELECT \\* FROM products WHERE testing_team = \\$1 AND name = \\$2;").
					WithArgs(1, "Updated product").
					WillReturnRows(sqlmock.NewRows(productColumns)) // Empty result set
			},
			want: nil,
			code: 0,
		},
		{
			name: "Product uniqueness check failed",
			db:   mockDB,
			productUpdateRequest: ProductPATCHRequest{
				Updates: map[string]interface{}{"name": "Updated product"},
			},
			existingProduct: testProduct,
			team:            1,
			path:            "/products/1",
			setupMocks: func() {
				// Mock the product lookup in GetProductWithEANCode
				mock.ExpectQuery("SELECT \\* FROM products WHERE testing_team = \\$1 AND name = \\$2;").
					WithArgs(1, "Updated product").
					WillReturnRows(sqlmock.NewRows(productColumns).
						AddRow(1, "Updated product", "Brand1", "1234567890123", "/hhhh/", "Comment1", true,
							"Type1", 1.0, 1.0, 1, time.Time{}, "Status1"))
			},
			want: fmt.Errorf("this product allready exists, try another name or ean_code, %d",
				http.StatusConflict),
			code: http.StatusConflict,
		},
		{
			name: "Invalid PATCH request body keys",
			db:   mockDB,
			productUpdateRequest: ProductPATCHRequest{
				Updates: map[string]interface{}{"status1": "not active"},
			},
			existingProduct: testProduct,
			team:            1,
			path:            "/products/1",
			setupMocks: func() {

			},
			want: fmt.Errorf("invalid PATCH request body keys, %d", http.StatusBadRequest),
			code: http.StatusBadRequest,
		},
		{
			name: "Invalid PATCH request body values",
			db:   mockDB,
			productUpdateRequest: ProductPATCHRequest{
				Updates: map[string]interface{}{"status": "not active"},
			},
			existingProduct: testProduct,
			team:            1,
			path:            "/products/1",
			setupMocks:      func() {},
			want:            fmt.Errorf("invalid PATCH request body values, %d", http.StatusBadRequest),
			code:            http.StatusBadRequest,
		},
		{
			name: "User cannot update this product",
			db:   mockDB,
			productUpdateRequest: ProductPATCHRequest{
				Updates: map[string]interface{}{"name": "Updated product"},
			},
			existingProduct: domain.Product{
				ID:              1,
				Name:            "Product1",
				Brand:           "Brand1",
				EANCode:         "1234567890123",
				ImageURL:        "image_url",
				Comment:         "Comment1",
				IsPublic:        false,
				Type:            "Type1",
				HighTemperature: 1.0,
				LowTemperature:  1.0,
				TestingTeam:     2,
				Version:         time.Time{},
				Status:          "Status1",
			},
			team: 1,
			path: "/products/1",
			setupMocks: func() {

			},
			want: fmt.Errorf("user cannot update this product, %d", http.StatusUnauthorized),
			code: http.StatusUnauthorized,
		},
		{
			name: "Researcher cannot update public products",
			db:   mockDB,
			productUpdateRequest: ProductPATCHRequest{
				Updates: map[string]interface{}{"name": "Updated product"},
			},
			existingProduct: domain.Product{
				ID:              1,
				Name:            "Product1",
				Brand:           "Brand1",
				EANCode:         "1234567890123",
				ImageURL:        "image_url",
				Comment:         "Comment1",
				IsPublic:        true,
				Type:            "Type1",
				HighTemperature: 1.0,
				LowTemperature:  1.0,
				TestingTeam:     2,
				Version:         time.Time{},
				Status:          "Status1",
			},
			team: 2,
			path: "/products/1",
			setupMocks: func() {

			},
			want: fmt.Errorf("researcher cannot update public products, %d", http.StatusUnauthorized),
			code: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock expectations
			tt.setupMocks()

			// Create a mock HTTP request and response recorder
			req := httptest.NewRequest(http.MethodPatch, tt.path, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")

			got, got1 := ValidateProductPATCHRequestBody(tt.db, tt.productUpdateRequest, tt.existingProduct, tt.team)
			assert.Equal(t, got, tt.want)
			assert.Equal(t, got1, tt.code)

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func Test_queryAndScanProduct(t *testing.T) {
	mockDB, mock := utils.InitMockDB(t)

	tests := []struct {
		name       string
		query      string
		args       []interface{}
		setupMocks func()
		want       domain.Product
		wantErr    bool
	}{
		{
			name:  "Successful query and scan",
			query: "SELECT * FROM products WHERE id = $1;",
			args:  []interface{}{1},
			setupMocks: func() {
				mock.ExpectQuery("SELECT \\* FROM products WHERE id = \\$1;").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "brand", "ean_code", "image_url", "comment", "is_public", "type", "high_temperature", "low_temperature", "testing_team", "version", "status"}).
						AddRow(1, "Product1", "Brand1", "1234567890123", "image_url", "Comment1", false, "Type1", 1.0, 1.0, 1, time.Time{}, "Status1"))
			},
			want: domain.Product{
				ID:              1,
				Name:            "Product1",
				Brand:           "Brand1",
				EANCode:         "1234567890123",
				ImageURL:        "image_url",
				Comment:         "Comment1",
				IsPublic:        false,
				Type:            "Type1",
				HighTemperature: 1.0,
				LowTemperature:  1.0,
				TestingTeam:     1,
				Version:         time.Time{},
				Status:          "Status1",
			},
			wantErr: false,
		},
		{
			name:  "Could not query",
			query: "SELECT * FROM products WHERE id = $1;",
			args:  []interface{}{1},
			setupMocks: func() {
				mock.ExpectQuery("SELECT \\* FROM products WHERE id = \\$1;").
					WithArgs(1).
					WillReturnError(errors.New("db error"))
			},
			want:    domain.Product{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock expectations
			tt.setupMocks()

			// Create a mock HTTP request and response recorder
			req := httptest.NewRequest(http.MethodGet, "/products", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer mockToken")

			// Call the function with the mock objects
			product, err := queryAndScanProduct(mockDB, tt.query, tt.args...)

			if (err != nil) != tt.wantErr {
				t.Errorf("queryAndScanProduct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(product, tt.want) {
				t.Errorf("queryAndScanProduct() got = %v, want %v", product, tt.want)
				return
			}

			// Ensure all expectations were met
			if err1 := mock.ExpectationsWereMet(); err1 != nil {
				t.Errorf("there were unfulfilled expectations: %v", err1)
			}
		})
	}
}
