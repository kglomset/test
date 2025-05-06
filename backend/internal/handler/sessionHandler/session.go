package sessionHandler

/*
// IsSessionActive is a dummy endpoint which always returns a json struct where active = true
func IsSessionActive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// The idea is that this is a ping endpoint, if wrapped in auth, it can tell if a
	// user is authenticated or not.
	response := map[string]bool{"active": true}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}*/
