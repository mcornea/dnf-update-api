package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/mcornea/dnf-update-api/pkg/updater"
)

type UpdateResponse struct {
	Status       string   `json:"status"`
	Updates      []string `json:"updates,omitempty"`
	ErrorMessage string   `json:"error,omitempty"`
}

func main() {
	// Get configuration from environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	apiToken := os.Getenv("API_TOKEN")
	if apiToken == "" {
		log.Fatal("API_TOKEN environment variable must be set")
	}

	// Define API routes
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/updates", tokenAuthMiddleware(apiToken, updatesHandler))
	http.HandleFunc("/api/upgrade", tokenAuthMiddleware(apiToken, upgradeHandler))

	// Start server
	log.Printf("Starting DNF Update API server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Basic health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

// Middleware to authenticate API token
func tokenAuthMiddleware(token string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized: Missing or invalid authorization header", http.StatusUnauthorized)
			return
		}

		providedToken := strings.TrimPrefix(authHeader, "Bearer ")
		if providedToken != token {
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

// Handler for listing available updates
func updatesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	updates, err := updater.CheckUpdates()
	if err != nil {
		sendErrorResponse(w, "Failed to check for updates", err, http.StatusInternalServerError)
		return
	}

	response := UpdateResponse{
		Status:  "NO_UPDATES",
		Updates: updates,
	}

	if len(updates) > 0 {
		response.Status = "AVAILABLE_UPDATES"
	}

	sendJSONResponse(w, response)
}

// Handler for triggering system upgrade
func upgradeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := updater.UpgradePackages()
	if err != nil {
		sendErrorResponse(w, "Failed to upgrade packages", err, http.StatusInternalServerError)
		return
	}

	response := UpdateResponse{
		Status: "UPGRADE_SUCCESSFUL",
	}
	sendJSONResponse(w, response)
}

// Helper function to send JSON responses
func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// Helper function to send error responses
func sendErrorResponse(w http.ResponseWriter, message string, err error, statusCode int) {
	log.Printf("%s: %v", message, err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := UpdateResponse{
		Status:       "ERROR",
		ErrorMessage: fmt.Sprintf("%s: %v", message, err),
	}
	json.NewEncoder(w).Encode(response)
}
