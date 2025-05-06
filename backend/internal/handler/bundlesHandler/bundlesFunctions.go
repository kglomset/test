package bundlesHandler

//coverage:ignore file
import (
	"backend/internal/domain"
	"backend/internal/resources"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"
)

// FetchBundles represents the request body of a POST request to create a new bundle.
func FetchBundles(w http.ResponseWriter, bundles *[]domain.ProductBundle, rows *sql.Rows) error {
	// Iterate over the rows and scan the data into the bundle struct.
	for rows.Next() {
		var bundle domain.ProductBundle

		if err := rows.Scan(
			&bundle.BundleID,
			&bundle.ProductID,
			&bundle.LayerNumber,
		); err != nil {
			return fmt.Errorf("could not scan bundle: %v", err)
		}

		// Append the bundle to the bundles slice.
		*bundles = append(*bundles, bundle)
	}

	// Check for any errors encountered during iteration
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error reading rows: %v", err)
	}

	return nil
}

/*
// getTestRankIDs retrieves a ranking with a specific ID from the database.
func getTestRankIDs(w http.ResponseWriter, db *sql.DB, team int) []int {
	// Get the test IDs for the testing team.
	testIDs := getTestIDs(w, db, team)

	var rows *sql.Rows
	var err error
	for _, testID := range testIDs {
		// Query the database for the rankings.
		rows, err = db.Query("SELECT * FROM test_ranks WHERE test_id = $1;", testID)
		if err != nil {
			http.Error(w, resources.CouldNotRetrieveBundles, http.StatusInternalServerError)
			log.Println("Could not retrieve any rankings for bundle: " + err.Error())
			return nil
		}
	}

	// Iterate over the rows and scan the data into the ranking struct.
	var testRank []domain.Ranking
	for rows.Next() {
		var ranking domain.Ranking

		if err = rows.Scan(
			&ranking.ID,
			&ranking.TestID,
			&ranking.Rank,
			&ranking.Version,
			&ranking.IsRankPublic,
			&ranking.Wins); err != nil {
			http.Error(w, resources.CouldNotRetrieveBundles, http.StatusInternalServerError)
			log.Println("Could not retrieve any rankings for the bundle: " + err.Error())
		}

		// Append the ranking to the rankings slice.
		rankings = append(rankings, ranking)
	}

	// Check if no rankings were found.
	if len(rankings) == 0 {
		http.Error(w, "No bundles found.", http.StatusOK)
		log.Println("No rankings found.")
		return nil
	}

	var rankingIDs []int
	for _, ranking := range rankings {
		rankingIDs = append(rankingIDs, ranking.ID)
	}
	return rankingIDs
}



// getTestIDs retrieves the ID of a test made by a testing team from the database.
func getTestIDs(w http.ResponseWriter, db *sql.DB, team int) []int {
	// Query the database for the rankings.
	rows, err := db.Query("SELECT * FROM tests WHERE testing_team = $1;", team)
	if err != nil {
		http.Error(w, "No bundles found for this testing team.", http.StatusInternalServerError)
		log.Println("Could not retrieve any rankings for bundle: " + err.Error())
		return nil
	}

	var tests []domain.Test
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
			&test.IsRankPublic,
			&test.TestingTeam); err != nil {
			http.Error(w, resources.CouldNotRetrieveBundles, http.StatusInternalServerError)
			log.Println("Could not retrieve any tests for the bundle: " + err.Error())
		}

		// Append the ranking to the tests slice.
		tests = append(tests, test)
	}

	// Check if no rankings were found.
	if len(tests) == 0 {
		http.Error(w, "No bundles found for this testing team.", http.StatusOK)
		log.Println("No tests found for this testing team.")
		return nil
	}

	var testIDs []int
	for _, test := range tests {
		testIDs = append(testIDs, test.ID)
	}
	return testIDs
}

*/

func getAllBundles(w http.ResponseWriter, db *sql.DB, bundles []domain.ProductBundle) []domain.ProductBundle {
	// Fetch all the bundles from the database for the different authenticated user's role and team memberships.
	rows, err := db.Query("SELECT * FROM product_bundles;")
	if err != nil {
		http.Error(w, resources.CouldNotRetrieveBundles, http.StatusInternalServerError)
		log.Println(resources.CouldNotRetrieveBundles, err.Error())
		return nil
	}
	if err = FetchBundles(w, &bundles, rows); err != nil {
		http.Error(w, resources.CouldNotRetrieveBundles, http.StatusInternalServerError)
		log.Println(resources.CouldNotRetrieveBundles, err.Error())
		return nil
	}

	return bundles
}

func getBundlesByBundleID(w http.ResponseWriter, db *sql.DB, bundles []domain.ProductBundle,
	id int, idStr string) []domain.ProductBundle {
	// Fetch the product with the given id from the database based on the user's role and team membership.
	rows, err := db.Query("SELECT * FROM product_bundles WHERE bundle_id = $1;", id)
	if err != nil {
		http.Error(w, "Could not retrieve bundle with id: "+idStr, http.StatusInternalServerError)
		log.Println(resources.CouldNotRetrieveProduct, err.Error())
		return nil
	}
	if err = FetchBundles(w, &bundles, rows); err != nil {
		http.Error(w, resources.CouldNotRetrieveBundles, http.StatusInternalServerError)
		log.Println(resources.CouldNotRetrieveBundles, err.Error())
		return nil
	}

	return bundles
}

func insertBundles(tx *sql.Tx, bundleID int, productID int, layerNo int) error {
	_, err := tx.Exec(`INSERT INTO product_bundles (
                             bundle_id, product_id, layer_no) 
							 VALUES ($1, $2, $3)`,
		bundleID, productID, layerNo)
	return err
}

func createBundles(tx *sql.Tx, bundle BundlePOSTRequest, team int) error {
	var bundleID int

	// Generate a new bundle_id using a sequence
	err := tx.QueryRow("SELECT nextval('bundle_id_seq')").Scan(&bundleID)
	if err != nil {
		return fmt.Errorf("could not get next bundle id: %v", err)
	}

	for layerNo, productID := range bundle.Products {
		err := insertBundles(tx, bundleID, productID, layerNo+1)
		if err != nil {
			return fmt.Errorf("could not insert bundles for product %d: %v", productID, err)
		}
	}

	err = insertNewProduct(tx, bundle, team)
	if err != nil {
		return fmt.Errorf("could not insert bundles for team %d: %v", team, err)
	}

	return nil
}

func insertNewProduct(tx *sql.Tx, bundle BundlePOSTRequest, team int) error {
	_, err := tx.Exec(`INSERT INTO products (
                	name,
					ean_code,
                    comment,
                    is_public,
					type,
					high_temperature,
					low_temperature,                                        
					testing_team, 
					version,
					status) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);`,
		bundle.ProductName,
		"",
		bundle.Comment,
		bundle.IsPublic,
		"bundle",
		0,
		0,
		team,
		time.Now(),
		bundle.Status)

	return err
}

func validatePermissions(bundle BundlePOSTRequest, team int) error {
	if bundle.IsPublic == true && domain.TeamRole(team) == domain.Researcher {
		return fmt.Errorf("researcher cannot create public tests, %d", http.StatusUnauthorized)
	}
	return nil
}
