package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/shaadcodes/ticket-system/models"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := models.HealthResponse{Status: "ok"}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "\n\nFailed to encode response!\n\n", http.StatusInternalServerError)
	}
}
