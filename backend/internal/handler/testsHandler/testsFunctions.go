package testsHandler

import (
	"backend/internal/domain"
	"backend/internal/middleware"
	"backend/internal/resources"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func FetchTests(w http.ResponseWriter, tests []domain.Test, rows *sql.Rows, err error) {
	for rows.Next() {
		var test domain.Test

		if err = rows.Scan(
			&test.ID,
			&test.Date,
			&test.Location,
			&test.Comment,
			&test.SnowConditions,
			&test.TrackConditions,
			&test.AirConditions,
			&test.Version,
			&test.IsPublic,
			&test.TestingTeam); err != nil {
			http.Error(w, "Could not retrieve all tests ", http.StatusInternalServerError)
			log.Println("Could not retrieve all tests " + err.Error())
			return
		}

		tests = append(tests, test)
	}

	if len(tests) == 0 {
		http.Error(w, "No tests found.", http.StatusOK)
		log.Println("No tests found.")
		return
	}

	// Write the response to the client.
	err = json.NewEncoder(w).Encode(tests)
	if err != nil {
		http.Error(w, "Could not encode tests.", http.StatusInternalServerError)
		log.Println("Could not encode tests.", err.Error())
		return
	}

	return
}

// GetTestWithID retrieves a test with a specific ID from the database.
func GetTestWithID(w http.ResponseWriter, db *sql.DB, testID int) domain.Test {
	var test domain.Test

	// Get the existing test from the database.
	err := db.QueryRow("SELECT * FROM tests WHERE id = $1;", testID).Scan(
		&test.ID,
		&test.Date,
		&test.Location,
		&test.Comment,
		&test.SnowConditions,
		&test.TrackConditions,
		&test.AirConditions,
		&test.Version,
		&test.IsPublic,
		&test.TestingTeam)
	if err != nil {
		http.Error(w, "Could not retrieve the test.", http.StatusNotFound)
		log.Println("Could not retrieve the test: " + err.Error())
		return domain.Test{}
	}
	return test
}

func ValidateTestPATCHRequestBody(w http.ResponseWriter, r *http.Request, db *sql.DB,
	testUpdateRequest TestPATCHRequest, existingTest domain.Test) (error, int) {
	var update TestUpdateFields
	b, _ := json.Marshal(testUpdateRequest.Updates)
	err := json.Unmarshal(b, &update)
	if err != nil {
		log.Println("Could not decode request body: " + err.Error())
		return fmt.Errorf("could not decode request body, %d", http.StatusInternalServerError), http.StatusInternalServerError
	}

	validate := validator.New()

	// Validate keys
	err = validate.Struct(testUpdateRequest)
	if err != nil {
		log.Println("Invalid PATCH request body keys: " + err.Error())
		return fmt.Errorf("invalid PATCH request body keys, %d", http.StatusBadRequest), http.StatusBadRequest
	}

	// Validate values
	err = validate.Struct(update)
	if err != nil {
		log.Println("Invalid PATCH request body values: " + err.Error())
		return fmt.Errorf("invalid PATCH request body values, %d", http.StatusBadRequest), http.StatusBadRequest
	}

	// Check if the test exists.
	if existingTest.IsPublic == true && testUpdateRequest.Updates["is_public"] == false {
		log.Println("Test is public and cannot be made private.")
		return fmt.Errorf("test is public and cannot be made private, %d", http.StatusBadRequest), http.StatusBadRequest
	}

	// Check if the test is publicly available and the user is a researcher.
	team := middleware.GetUserTeamRole(w, r, db)
	if existingTest.IsPublic == true && domain.TeamRole(team) == domain.Researcher {
		log.Println("Researcher cannot update public tests")
		return fmt.Errorf("researcher cannot update public tests, %d", http.StatusUnauthorized), http.StatusUnauthorized
	}

	if existingTest.IsPublic == false && testUpdateRequest.Updates["is_public"] == true &&
		domain.TeamRole(team) == domain.Researcher {
		log.Println("Researcher cannot make tests public")
		return fmt.Errorf("researcher cannot make tests public, %d", http.StatusUnauthorized), http.StatusUnauthorized
	}

	// Check if the user is authorized to update the test.
	if existingTest.TestingTeam != team {
		log.Println("User cannot update this test")
		return fmt.Errorf("user cannot update this test, %d", http.StatusUnauthorized), http.StatusUnauthorized
	}

	return nil, 0
}

func validatePermissions(test TestPOSTRequest, team int) error {
	if test.IsPublic == true && domain.TeamRole(team) == domain.Researcher {
		return fmt.Errorf("researcher cannot create public tests, %d", http.StatusUnauthorized)
	}
	return nil
}

func insertSnowConditions(tx *sql.Tx, conditions SnowConditionsPOST) (int, error) {
	var snowConditionID int
	err := tx.QueryRow(`INSERT INTO snow_conditions 
    						(temperature, snow_type, snow_humidity) 
							VALUES ($1, $2, $3) 
							RETURNING id;`,
		conditions.Temperature,
		conditions.SnowType,
		conditions.SnowHumidity).Scan(&snowConditionID)
	return snowConditionID, err
}

func insertAirConditions(tx *sql.Tx, conditions AirConditionsPOST) (int, error) {
	var airConditionID int
	err := tx.QueryRow(`INSERT INTO air_conditions 
    						(temperature, humidity, wind, cloud) 
							VALUES ($1, $2, $3, $4)
							RETURNING id;`,
		conditions.Temperature,
		conditions.Humidity,
		conditions.Wind,
		conditions.Cloud).Scan(&airConditionID)
	return airConditionID, err
}

func insertTrackConditions(tx *sql.Tx, conditions TrackConditionsPOST) (int, error) {
	var trackConditionID int
	err := tx.QueryRow(`INSERT INTO track_conditions 
							(track_hardness, track_type) 
							VALUES ($1, $2)
							RETURNING id;`,
		conditions.TrackHardness,
		conditions.TrackType).Scan(&trackConditionID)
	return trackConditionID, err
}

func createTest(tx *sql.Tx, test TestPOSTRequest, team int) (int, error) {
	// Insert snow conditions to database
	snowConditionID, err := insertSnowConditions(tx, test.SnowConditions)
	if err != nil {
		log.Printf("Snow Conditions before error: %+v", test.SnowConditions)
		return 0, fmt.Errorf("failed to insert snow conditions: %w", err)
	}

	// Insert air conditions to database
	airConditionID, err := insertAirConditions(tx, test.AirConditions)
	if err != nil {
		return 0, fmt.Errorf("failed to insert air conditions:  %w", err)
	}

	// Insert track conditions to database
	trackConditionID, err := insertTrackConditions(tx, test.TrackConditions)
	if err != nil {
		return 0, fmt.Errorf("failed to insert track conditions:  %w", err)
	}

	// Insert test to database
	var testID int
	err = tx.QueryRow(`INSERT INTO tests (
                   			test_date, location, comment, sc_id, tc_id, ac_id, 
                   			version, is_public, testing_team) 
							VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
							RETURNING id;`,
		test.Date,
		test.Location,
		test.Comment,
		snowConditionID,
		trackConditionID,
		airConditionID,
		time.Now(),
		test.IsPublic,
		team).Scan(&testID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert test:  %w", err)
	}
	return testID, nil
}

/*
// Insert test products into the database
func insertTestProducts(tx *sql.Tx, rankingID int, testProduct TestProductsPOSTRequest, topLayerID int) error {
	_, err := tx.Exec(`INSERT INTO test_products (
		ranking_id, glider, mid_layer, top_layer)
		VALUES ($1, $2, $3, $4);`,
		rankingID,
		testProduct.GliderID,
		testProduct.MidLayerID,
		topLayerID)
	return err
}

// Create a top layer with its products
func createTopLayer(tx *sql.Tx, productIDs []int) (int, error) {
	var topLayerID int
	err := tx.QueryRow(`INSERT INTO top_layers DEFAULT VALUES RETURNING id;`).Scan(&topLayerID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert top layers: " + err.Error())
	}

	for layerNo, productID := range productIDs {
		_, err = tx.Exec(`INSERT INTO top_layer_products (
			top_layer_id, product_id, layer_no)
			VALUES ($1, $2, $3);`, topLayerID, productID, layerNo+1)
		if err != nil {
			return 0, err
		}
	}
	return topLayerID, nil
}

*/

// Fetch the products availability status for use in the test ranking
func getAvailabilityFromProduct(tx *sql.Tx, productID int) (bool, error) {
	var availability bool
	err := tx.QueryRow("SELECT is_public FROM products WHERE id = $1", productID).Scan(&availability)
	return availability, err
}

// Insert test rankings into the database
func insertTestRanking(tx *sql.Tx, testID int, testRank TestRanksPOST) error {
	var isRankPublic, err = getAvailabilityFromProduct(tx, testRank.ProductID)
	if err != nil {
		return fmt.Errorf("failed to get product availability: %w", err)
	}
	_, err = tx.Exec(`INSERT INTO test_ranks (
                      	test_id, product_id, rank, distance_behind, version, is_rank_public)
						VALUES ($1, $2, $3, $4, $5, $6)`,
		testID,
		testRank.ProductID,
		testRank.Rank,
		testRank.DistanceBehind,
		time.Now(),
		isRankPublic)
	return err
}

// Create rankings and associated products
func createTestRankings(tx *sql.Tx, testID int, testRankings []TestRanksPOST) error {
	for _, rank := range testRankings {
		// Insert winner ranking
		err := insertTestRanking(tx, testID, rank)
		if err != nil {
			return fmt.Errorf("failed to insert winner ranking: %w", err)
		}
	}
	return nil
}

func createTestUpdateQueries(w http.ResponseWriter, r *http.Request, db *sql.DB, testUpdateRequest TestPATCHRequest,
	existingTestVersion time.Time, testID int, productID int) time.Time {
	// Create the updatedFields and newValues arrays for the query, and increment the index for the newValues array.
	var updatedFields []string
	var newValues []interface{}
	i := 1 // Index for the newValues array.

	for field, value := range testUpdateRequest.Updates {
		updatedFields = append(updatedFields, fmt.Sprintf("%s = $%d", field, i)) // field = $x for the db query.
		newValues = append(newValues, value)
		i++
	}

	// If the update request contains no fields, return an error.
	if len(updatedFields) == 0 {
		http.Error(w, resources.NoFieldsToUpdate, http.StatusBadRequest)
		log.Println(resources.NoFieldsToUpdate)
	}

	// Track which update functions need to be called
	containsACUpdateFields := false
	containsSCUpdateFields := false
	containsTCUpdateFields := false
	containsRankUpdateFields := false
	containsTestUpdateFields := false

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, resources.TransactionStartFailed, http.StatusInternalServerError)
		log.Println("Could not start transaction: " + err.Error())
		return time.Time{}
	}

	// Validate fields
	for field := range testUpdateRequest.Updates {
		switch {

		case validACFields[field]:
			containsACUpdateFields = true
		case validSCFields[field]:
			containsSCUpdateFields = true
		case validTCFields[field]:
			containsTCUpdateFields = true
		case validRankFields[field]:
			containsRankUpdateFields = true
		case validTestFields[field]:
			containsTestUpdateFields = true
		default:
			http.Error(w, fmt.Sprintf("Invalid field: %s", field), http.StatusBadRequest)
			log.Printf("Invalid field: %s", field)
			return time.Time{}
		}
	}

	var newVersion time.Time
	if containsACUpdateFields {
		newVersion = updateAirConditions(w, tx, updatedFields, newValues, testID)
	}

	if containsSCUpdateFields {
		newVersion = updateSnowConditions(w, tx, updatedFields, newValues, testID)
	}

	if containsTCUpdateFields {
		newVersion = updateTrackConditions(w, tx, updatedFields, newValues, testID)
	}

	if containsRankUpdateFields && productID > 0 {
		newVersion = updateTestRankings(w, tx, newVersion, updatedFields,
			newValues, testID, productID)
	} else if containsRankUpdateFields && productID == 0 {
		http.Error(w, "Invalid request URL, use '/tests/{test_id}/product/{product_id}' to update test ranks.",
			http.StatusBadRequest)
		log.Println("Invalid request URL: " + r.URL.Path)
		return time.Time{}
	}

	if containsTestUpdateFields {
		newVersion = updateTestFields(w, tx, newVersion, existingTestVersion, updatedFields, newValues, testID)
	}

	// Commit the transaction.
	err = tx.Commit()
	if err != nil {
		http.Error(w, resources.TransactionCommitFailed, http.StatusInternalServerError)
		log.Println("Could not commit transaction: " + err.Error())
		return time.Time{}
	}

	return newVersion
}

