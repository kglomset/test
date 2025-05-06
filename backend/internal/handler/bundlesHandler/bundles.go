package bundlesHandler

//coverage:ignore file
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
	"strings"
)

// BundlesHandler routes HTTP requests for bundles to the appropriate handler function.
//
// It supports the following methods:
// - GET: Retrieves a list of bundles.
//
// Each method has its own dedicated request handler function, which is documented separately using Swagger annotations.
func BundlesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.Method {
		case http.MethodGet:
			BundlesRequestGET(w, r, db)
		case http.MethodPost:
			BundlesRequestPOST(w, r, db)
		default:
			http.Error(w, resources.MethodNotAllowed, http.StatusNotImplemented)
			return
		}
	}
}

// BundlesRequestGET handles GET requests for bundles.
//
//	@Summary		Get a list of bundles
//	@Description	Retrieves a list of bundles based on query parameters.
//	@Tags			Bundles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	query		int						false	"Bundle ID"
//	@Success		200	{array}		domain.ProductBundle	"Successful response with a list of bundles"
//	@Failure		500	{string}	string					"Could not retrieve all bundles."
//	@Router			/bundles [get]
//	@Router			/bundles/ [get]
func BundlesRequestGET(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Retrieve the part of the URL that contains the bundle id.
	idParam := strings.TrimPrefix(r.URL.Path, "/bundles/")
	idStr := regexp.MustCompile(`\d+`).FindString(idParam)

	var id int
	// Get the product ID from the URL query parameter.
	if idStr != "" {
		id, _ = utils.GetIDFromURLQuery(w, idStr)
	}

	var bundles []domain.ProductBundle

	// Check if the query does not contain an id.
	if id == 0 {
		bundles = getAllBundles(w, db, bundles)
	} else {
		bundles = getBundlesByBundleID(w, db, bundles, id, idStr)
	}

	// Check if no bundles were found.
	if len(bundles) == 0 {
		http.Error(w, "No bundles found.", http.StatusOK)
		return
	}

	// Write the response to the client.
	if err := json.NewEncoder(w).Encode(bundles); err != nil {
		http.Error(w, "Could not encode bundles.", http.StatusInternalServerError)
		log.Println("Could not encode bundles.", err.Error())
		return
	}
}

// BundlesRequestPOST handles POST requests for creating new bundles.
//
//	@Summary		Create a new bundle
//	@Description	Creates a new bundle based on the provided JSON request body.
//	@Tags			Bundles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			bundle	body		domain.ProductBundle	true	"Bundle details"
//	@Success		201		{string}	string					"Bundle created successfully"
//	@Failure		400		{string}	string					"Invalid request body"
//	@Failure		401		{string}	string					"Unauthorized"
//	@Failure		500		{string}	string					"Failed to commit the transaction."
//	@Router			/bundles [post]
//	@Router			/bundles/ [post]
func BundlesRequestPOST(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	team := middleware.GetUserTeamRole(w, r, db)

	bundle, err := utils.ParseAndValidateRequest[BundlePOSTRequest](r)
	bundleJSON, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		log.Println("Error marshaling test:", err)
	} else {
		log.Println("Test JSON:", string(bundleJSON))
	}
	if err != nil {
		http.Error(w, resources.InvalidPOSTRequest, http.StatusBadRequest)
		log.Println(resources.InvalidPOSTRequest + ": " + err.Error())
		return
	}

	// Validate role permissions
	if permErr := validatePermissions(bundle, team); permErr != nil {
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

	err = createBundles(tx, bundle, team)
	if err != nil {
		http.Error(w, resources.InvalidPOSTRequest, http.StatusBadRequest)
		log.Println(resources.InvalidPOSTRequest + ": " + err.Error())
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
