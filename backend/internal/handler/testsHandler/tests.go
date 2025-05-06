package testsHandler

import (
	"backend/internal/domain"
	"backend/internal/middleware"
	"backend/internal/resources"
	"backend/internal/utils"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TestsHandler routes HTTP requests for tests to the appropriate handler function.
//
// It supports the following methods:
// - GET: Retrieves a list of tests based on filters.
// - POST: Creates a new test.
// - PUT: Updates an existing test.
//
// Each method has its own dedicated request handler function, which is documented separately using Swagger annotations.
func TestsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.Method {
		case http.MethodGet:
			TestsRequestGET(w, r, db)
		case http.MethodPost:
			TestsRequestPOST(w, r, db)
		case http.MethodPatch:
			TestsRequestPATCH(w, r, db)
		default:
			http.Error(w, resources.MethodNotAllowed, http.StatusNotImplemented)
			return
		}
	}
}

// TestsRequestGET handles GET requests for tests.
//
//	@Summary		Get a list of tests
//	@Description	Retrieves a list of tests based on query parameters.
//	@Tags			Tests
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			public		query		string		false	"All Public Tests"
//	@Param			start_date	query		string		false	"Start Date in YYYY-MM-DD format"
//	@Param			end_date	query		string		false	"End date in YYYY-MM-DD format"
//	@Success		200			{array}		domain.Test	"Successful response with a list of tests"
//	@Failure		400			{string}	string		"Invalid start date format."
//	@Failure		500			{string}	string		"Could not retrieve all tests."
//	@Router			/tests [get]
//	@Router			/tests/ [get]
func TestsRequestGET(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	public := r.URL.Query().Get("public")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	var tests []domain.Test

	if public == "true" {
		// Fetch all the public tests from the database.
		rows, err := db.Query("SELECT * FROM tests WHERE is_public = $1;", true)
		FetchTests(w, tests, rows, err)
		return
	}

	// Get the user's team membership.
	team := middleware.GetUserTeamRole(w, r, db)

	if startDate != "" && endDate != "" {
		if _, err := time.Parse(time.DateOnly, startDate); err != nil {
			http.Error(w, "Invalid start date format.", http.StatusBadRequest)
			log.Println("Invalid start date format: " + err.Error())
			return
		}

		if _, err := time.Parse(time.DateOnly, endDate); err != nil {
			http.Error(w, "Invalid end date format.", http.StatusBadRequest)
			log.Println("Invalid end date format: " + err.Error())
			return
		}

		// Fetch all the tests from the database where the date is between the start and end date.
		rows, err := db.Query(
			`SELECT * FROM tests WHERE testing_team = $1 AND 
                          test_date >= to_date($2, 'YYYY-MM-DD') AND test_date <= to_date($3, 'YYYY-MM-DD');`,
			team, startDate, endDate)
		FetchTests(w, tests, rows, err)
		return
	}

	// Fetch all the tests from the database (both private and public) for the different authenticated user's
	rows, err := db.Query("SELECT * FROM tests WHERE testing_team = $1;", team)
	FetchTests(w, tests, rows, err)
	return
}

