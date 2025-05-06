package rankingsHandler

//coverage:ignore file
import (
	"backend/internal/domain"
	"backend/internal/middleware"
	"backend/internal/resources"
	"backend/internal/utils"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"
)

// RankingsHandler routes HTTP requests for rankings to the appropriate handler function.
//
// It supports the following methods:
// - GET: Retrieves a list of rankings based on filters.
// - POST: Creates a new ranking.
func RankingsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.Method {
		case http.MethodGet:
			RankingsRequestGET(w, r, db)
		case http.MethodPost:
			RankingsRequestPOST(w, r, db)
		case http.MethodPatch:
			RankingsRequestPATCH(w, r, db)
		default:
			http.Error(w, resources.MethodNotAllowed, http.StatusNotImplemented)
			return
		}
	}
}

// RankingsRequestGET handles GET requests for rankings.
//
//	@Summary		Get a list of rankings
//	@Description	Retrieves a list of rankings based on query parameters.
//	@Tags			Rankings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			public	query		string			false	"Public rankings"
//	@Param			test_id	query		int				false	"Test ID"
//	@Param			rank	query		int				false	"Rank"
//	@Success		200		{array}		domain.TestRank	"Successful response with a list of rankings"
//	@Failure		400		{string}	string			"Invalid test_id parameter"
//	@Failure		400		{string}	string			"Invalid rank parameter"
//	@Failure		500		{string}	string			"Could not retrieve all rankings"
//	@Router			/rankings [get]
//	@Router			/rankings/{rank_id} [get]
func RankingsRequestGET(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	public := r.URL.Query().Get("public")
	testIDParam := r.URL.Query().Get("test_id")
	rankParam := r.URL.Query().Get("rank")

	var testRanks []domain.TestRank

	// Get all public testRanks.
	if public == "true" {
		rows, err := db.Query("SELECT * FROM testRanks WHERE is_rank_public = $1;",
			true)
		FetchRankings(w, testRanks, rows, err)
		return
	}

	if testIDParam != "" && rankParam != "" {
		// Convert the testID and rank parameters to integers.
		testID, err1 := strconv.Atoi(testIDParam)
		rank, err2 := strconv.Atoi(rankParam)

		// Check if the testID parameter is invalid.
		if err1 != nil {
			http.Error(w, "Invalid test_id parameter)!", http.StatusBadRequest)
			log.Println("Invalid test_id parameter: " + err1.Error())
			return
		}

		// Check if the rank parameter is invalid.
		if err2 != nil {
			http.Error(w, "Invalid rank parameter!", http.StatusBadRequest)
			log.Println("Invalid rank parameter: " + err2.Error())
			return
		}

		// Fetch the testRanks with the given testID and rank from the database.
		rows, err := db.Query("SELECT * FROM testRanks WHERE test_id = $1 AND rank = $2;", testID, rank)
		FetchRankings(w, testRanks, rows, err)
		return
	}

	// Get the user's team membership.
	team := middleware.GetUserTeamRole(w, r, db)

	// Fetch the product with the given name from the database based on the user's team membership.
	rows, err := db.Query("SELECT * FROM testRanks WHERE testing_team = $1;", team)
	FetchRankings(w, testRanks, rows, err)
	return
}

// RankingsRequestPOST handles POST requests for rankings.
//
//	@Summary		Create a new ranking
//	@Description	Creates a new ranking
//	@Tags			Rankings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			ranking	body	RankingsPOSTRequest	true	"New ranking information"
//	@Success		201		"Ranking created successfully"
//	@Failure		400		{string}	string	"Invalid request body"
//	@Failure		409		{string}	string	"Product is already part of the test"
//	@Failure		500		{string}	string	"Could not create ranking"
//	@Router			/rankings [post]
//	@Router			/rankings/ [post]
func RankingsRequestPOST(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	ranking, err := utils.ParseAndValidateRequest[RankingsPOSTRequest](r)
	if err != nil {
		http.Error(w, resources.InvalidPOSTRequest, http.StatusBadRequest)
		log.Println(resources.InvalidPOSTRequest + ": " + err.Error())
		return
	}

	var id int
	err = db.QueryRow("SELECT COUNT(*) FROM rankings;").Scan(&id)
	if err != nil {
		http.Error(w, "Ranking creation failed", http.StatusBadRequest)
		log.Println("Unable to get the total ranking count: " + err.Error())
		return
	}

	// Check if the product is already part of a test.
	partOfTest, checkErr := IsProductPartOfTest(db, ranking)
	if checkErr != nil {
		http.Error(w, "Could not create the new ranking. Please try again.", http.StatusInternalServerError)
		log.Println("Could not check if the product is part of the test: " + checkErr.Error())
		return
	}

	if !partOfTest {
		// Check if the new ranking is set to public and the user is a researcher.
		if ranking.IsPublic == true {
			http.Error(w, "Researcher cannot create public rankings", http.StatusBadRequest)
			log.Println("Researcher cannot create public rankings")
			return
		} else {
			_, err = db.Exec(
				`INSERT INTO rankings (
                      test_id, 
                      rank,
                      wins,
                      is_public,
                      version) 
				VALUES ($1, $2, $3, $4, $5);`,
				ranking.TestID,
				ranking.Rank,
				ranking.Wins,
				ranking.IsPublic, time.Now())
			if err != nil {
				http.Error(w, "Could not create ranking.", http.StatusInternalServerError)
				log.Println("Could not create ranking: " + err.Error())
				return
			}

			// Return a successful response.
			w.WriteHeader(http.StatusCreated)
			return
		}
	} else {
		http.Error(w, "The product is already part of this test.", http.StatusConflict)
		log.Println("The product is already part of the test.")
		return
	}
}

func RankingsRequestPATCH(w http.ResponseWriter, r *http.Request, db *sql.DB) {

}
