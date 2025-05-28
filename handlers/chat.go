package handlers

import (
	"encoding/json"
	"net/http"
)

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Live chat coming soon!"})
}