// Update the air conditions in the database.
func updateAirConditions(w http.ResponseWriter, tx *sql.Tx, updatedFields []string,
	newValues []interface{}, testID int) time.Time {
	validUpdatedFields, validIdx, validNewValues := mapUpdatedTestConditionFields(updatedFields, newValues,
		validACFields, actualFieldNamesAC)

	// If no valid fields to update, return early
	if len(validUpdatedFields) == 0 {
		http.Error(w, resources.NoFieldsToUpdate, http.StatusBadRequest)
		log.Println(resources.NoFieldsToUpdate)
		return time.Time{}
	}

	// Get the air condition ID for this test
	var acID int
	err := tx.QueryRow("SELECT ac_id FROM tests WHERE id = $1", testID).Scan(&acID)
	if err != nil {
		http.Error(w, "Could not find the air conditions for this test", http.StatusNotFound)
		log.Println("Could not find air conditions ID: " + err.Error())
		return time.Time{}
	}

	// Create the query to update the air conditions in the database
	query := fmt.Sprintf("UPDATE air_conditions SET %s WHERE id = $%d",
		strings.Join(validUpdatedFields, ", "),
		validIdx,
	)
	validNewValues = append(validNewValues, acID)

	// Execute the query
	_, err = tx.Exec(query, validNewValues...)
	if err != nil {
		http.Error(w, "Could not update the air conditions because of a conflict, please refresh.", http.StatusConflict)
		log.Println("Could not update the air conditions because of a conflict, please refresh: " + err.Error())
		return time.Time{}
	}
	return time.Now()
}

