package productsHandler

import (
	"backend/internal/domain"
	"backend/internal/middleware"
	"backend/internal/resources"
	"backend/internal/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// ProductsHandler routes HTTP requests for products to the appropriate handler function.
//
// It supports the following methods:
// - GET: Retrieves a list of products.
// - POST: Creates a new product.
// - PATCH: Updates an existing product.
//
// Each method has its own dedicated request handler function, which is documented separately using Swagger annotations.
func ProductsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.Method {
		case http.MethodGet:
			ProductsRequestGET(w, r, db)
		case http.MethodPost:
			ProductsRequestPOST(w, r, db)
		case http.MethodPatch:
			ProductsRequestPATCH(w, r, db)
		default:
			http.Error(w, resources.MethodNotAllowed, http.StatusNotImplemented)
			return
		}
	}
}

// ProductsRequestGET retrieves a list of products from the database and sends a response to the client.
//
//	@Summary		Get a list of products
//	@Description	Retrieves a list of products based on query parameters.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		query		string			false	"Product id"
//	@Param			public	query		string			false	"Publicly Available Product"
//	@Param			fields	query		string			false	"Fields to retrieve"
//	@Success		200		{array}		domain.Product	"Successful response with a list of products"
//	@Failure		500		{string}	string			"Could not retrieve all products."
//	@Router			/products [get]
//	@Router			/products/{products_id} [get]
func ProductsRequestGET(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Retrieve the part of the URL that contains the product id.
	idParam := strings.TrimPrefix(r.URL.Path, "/products/") // /products/1
	idStr := regexp.MustCompile(`\d+`).FindString(idParam)

	// Get the public parameter from the URL query parameter.
	public := r.URL.Query().Get("public") // /products?public=true

	// Get the fields parameter from the URL query parameter.
	fields := strings.Split(r.URL.Query().Get("fields"), ",") // /products/1?fields=name,description

	var id int
	// Get the product ID from the URL query parameter.
	if idStr != "" {
		id, _ = utils.GetIDFromURLQuery(w, idStr)
	}

	var products []domain.Product

	// Check if the query contains the public parameter.
	if public == "true" {
		// Fetch all the public products from the database.
		rows, err := db.Query(
			"SELECT * FROM products WHERE is_public = $1;", true)
		FetchProducts(w, products, rows, err)
		return
	}

	// Get the user's team membership.
	team := middleware.GetUserTeamRole(w, r, db)

	// Check if the query does not contain an id.
	if id == 0 {
		// Fetch all the products from the database (both private and public) for the different authenticated user's
		// team memberships.
		rows, err := db.Query("SELECT * FROM products WHERE testing_team = $1;", team)
		FetchProducts(w, products, rows, err)
		return
	}

	// Fetch specific fields of the product from the database.
	if len(fields) != 0 && fields[0] != "" {
		// Construct the query
		query := fmt.Sprintf(`SELECT %s FROM products WHERE id = $1 AND testing_team = $2;`,
			strings.Join(fields, ", "))
		GetProductFields(w, db, query, fields, id, team)
		return
	}

	// Fetch the product with the given id from the database based on the user's role and team membership.
	rows, err := db.Query("SELECT * FROM products WHERE id = $1 AND testing_team = $2;",
		id, team)
	FetchProducts(w, products, rows, err)
	return
}

// ProductsRequestPOST creates a new product in the database and sends a response to the client.
//
//	@Summary		Create a new product
//	@Description	Adds a new product to the database.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			product	body	ProductPOSTRequest	true	"New product information"
//	@Success		201		"Product created successfully"
//	@Failure		400		{string}	string	"Invalid request data"
//	@Failure		403		{string}	string	"Researcher cannot create public products"
//	@Failure		409		{string}	string	"Product already exists"
//	@Failure		500		{string}	string	"Could not create product."
//	@Router			/products [post]
//	@Router			/products/ [post]
func ProductsRequestPOST(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get the user's team membership.
	team := middleware.GetUserTeamRole(w, r, db)

	// Decode the request body into the product struct.
	product, err := utils.ParseAndValidateRequest[ProductPOSTRequest](r)
	if err != nil {
		http.Error(w, resources.InvalidPOSTRequest, http.StatusBadRequest)
		log.Println(resources.InvalidPOSTRequest + ": " + err.Error())
		return
	}

	// Get the total product count.
	var id int
	err = db.QueryRow("SELECT COUNT(*) FROM products;").Scan(&id)
	if err != nil {
		http.Error(w, "Product creation failed", http.StatusBadRequest)
		log.Println("Unable to get the total product count: " + err.Error())
		return
	}

	// Get the existing product from the database.
	existingProduct, code := GetProductWithEANCode(db, product.EANCode, product.Name, team)

	// Check if the product already exists in the database.
	if existingProduct.EANCode == product.EANCode && existingProduct.EANCode != "" {
		http.Error(w, "A product with this EAN code already exists", http.StatusConflict)
		log.Println("A product with this EAN code already exists")
		return
	}

	// if product is not present in the database from before then it can be added in this manner
	if code == http.StatusNotFound {
		// Check if the product is set to public and the user is a researcher.
		if product.IsPublic == true && domain.TeamRole(team) == domain.Researcher {
			http.Error(w, "Researcher cannot create public products", http.StatusForbidden)
			log.Println("Researcher cannot create public products")
			return
		} else {
			// Insert the new product into the database.
			err = insertNewProduct(db, product, team)
			if err != nil {
				http.Error(w, "Unable to add new product", http.StatusBadRequest)
				log.Println("Unable to add new product: " + err.Error())
				return
			}
		}
	} else if code == http.StatusConflict {
		http.Error(w, "This Product already exists, try another name if the EAN code is blank", http.StatusConflict)
		log.Println("This Product already exists, try another name if the EAN code is blank")
		return
	}

	// Return a successful response.
	w.WriteHeader(http.StatusCreated)
	return
}

