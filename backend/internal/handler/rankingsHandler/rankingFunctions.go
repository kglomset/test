package rankingsHandler

//coverage:ignore file
import (
	"backend/internal/domain"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

// FetchRankings is a function that fetches all rankings from the database.
func FetchRankings(w http.ResponseWriter, testRanks []domain.TestRank, rows *sql.Rows, err error) {
	for rows.Next() {
		var rank domain.TestRank

		if err = rows.Scan(
			&rank.TestID,
			&rank.ProductID,
			&rank.Rank,
			&rank.DistanceBehind,
			&rank.Version,
			&rank.IsPublic); err != nil {
			http.Error(w, "Could not retrieve all testRanks.", http.StatusInternalServerError)
			log.Println("Could not retrieve all testRanks: " + err.Error())
			return
		}

		// Append the rank to the testRanks slice.
		testRanks = append(testRanks, rank)
	}

	if len(testRanks) == 0 {
		http.Error(w, "No testRanks found.", http.StatusOK)
		log.Println("No testRanks found.")
		return
	}

	// Write the response to the client.
	err = json.NewEncoder(w).Encode(testRanks)
	if err != nil {
		http.Error(w, "Could not encode testRanks.", http.StatusInternalServerError)
		log.Println("Could not encode testRanks: " + err.Error())
		return
	}

	return
}

// Todo: Sjekk om denne metoden faktisk skal brukes eller ikke. Har bare supresset errors fra den enn så lenge
func IsProductPartOfTest(db *sql.DB, ranking RankingsPOSTRequest) (bool, error) {
	// Check if the product is part of the test.
	var counter int
	err := db.QueryRow("SELECT COUNT(*) FROM rankings WHERE product_id = $1 AND test_id = $2;",
		// ranking.ProductID,
		ranking.TestID).Scan(&counter)
	if err != nil {
		return false, err
	}

	// If the product is part of the test, return true.
	if counter != 0 {
		return true, nil
	}

	return false, nil
}

/*
// UpdateAmountOfTests is a function that updates the amount of tests for a product when a ranking is made.
func UpdateAmountOfTests(w http.ResponseWriter, db *sql.DB, ranking RankingsPOSTRequest) {
	// Get the amount of tests for the product.
	var amountOfTests int
	err := db.QueryRow("SELECT amount_of_tests FROM products WHERE id = $1;", ranking.ProductID).Scan(&amountOfTests)
	if err != nil {
		http.Error(w, "Unable to create ranking for this product, please try another one.", http.StatusNotFound)
		log.Println("Could not retrieve the amount of tests for the product: " + err.Error())
		return
	}

	// Update the product's amount of tests.
	_, err = db.Exec("UPDATE products SET amount_of_tests = $1 WHERE id = $2;",
		amountOfTests+1, ranking.ProductID)
	if err != nil {
		http.Error(w, "Unable to create the new ranking. Please try again.", http.StatusInternalServerError)
		log.Println("Could not update the amount of tests for the product: " + err.Error())
		return
	}
}

*/