// TestsRequestPOST is the request handler for creating a new test.
//
//	@Summary		Create a new test
//	@Description	Adds a new test to the database.
//	@Tags			Tests
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			test	body	TestPOSTRequest	true	"New test information"
//	@Success		201		"Test created successfully"
//	@Failure		500		{string}	string	"Could not create test."
//	@Router			/tests [post]
//	@Router			/tests/ [post]
func TestsRequestPOST(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get the user's team membership.
	team := middleware.GetUserTeamRole(w, r, db)

	// Parse and validate request
	test, err := utils.ParseAndValidateRequest[TestPOSTRequest](r)
	if err != nil {
		http.Error(w, resources.InvalidPOSTRequest, http.StatusBadRequest)
		log.Println(resources.InvalidPOSTRequest + ": " + err.Error())
		return
	}

	// Log the valid request
	var testJSON []byte
	testJSON, err = json.MarshalIndent(test, "", "  ")
	if err != nil {
		log.Println("Error marshaling test:", err)
	} else {
		log.Println("Test JSON:", string(testJSON))
	}

	// Validate role permissions
	if permErr := validatePermissions(test, team); permErr != nil {
		http.Error(w, "Researcher cannot create public tests", http.StatusBadRequest)
		log.Println("Researcher cannot create public tests")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, resources.TransactionStartFailed, http.StatusInternalServerError)
		return
	}
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				log.Println(resources.RollbackFailed + rollbackErr.Error())
			}
		}
	}()

	// Create test with all relates entities
	testID, err := createTest(tx, test, team)
	if err != nil {
		http.Error(w, "Failed to create test: "+err.Error(), http.StatusInternalServerError)
		log.Println("Failed to create test: " + err.Error())
		return
	}

	// Create rankings
	err = createTestRankings(tx, testID, test.TestRanks)
	if err != nil {
		http.Error(w, "Failed to create rankings: "+err.Error(), http.StatusInternalServerError)
		log.Println("Failed to create rankings: " + err.Error())
		return
	}

	//Commit transaction
	err = tx.Commit()
	if err != nil {
		http.Error(w, "Failed to commit the transaction.", http.StatusInternalServerError)
		log.Println("Failed to commit the transaction: " + err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// TestsRequestPATCH is the request handler for updating an existing test.
//
//	@Summary		Update an existing test's information, ranks, ac, tc and/or sc.
//	@Description	Updates an existing test in the database.
//	@Tags			Tests
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			test_id		path		int					true	"Test ID"
//	@Param			product_id	path		int					true	"Product ID"
//	@Param			test		body		TestPATCHRequest	true	"Test updates"
//	@Success		200			{string}	string				"Test updated successfully"
//	@Failure		400			{string}	string				"Missing test ID"
//	@Failure		400			{string}	string				"Invalid test ID"
//	@Failure		400			{string}	string				"Invalid patch request."
//	@Failure		409			{string}	string				"Detected a conflict for the current test, please refresh."
//	@Failure		500			{string}	string				"Could not JSON encode the response."
//	@Router			/tests/ [patch]
//	@Router			/tests/{test_id}/products/{product_id} [patch]
func TestsRequestPATCH(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	updateTestRanksPath := regexp.MustCompile(`^/tests/(\d+)/products/(\d+)$`)
	matches := updateTestRanksPath.FindStringSubmatch(r.URL.Path)

	var testID, productID int
	var err error

	if len(matches) == 3 {
		// This is a product update request
		testID, err = strconv.Atoi(matches[1])
		if err != nil {
			http.Error(w, "Invalid test ID", http.StatusBadRequest)
			log.Println("Invalid test ID: " + err.Error())
			return
		}

		productID, err = strconv.Atoi(matches[2])
		if err != nil {
			http.Error(w, "Invalid product ID", http.StatusBadRequest)
			log.Println("Invalid product ID: " + err.Error())
			return
		}
	} else {
		// Regular test update, parse the ID from the URL
		idParam := strings.TrimPrefix(r.URL.Path, "/tests/")
		testID, err = utils.GetIDFromURLQuery(w, idParam)
		if err != nil {
			return
		}
		productID = 0
	}

	// Decode the request body into the testUpdateRequest struct.
	var testUpdateRequest TestPATCHRequest
	err = utils.DecodeRequestBody(w, r, &testUpdateRequest)
	if err != nil {
		http.Error(w, resources.InvalidPATCHRequest, http.StatusBadRequest)
		log.Println(resources.InvalidPATCHRequest + ": " + err.Error())
		return
	}

	// Get the existing test from the database.
	existingTest := GetTestWithID(w, db, testID)

	// Validate the PATCH request body
	var code int
	err, code = ValidateTestPATCHRequestBody(w, r, db, testUpdateRequest, existingTest)
	if err != nil {
		http.Error(w, "Validation error: "+err.Error(), code)
		return
	}

	// Parse the version timestamps for the existing and updated test.
	existingTestVersion, testUpdateVersion := utils.ParseVersionTimestamps(w,
		testUpdateRequest.Version, existingTest.Version)

	// Solve the concurrency challenge by checking if the product has been updated since the last sync.
	if existingTestVersion.After(testUpdateVersion) {
		http.Error(w, "Detected a conflict for the current test, please refresh.", http.StatusConflict)
		log.Println("Detected a conflict for the current test, please refresh.")
		return
	}
	newVersion := createTestUpdateQueries(w, r, db, testUpdateRequest,
		existingTestVersion, testID, productID)

	// Send a response if the test update was successful.
	if !newVersion.IsZero() {
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Test updated successfully",
			"version": newVersion,
		})
		if err != nil {
			http.Error(w, "Could not JSON encode the response.", http.StatusInternalServerError)
			log.Println("Could not JSON encode the response: " + err.Error())
			return
		}
	}
}
