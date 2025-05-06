package productsHandler

import (
	"backend/internal/domain"
	"database/sql"
	"time"
)

// insertNewProduct handles inserting a new product into the database
func insertNewProduct(db *sql.DB, product ProductPOSTRequest, team int) error {
	_, err := db.Exec(`INSERT INTO products (
                	name,
                    brand,
					ean_code,
                    image_url,
                    comment,
                    is_public,
					type,
					high_temperature,
					low_temperature,                                        
					testing_team, 
					version,
					status) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);`,
		product.Name,
		product.Brand,
		product.EANCode,
		product.ImageURL,
		product.Comment,
		product.IsPublic,
		product.Type,
		product.HighTemperature,
		product.LowTemperature,
		team,
		time.Now(),
		product.Status)

	return err
}

func ScanProducts(rows *sql.Rows) ([]domain.Product, error) {
	var products []domain.Product
	for rows.Next() {
		var product domain.Product

		if err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Brand,
			&product.EANCode,
			&product.ImageURL,
			&product.Comment,
			&product.IsPublic,
			&product.Type,
			&product.HighTemperature,
			&product.LowTemperature,
			&product.TestingTeam,
			&product.Version,
			&product.Status); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

// queryAndScanProduct executes a query and scans the result into a Product struct.
func queryAndScanProduct(db *sql.DB, query string, args ...interface{}) (domain.Product, error) {
	var product domain.Product
	err := db.QueryRow(query, args...).Scan(
		&product.ID,
		&product.Name,
		&product.Brand,
		&product.EANCode,
		&product.ImageURL,
		&product.Comment,
		&product.IsPublic,
		&product.Type,
		&product.HighTemperature,
		&product.LowTemperature,
		&product.TestingTeam,
		&product.Version,
		&product.Status)

	return product, err
}
