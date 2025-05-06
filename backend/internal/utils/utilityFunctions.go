package utils

//coverage:ignore file
import (
	"backend/internal/resources"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-playground/validator/v10"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func ParseVersionTimestamps(w http.ResponseWriter,
	updateRequestVersion time.Time, existingEntityVersion time.Time) (time.Time, time.Time) {

	// Update the version of the update request, and parse the time.
	updateRequestVersion = time.Now().Round(0)

	// Version timestamp format
	timestampFormat := "2006-01-02 15:04:05.999999"

	t1String := existingEntityVersion.Format(timestampFormat)
	t2String := updateRequestVersion.Format(timestampFormat)

	// Parse the time of the existing entity version.
	t1, errT1 := time.Parse(timestampFormat, t1String)
	if errT1 != nil {
		http.Error(w, "Error parsing time.", http.StatusInternalServerError)
		log.Println("Error parsing time of the existing entity version:", errT1)
		return time.Time{}, time.Time{}
	}

	// Parse the time of the new update version.
	t2, errT2 := time.Parse(timestampFormat, t2String)
	if errT2 != nil {
		http.Error(w, "Error parsing time.", http.StatusInternalServerError)
		log.Println("Error parsing time of the new update version:", errT2)
		return time.Time{}, time.Time{}
	}

	return t1, t2
}

// CreateUpdateQuery creates the query for updating an entity in the database.
func CreateUpdateQuery(w http.ResponseWriter, entityUpdates map[string]interface{}, updatedFields []string,
	newValues []interface{}, i int) ([]string, []interface{}, int) {
	// Loop through the fields in the entityUpdates map and add them to the updatedFields and newValues arrays.
	for field, value := range entityUpdates {
		updatedFields = append(updatedFields, fmt.Sprintf("%s = $%d", field, i)) // field = $x for the db query.
		newValues = append(newValues, value)
		i++
	}

	// If the update request contains no fields, return an error.
	if len(updatedFields) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		log.Println("No fields to update")
		return nil, nil, 0
	}

	// Add the version field to the updatedFields and newValues arrays.
	updatedFields = append(updatedFields, "version = $"+strconv.Itoa(i))
	newValues = append(newValues, time.Now())
	i++

	return updatedFields, newValues, i
}

// GetIDFromURLQuery retrieves the ID from the URL query parameter.
func GetIDFromURLQuery(w http.ResponseWriter, idParam string) (int, error) {
	// Get the ID from the URL parameter.
	if idParam == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		log.Println("Missing ID")
		return 0, fmt.Errorf("missing ID")
	}

	// Parse the ID as int from the URL parameter.
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		log.Println("Invalid ID: " + err.Error())
		return 0, fmt.Errorf("invalid ID")
	}

	return id, nil
}

func DecodeRequestBody(w http.ResponseWriter, r *http.Request, entity any) error {
	// Decode the request body into the entity struct.
	err := json.NewDecoder(r.Body).Decode(entity)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		log.Println("Invalid request body: " + err.Error())
		return err
	}
	return nil
}

func InitMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}

	return mockDB, mock
}

// Parse and validate the incoming request for any struct type
func ParseAndValidateRequest[T any](r *http.Request) (T, error) {
	var data T
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return data, fmt.Errorf("%s: %s", resources.InvalidPOSTRequest, err.Error())
	}

	validate := validator.New()
	err = validate.Struct(data)
	if err != nil {
		return data, err
	}
	return data, nil
}

// GetClientIPFromRequest retrieves the client's IP address from the request.
func GetClientIPFromRequest(r *http.Request) string {
	// Get the remote address of the request
	address := r.RemoteAddr

	// Check if the X-Forwarded-For header is set
	if forwardedIP := r.Header.Get("X-Forwarded-For"); forwardedIP != "" {
		ips := strings.Split(forwardedIP, ",")
		address = strings.TrimSpace(ips[0]) // Take the first IP
	}

	// Extract the IP if the address contains a port
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		host = address
	}

	// Parse the IP
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() && ip.To4() == nil {
		// Convert IPv6 loopback (::1) to IPv4 (127.0.0.1)
		host = "127.0.0.1"
	}

	// Check if the IP is a valid IPv4 address
	return host
}
