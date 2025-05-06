package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// in-memory accounts: email -> password
var accounts = map[string]string{
	"oskar@ntnu.no":    "po",
	"emil@ntnu.no":     "pe",
	"kristian@ntnu.no": "pk",
	"dev@ntnu.no":      "pd",
}

// tokenStore maps generated tokens to the account's email
var tokenStore = map[string]string{}

// tokenExpirations maps generated tokens to their expiration time
var tokenExpirations = map[string]time.Time{}

// LoginRequest represents the expected JSON payload for login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the JSON response containing an expiration timestamp and session token.
type LoginResponse struct {
	ExpiresAt    string `json:"expires_at"`
	SessionToken string `json:"session_token"`
}

func main() {
	// protected endpoints using different error codes based on the route
	// /profile returns 401 if not authenticated
	http.HandleFunc("/profile", authMiddleware(http.StatusUnauthorized, handleProfile))
	// any endpoint under /auth/ returns 403 if authentication fails
	http.HandleFunc("/auth/", authMiddleware(http.StatusUnauthorized, authPrefixHandler))

	// unprotected endpoints
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/todos/", handleTodo)
	http.HandleFunc("/delay/", handleDelay)
	http.HandleFunc("/error/", handleError)
	http.HandleFunc("/post", handlePostExample)
	// new endpoint for session validation
	http.HandleFunc("/is-session-active", handleIsSessionActive)
	http.HandleFunc("/", handleRoot)

	port := 9988
	fmt.Printf("Starting test server on port %d...\n", port)
	fmt.Println("Test URLs:")
	fmt.Printf("- Login: http://localhost:%d/login (POST email & password)\n", port)
	fmt.Printf("- Profile (requires valid token, returns 401 if missing): http://localhost:%d/profile\n", port)
	fmt.Printf("- Auth-protected todo (returns 403 if no token): http://localhost:%d/auth/todos/1\n", port)
	fmt.Printf("- Todo (unprotected): http://localhost:%d/todos/1\n", port)
	fmt.Printf("- Delayed response: http://localhost:%d/delay/2/todos/1\n", port)
	fmt.Printf("- Random error: http://localhost:%d/error/50/todos/1\n", port)
	fmt.Printf("- See post data: http://localhost:%d/post\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// authMiddleware is a middleware that checks for a valid token
// if the token is missing or invalid, it returns the specified error code
func authMiddleware(errorCode int, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var token string
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

		if token == "" {
			http.Error(w, "Unauthorized: missing token", errorCode)
			return
		}

		email, ok := tokenStore[token]
		if !ok {
			http.Error(w, "Unauthorized: invalid token", errorCode)
			return
		}

		// save the email in the request context.
		ctx := context.WithValue(r.Context(), "userEmail", email)
		next(w, r.WithContext(ctx))
	}
}

// authPrefixHandler strips the "/auth" prefix from the URL and forwards the request
func authPrefixHandler(w http.ResponseWriter, r *http.Request) {
	// remove the "/auth" prefix.
	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/auth")
	// forward to the default mux.
	http.DefaultServeMux.ServeHTTP(w, r)
}

// handleLogin validates the login credentials and generates a token
func handleLogin(w http.ResponseWriter, r *http.Request) {
	printRequestInfo(r)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate credentials.
	if storedPass, ok := accounts[req.Email]; !ok || storedPass != req.Password {
		// login failure returns 401.
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// generate a token (for testing, a random int converted to a string)
	token := fmt.Sprintf("%d", rng.Int63())
	tokenStore[token] = req.Email

	// set expiration from now
	expiresAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	expTime, err := time.Parse(time.RFC3339, expiresAt)
	if err == nil {
		tokenExpirations[token] = expTime
	}

	resp := LoginResponse{
		ExpiresAt:    expiresAt,
		SessionToken: token,
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		return
	}
}

// handleProfile returns the email of the authenticated user
func handleProfile(w http.ResponseWriter, r *http.Request) {
	email, ok := r.Context().Value("userEmail").(string)
	if !ok {
		http.Error(w, "Unable to determine user", http.StatusInternalServerError)
		return
	}
	response := map[string]string{"email": email}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func printRequestInfo(r *http.Request) {
	fmt.Printf("Received request: %s %s\n", r.Method, r.URL.Path)
	fmt.Println("Headers:")
	for name, values := range r.Header {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", name, value)
		}
	}
	fmt.Println()
}

// handleRoot provides a simple description of the available endpoints
func handleRoot(w http.ResponseWriter, r *http.Request) {
	printRequestInfo(r)

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	_, err := fmt.Fprintf(w, "Test server is running!\n\nEndpoints:\n"+
		"/todos/{id} - Get a todo (unprotected)\n"+
		"/delay/{seconds}/... - Add delay to any endpoint\n"+
		"/error/{percent}/... - Add random error to any endpoint\n"+
		"/login - POST email and password to login\n"+
		"/profile - Get profile info (returns 401 if not authenticated)\n"+
		"/auth/{...} - Protected endpoints (returns 401 if not authenticated")
	// 403 is used when you are authenticated but do not have access
	if err != nil {
		return
	}
}

// handleTodo returns a simple todos item
func handleTodo(w http.ResponseWriter, r *http.Request) {
	printRequestInfo(r)

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "Missing todo ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	todo := struct {
		ID        int    `json:"id"`
		Title     string `json:"title"`
		Completed bool   `json:"completed"`
	}{
		ID:        id,
		Title:     fmt.Sprintf("Todo item #%d", id),
		Completed: id%2 == 0,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(todo)
	if err != nil {
		return
	}
}

// handleDelay applies an artificial delay to any endpoint
func handleDelay(w http.ResponseWriter, r *http.Request) {
	printRequestInfo(r)

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "Missing delay seconds", http.StatusBadRequest)
		return
	}

	seconds, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, "Invalid delay seconds", http.StatusBadRequest)
		return
	}

	fmt.Printf("Applying delay: %d seconds\n", seconds)
	time.Sleep(time.Duration(seconds) * time.Second)

	newPath := "/" + strings.Join(parts[3:], "/")
	fmt.Printf("Forwarding request to: %s\n", newPath)
	r.URL.Path = newPath
	http.DefaultServeMux.ServeHTTP(w, r)
}

