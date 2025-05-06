package productsHandler

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
)

// FetchProducts represents the request body of a POST request to create a new product.
func FetchProducts(w http.ResponseWriter, products []domain.Product, rows *sql.Rows, err error) {
	products, err = ScanProducts(rows)
	if err != nil {
		http.Error(w, "Could not retrieve all products.", http.StatusInternalServerError)
		log.Println("Could not retrieve all products.", err.Error())
		return
	}

	if len(products) != 0 {
		// Write the response to the client.
		err = json.NewEncoder(w).Encode(products)
		if err != nil {
			http.Error(w, "Could not encode products.", http.StatusInternalServerError)
			log.Println("Could not encode products.", err.Error())
			return
		}
	} else {
		http.Error(w, "No products found.", http.StatusOK)
		log.Println("No products found.")
		return
	}
}

// GetProductWithEANCode retrieves a product with a specific EAN code from the database.
func GetProductWithEANCode(db *sql.DB, eanCode string, name string, team int) (domain.Product, int) {

	query := "SELECT * FROM products WHERE testing_team = $1"
	args := []interface{}{team}

	if eanCode != "" {
		query += " AND ean_code = $2;"
		args = append(args, eanCode)
	} else {
		query += " AND name = $2;"
		args = append(args, name)
	}

	product, err := queryAndScanProduct(db, query, args...)
	if err != nil {
		log.Printf("Product lookup failed (EAN: %s, Name: %s, Team: %d) - Error: %v", eanCode, name, team, err)
		return domain.Product{}, http.StatusNotFound
	}
	return product, http.StatusConflict
}

// GetProductWithID retrieves a product with a specific ID from the database.
func GetProductWithID(w http.ResponseWriter, db *sql.DB, productID int) domain.Product {
	var product domain.Product
	query := "SELECT * FROM products WHERE id = $1;"
	// Get the existing product from the database.
	product, err := queryAndScanProduct(db, query, productID)
	if err != nil {
		http.Error(w, resources.CouldNotRetrieveProduct, http.StatusNotFound)
		log.Println("Could not retrieve the product: " + err.Error())
		return domain.Product{}
	}
	return product
}

func GetProductFields(w http.ResponseWriter, db *sql.DB, query string, fields []string, productID int, team int) {
	validFields := map[string]bool{
		"id":               true,
		"name":             true,
		"brand":            true,
		"ean_code":         true,
		"image_url":        true,
		"type":             true,
		"high_temperature": true,
		"low_temperature":  true,
		"comment":          true,
		"testing_team":     true,
		"is_public":        true,
		"version":          true,
		"status":           true,
	}

	// Validate fields
	for _, field := range fields {
		if !validFields[field] {
			http.Error(w, fmt.Sprintf("Invalid field: %s", field), http.StatusBadRequest)
			log.Printf("Invalid field: %s", field)
			return
		}
	}

	// If no fields specified, return all fields
	if len(fields) == 0 {
		fields = []string{"*"}
	}

	// Execute the query
	row := db.QueryRow(query, productID, team)

	// Create a map to hold the response values
	responseFields := make(map[string]interface{})

	// Create a slice to hold the values and value pointers
	values := make([]interface{}, len(fields))
	valuePointers := make([]interface{}, len(fields))
	for i := range fields {
		valuePointers[i] = &values[i]
	}

	// Scan the result into the responseFields map
	err := row.Scan(valuePointers...)
	if err != nil {
		http.Error(w, "Could not retrieve the requested product.", http.StatusInternalServerError)
		log.Println("Could not retrieve the requested product.", err.Error())
		return
	}

	// Add the values to the responseFields map
	for i, field := range fields {
		responseFields[field] = values[i]
	}

	err = json.NewEncoder(w).Encode(responseFields)
	if err != nil {
		http.Error(w, "Could not encode the response.", http.StatusInternalServerError)
		log.Println("Could not encode the response.", err.Error())
		return
	}

}