// Update the track conditions in the database.
func updateTrackConditions(w http.ResponseWriter, tx *sql.Tx, updatedFields []string, newValues []interface{}, testID int) time.Time {
	validUpdatedFields, validIdx, validNewValues := mapUpdatedTestConditionFields(updatedFields, newValues,
		validTCFields, actualFieldNamesTC)

	// If no valid fields to update, return early
	if len(validUpdatedFields) == 0 {
		http.Error(w, resources.NoFieldsToUpdate, http.StatusBadRequest)
		log.Println(resources.NoFieldsToUpdate)
		return time.Time{}
	}

	// Get the track condition ID for this test
	var tcID int
	err := tx.QueryRow("SELECT tc_id FROM tests WHERE id = $1", testID).Scan(&tcID)
	if err != nil {
		http.Error(w, "Could not find the track conditions for this test", http.StatusNotFound)
		log.Println("Could not find track conditions ID: " + err.Error())
		return time.Time{}
	}

	// Create the query to update the track conditions in the database.
	query := fmt.Sprintf("UPDATE track_conditions SET %s WHERE id = $%d",
		strings.Join(validUpdatedFields, ", "), // Use validUpdatedFields instead of updatedFields
		validIdx,
	)
	validNewValues = append(validNewValues, tcID) // Add tcID to validNewValues

	// Execute the query.
	_, err = tx.Exec(query, validNewValues...) // Use validNewValues instead of newValues
	if err != nil {
		http.Error(w, "Could not update the track conditions because of a conflict, please refresh.", http.StatusConflict)
		log.Println("Could not update the track conditions because of a conflict, please refresh: " + err.Error())
		return time.Time{}
	}
	return time.Now()
}