// ProductsRequestPATCH updates an existing product in the database and sends a response to the client.
//
//	@Summary		Update a product
//	@Description	Updates fields of an existing product.
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		query		int					true	"Product ID"
//	@Param			product	body		ProductPATCHRequest	true	"Product update fields"
//	@Success		200		{string}	string				"Product updated successfully"
//	@Failure		400		{string}	string				"Invalid request data"
//	@Failure		500		{string}	string				"Could not update product."
//	@Failure		409		{string}	string				"Detected a conflict for the current product, please refresh."
//	@Router			/products [patch]
func ProductsRequestPATCH(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get the product ID from the URL parameter.
	idParam := strings.TrimPrefix(r.URL.Path, "/products/")
	productID, err := utils.GetIDFromURLQuery(w, idParam)

	// Get the team role of the user from the session token.
	team := middleware.GetUserTeamRole(w, r, db)

	// Decode the request body into the productUpdateRequest struct.
	var productUpdateRequest ProductPATCHRequest
	err = utils.DecodeRequestBody(w, r, &productUpdateRequest)
	if err != nil {
		http.Error(w, resources.InvalidPATCHRequest, http.StatusBadRequest)
		log.Println(resources.InvalidPATCHRequest + ": " + err.Error())
		return
	}

	// Get the existing product from the database.
	existingProduct := GetProductWithID(w, db, productID)

	// Validate the PATCH request body
	var code int
	err, code = ValidateProductPATCHRequestBody(db, productUpdateRequest, existingProduct, team)
	if err != nil {
		http.Error(w, "Validation error: "+err.Error(), code)
		return
	}

	// Parse the version timestamps for the existing product and the product update request.
	existingProductVersion, productUpdateVersion := utils.ParseVersionTimestamps(w,
		productUpdateRequest.Version, existingProduct.Version)

	// Solve the concurrency challenge by checking if the product has been updated since the last sync.
	if existingProductVersion.After(productUpdateVersion) {
		http.Error(w, "Detected a conflict for the current product, please refresh.", http.StatusConflict)
		log.Println("Detected a conflict for the current product, please refresh.")
		return
	}

	// Create the updatedFields and newValues arrays for the query, and increment the index for the newValues array.
	var updatedFields []string
	var newValues []interface{}
	i := 1 // Index for the newValues array.
	updatedFields, newValues, i = utils.CreateUpdateQuery(w, productUpdateRequest.Updates, updatedFields, newValues, i)

	var newVersion time.Time

	// Check if the product is being updated to a public product and if the product is unique,
	// which will not trigger the DirectReferenceUpdateProductAppearances function.
	uniqueProduct := isProductUnique(db, productUpdateRequest, team)

	// if a private product that already exists as a public product is being updated to a public product.
	if !uniqueProduct && ((productUpdateRequest.Updates["is_public"] != existingProduct.IsPublic) &&
		productUpdateRequest.Updates["is_public"] != nil) {
		// Direct reference update for products, tests, and rankings that an updated product is part of.
		DirectReferenceUpdateProductAppearances(w, r, db, existingProduct.EANCode, productID)
		newVersion = productUpdateVersion
	} else {
		// Create the query to update the product in the database.
		query := fmt.Sprintf("UPDATE products SET %s WHERE id = $%d AND version = $%d RETURNING version",
			strings.Join(updatedFields, ", "), // updatedFields parameter ($x0, $x1, ...)
			i,                                 // productID parameter ($xn), where n is the number of
			// indices for the updatedFields array in the query.
			i+1, // Version parameter, gets the next index after the productID.
		)
		newValues = append(newValues, productID, existingProductVersion)

		// Execute the query and get the new version of the product.
		err = db.QueryRow(query, newValues...).Scan(&newVersion)
		if err != nil {
			http.Error(w, "Could not update the product because of a conflict, please refresh.", http.StatusConflict)
			log.Println("Could not update the product because of a conflict, please refresh: " + err.Error())
		}
	}

	// Send a response if the product update was successful.
	if !newVersion.IsZero() {
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Product updated successfully",
			"version": newVersion,
		})
		if err != nil {
			http.Error(w, "Could not JSON encode the response.", http.StatusInternalServerError)
			log.Println("Could not JSON encode the response: " + err.Error())
			return
		}
	}
}