// handleError randomly returns an error based on the provided error percentage
func handleError(w http.ResponseWriter, r *http.Request) {
	printRequestInfo(r)

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "Missing error percentage", http.StatusBadRequest)
		return
	}

	percent, err := strconv.Atoi(parts[2])
	if err != nil || percent < 0 || percent > 100 {
		http.Error(w, "Invalid error percentage (should be 0-100)", http.StatusBadRequest)
		return
	}

	fmt.Printf("Error probability: %d%%\n", percent)

	if rng.Intn(100) < percent {
		statusCodes := []int{400, 401, 403, 404, 429, 500, 502, 503}
		statusCode := statusCodes[rng.Intn(len(statusCodes))]
		fmt.Printf("Generated error: HTTP %d\n", statusCode)
		http.Error(w, fmt.Sprintf("Randomly generated error (HTTP %d)", statusCode), statusCode)
		return
	}

	newPath := "/" + strings.Join(parts[3:], "/")
	fmt.Printf("Forwarding request to: %s\n", newPath)
	r.URL.Path = newPath
	http.DefaultServeMux.ServeHTTP(w, r)
}

// handlePostExample echoes back POST/PUT/PATCH data
func handlePostExample(w http.ResponseWriter, r *http.Request) {
	printRequestInfo(r)

	switch r.Method {
	case "POST", "PUT", "PATCH":
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		fmt.Printf("Received %s request with body:\n%s\n", r.Method, string(body))
		contentType := r.Header.Get("Content-Type")
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write(body)
		if err != nil {
			return
		}
	default:
		w.Header().Set("Allow", "POST, PUT, PATCH")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleIsSessionActive returns json with session active status
func handleIsSessionActive(w http.ResponseWriter, r *http.Request) {
	printRequestInfo(r)

	var token string
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	}

	expTime, exists := tokenExpirations[token]
	active := exists && time.Now().Before(expTime)

	response := map[string]bool{"active": active}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
