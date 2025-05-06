package registrationHandler

import (
	"backend/internal/resources"
	"backend/internal/utils"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

// RegistrationHandler handles user registration requests.
//
//	@Summary		Register user
//	@Description	Registers a new user
//	@Tags			Registration
//	@Accept			json
//	@Produce		json
//	@Param			credentials	body	RegistrationPOSTRequest	true	"User credentials"
//	@Success		201			"User created successfully"
//	@Failure		400			{string}	string	"Invalid request body"
//	@Failure		405			{string}	string	"Method not allowed"
//	@Failure		409			{string}	string	"User already exists"
//	@Failure		500			{string}	string	"Could not create user"
//	@Router			/register/ [post]
func RegistrationHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.Method {
		case http.MethodPost:
			RegistrationRequestPOST(w, r, db)
			return
		default:
			http.Error(w, resources.MethodNotAllowed, http.StatusNotImplemented)
			return
		}
	}
}

func RegistrationRequestPOST(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	credentials, err := utils.ParseAndValidateRequest[RegistrationPOSTRequest](r)
	if err != nil {
		http.Error(w, resources.InvalidPOSTRequest, http.StatusBadRequest)
		log.Println(resources.InvalidPOSTRequest + ": " + err.Error())
		return
	}

	var credJSON []byte
	credJSON, err = json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		http.Error(w, "Could not parse request body.", http.StatusInternalServerError)
		log.Println("Could not Marshal JSON:", string(credJSON))
	}

	var tx *sql.Tx
	tx, err = db.Begin()
	if err != nil {
		http.Error(w, resources.TransactionStartFailed, http.StatusInternalServerError)
		log.Println(resources.TransactionStartFailed + ": " + err.Error())
		return
	}

	rollbackOnError := true
	defer func() {
		if rollbackOnError {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				http.Error(w, resources.RollbackFailed, http.StatusInternalServerError)
				log.Println(resources.RollbackFailed + rollbackErr.Error())
			}
		}
	}()

	// Handle registration process
	_, err = registerUserAndTeam(tx, credentials)
	if err != nil {
		// Check for specific errors to return appropriate status codes
		if err.Error() == "user already exists" {
			http.Error(w, err.Error(), http.StatusConflict)
			log.Println("User already exists: " + err.Error())
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Failed to register user: " + err.Error())
		}
		return
	}

	//Commit transaction
	err = tx.Commit()
	if err != nil {
		http.Error(w, "Failed to commit the transaction.", http.StatusInternalServerError)
		log.Println("Failed to commit the transaction: " + err.Error())
		return
	}

	// Transaction was committed successfully, don't roll back
	rollbackOnError = false

	w.WriteHeader(http.StatusCreated)
}