// ValidateProductPATCHRequestBody validates the product request body of a PATCH request.
func ValidateProductPATCHRequestBody(db *sql.DB,
	productUpdateRequest ProductPATCHRequest, existingProduct domain.Product, team int) (error, int) {
	// Validate the productUpdateRequest struct.
	var update ProductUpdateFields
	b, _ := json.Marshal(productUpdateRequest.Updates)
	err := json.Unmarshal(b, &update)
	if err != nil {
		log.Println("Could not decode request body: " + err.Error())
		return fmt.Errorf("could not decode request body, %d", http.StatusInternalServerError), http.StatusInternalServerError
	}

	validate := validator.New()

	// Validate keys
	err = validate.Struct(productUpdateRequest)
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

	// Check if the public product exists.
	if existingProduct.IsPublic == true && productUpdateRequest.Updates["is_public"] == false {
		log.Println("Product is public and cannot be made private.")
		return fmt.Errorf("product is public and cannot be made private, %d", http.StatusBadRequest), http.StatusBadRequest
	}

	// Check if the product is publicly available and the user is a researcher.
	if existingProduct.IsPublic == true && domain.TeamRole(team) == domain.Researcher {
		log.Println("Researcher cannot update public products")
		return fmt.Errorf("researcher cannot update public products, %d", http.StatusUnauthorized), http.StatusUnauthorized
	}

	// Check if the user is authorized to update the product.
	if existingProduct.TestingTeam != team {
		log.Println("User cannot update this product")
		return fmt.Errorf("user cannot update this product, %d", http.StatusUnauthorized), http.StatusUnauthorized
	}

	eanCode, _ := productUpdateRequest.Updates["ean_code"].(string)
	name, _ := productUpdateRequest.Updates["name"].(string)

	// Check if the request wants to update the EAN code or name.
	if eanCode != "" || name != "" {
		foundProduct, code := GetProductWithEANCode(db, eanCode, name, team)

		if (eanCode == foundProduct.EANCode &&
			name == foundProduct.Name) ||
			(name == foundProduct.Name &&
				eanCode == "") {
			log.Println("Update not allowed: this product already exists")
			return fmt.Errorf("this product allready exists, try another name or ean_code, %d",
				code), code
		}
	}
	return nil, 0
}

// isProductUnique checks if the private product is unique in the database, and is not equal to a public product.
func isProductUnique(db *sql.DB, productUpdateRequest ProductPATCHRequest, team int) bool {
	// Get the EAN code and name from the update request.
	eanCode, _ := productUpdateRequest.Updates["ean_code"].(string)
	name, _ := productUpdateRequest.Updates["name"].(string)

	// Check if the product exists.
	_, code := GetProductWithEANCode(db, eanCode, name, team)
	if code == http.StatusConflict {
		return false
	}
	return true
}

// DirectReferenceUpdateProductAppearances updates the appearances of the updated product in the rankings table and
// deletes the private product from the products table.
func DirectReferenceUpdateProductAppearances(w http.ResponseWriter, r *http.Request, db *sql.DB, eanCode string, id int) {
	var publicProduct domain.Product

	// Get the user's role and team.
	team := middleware.GetUserTeamRole(w, r, db)

	privateQuery := "SELECT * FROM products WHERE id = $1 AND ean_code = $2 AND is_public = $3 AND testing_team = $4;"
	// Get the private product from the database.
	privateProduct, err := queryAndScanProduct(db, privateQuery, id, eanCode, false, team)
	if err != nil {
		http.Error(w, resources.CouldNotRetrieveProduct, http.StatusNotFound)
		log.Println("Could not retrieve the private product: " + err.Error())
		return
	}

	publicQuery := "SELECT * FROM products WHERE ean_code = $1 AND is_public = $2;"
	// Get the public product from the database.
	publicProduct, err = queryAndScanProduct(db, publicQuery, eanCode, true)
	if err != nil {
		http.Error(w, resources.CouldNotRetrieveProduct, http.StatusNotFound)
		log.Println("Could not retrieve the public product: " + err.Error())
		return
	}

	// Update the appearances of the updated product in the rankings table.
	_, err = db.Exec("UPDATE test_ranks SET product_id = $1 WHERE product_id = $2;",
		publicProduct.ID, privateProduct.ID)

	// Delete the private product from the database.
	_, err = db.Exec("DELETE FROM products WHERE id = $1 AND testing_team = $2;",
		privateProduct.ID, team)
}