// Update the snow conditions in the database.
func updateSnowConditions(w http.ResponseWriter, tx *sql.Tx, updatedFields []string, newValues []interface{}, testID int) time.Time {
	validUpdatedFields, validIdx, validNewValues := mapUpdatedTestConditionFields(updatedFields, newValues,
		validSCFields, actualFieldNamesSC)

	// If no valid fields to update, return early
	if len(validUpdatedFields) == 0 {
		http.Error(w, resources.NoFieldsToUpdate, http.StatusBadRequest)
		log.Println(resources.NoFieldsToUpdate)
		return time.Time{}
	}

	// Get the snow condition ID for this test
	var scID int
	err := tx.QueryRow("SELECT sc_id FROM tests WHERE id = $1", testID).Scan(&scID)
	if err != nil {
		http.Error(w, "Could not find the snow conditions for this test", http.StatusNotFound)
		log.Println("Could not find snow conditions ID: " + err.Error())
		return time.Time{}
	}

	// Create the query to update the snow conditions in the database.
	query := fmt.Sprintf("UPDATE snow_conditions SET %s WHERE id = $%d",
		strings.Join(validUpdatedFields, ", "), // Use validUpdatedFields instead of updatedFields
		validIdx,
	)
	validNewValues = append(validNewValues, scID) // Add scID to validNewValues

	// Execute the query.
	_, err = tx.Exec(query, validNewValues...) // Use validNewValues instead of newValues
	if err != nil {
		http.Error(w, "Could not update the snow conditions because of a conflict, please refresh.", http.StatusConflict)
		log.Println("Could not update the snow conditions because of a conflict, please refresh: " + err.Error())
		return time.Time{}
	}
	return time.Now()
}

