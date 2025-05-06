package services

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
)

// InitDB initialize database connection
func InitDB() *sql.DB {
	envKeys := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"}
	envValues := make(map[string]string)

	for _, key := range envKeys {
		value, exists := os.LookupEnv(key)
		if !exists {
			log.Fatalf("Environment variable %s is missing!", key)
		}
		envValues[key] = value
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		envValues["DB_HOST"], envValues["DB_PORT"], envValues["DB_USER"],
		envValues["DB_PASSWORD"], envValues["DB_NAME"])

	db, err := sql.Open("postgres",
		dsn)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	return db
}