// Update the test rankings in the database.
func updateTestRankings(w http.ResponseWriter, tx *sql.Tx, newVersion time.Time,
	updatedFields []string, newValues []interface{}, testID int, productID int) time.Time {
	validUpdatedFields, validIdx, validNewValues := mapUpdatedTestAndRankingsFields(updatedFields, newValues, validRankFields)

	// If no valid fields to update, return early
	if len(validUpdatedFields) == 0 {
		http.Error(w, resources.NoFieldsToUpdate, http.StatusBadRequest)
		log.Println(resources.NoFieldsToUpdate)
		return time.Time{}
	}

	// Create the query to update the test rankings in the database.
	query := fmt.Sprintf("UPDATE test_ranks SET %s WHERE test_id = $%d AND product_id = $%d RETURNING version",
		strings.Join(validUpdatedFields, ", "), // validUpdatedFields parameter ($1, $2, ...)
		validIdx,                               // testID parameter
		validIdx+1,                             // productID parameter
	)
	validNewValues = append(validNewValues, testID, productID)

	// Execute the query and get the new version of the test.
	err := tx.QueryRow(query, validNewValues...).Scan(&newVersion)
	if err != nil {
		http.Error(w, "Could not update the test rankings because of a conflict, please refresh.", http.StatusConflict)
		log.Println("Could not update the test rankings because of a conflict, please refresh: " + err.Error())
		return time.Time{}
	}
	return newVersion
}

func updateTestFields(w http.ResponseWriter, tx *sql.Tx, newVersion time.Time, existingTestVersion time.Time,
	updatedFields []string, newValues []interface{}, testID int) time.Time {
	validUpdatedFields, validIdx, validNewValues := mapUpdatedTestAndRankingsFields(updatedFields, newValues, validTestFields)

	// If no valid fields to update, return early
	if len(validUpdatedFields) == 0 {
		http.Error(w, resources.NoFieldsToUpdate, http.StatusBadRequest)
		log.Println(resources.NoFieldsToUpdate)
		return time.Time{}
	}

	// Create the query to update the test in the database.
	query := fmt.Sprintf("UPDATE tests SET %s WHERE id = $%d AND version = $%d RETURNING version",
		strings.Join(validUpdatedFields, ", "), // validUpdatedFields parameter ($1, $2, ...)
		validIdx,                               // testID parameter
		validIdx+1,                             // Version parameter
	)
	validNewValues = append(validNewValues, testID, existingTestVersion)

	// Execute the query and get the new version of the test.
	err := tx.QueryRow(query, validNewValues...).Scan(&newVersion)
	if err != nil {
		http.Error(w, "Could not update the test because of a conflict, please refresh.", http.StatusConflict)
		log.Println("Could not update the test because of a conflict, please refresh: " + err.Error())
		return time.Time{}
	}

	return newVersion
}

func mapUpdatedTestAndRankingsFields(updatedFields []string, newValues []interface{}, validFields map[string]bool) (
	[]string, int, []interface{}) {
	var validUpdatedFields []string
	var validNewValues []interface{}
	validIdx := 1

	// Map the updated field names to the actual field names in the database and only allow fields
	// for this table to get included in the query.
	for idx, field := range updatedFields {
		parts := strings.SplitN(field, " = ", 2)
		if len(parts) >= 2 {
			fieldName := parts[0]
			if validFields[fieldName] {
				// Check if the field is a valid rank field
				validUpdatedFields = append(validUpdatedFields, fieldName+" = $"+strconv.Itoa(validIdx))
				validNewValues = append(validNewValues, newValues[idx])
				validIdx++
			}
		}
	}

	// If no valid fields to update, return early
	if len(validUpdatedFields) == 0 {
		return nil, 0, nil
	}

	// Add the version field to the validUpdateFields and validNewValues arrays.
	validUpdatedFields = append(validUpdatedFields, "version = $"+strconv.Itoa(validIdx))
	validNewValues = append(validNewValues, time.Now())
	validIdx++

	return validUpdatedFields, validIdx, validNewValues
}

func mapUpdatedTestConditionFields(updatedFields []string, newValues []interface{},
	validFields map[string]bool, actualNames map[string]string) ([]string, int, []interface{}) {
	var validUpdatedFields []string
	var validNewValues []interface{}
	validIdx := 1

	// Map the updated field names to the actual field names in the database and only allow fields
	// for this table to get included in the query.
	for idx, field := range updatedFields {
		parts := strings.SplitN(field, " = ", 2)
		if len(parts) == 2 {
			fieldName := parts[0]
			if validFields[fieldName] {
				// Check if the field is a valid snow condition field
				if actualName, exists := actualNames[fieldName]; exists {
					validUpdatedFields = append(validUpdatedFields, actualName+" = $"+strconv.Itoa(validIdx))
					validNewValues = append(validNewValues, newValues[idx])
					validIdx++
				}
			}
		}
	}

	// If no valid fields to update, return early
	if len(validUpdatedFields) == 0 {
		return nil, 0, nil
	}

	return validUpdatedFields, validIdx, validNewValues
}
