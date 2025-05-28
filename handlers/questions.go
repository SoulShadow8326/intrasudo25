package handlers

import (
	"encoding/json"
	"net/http"
)

func AttemptQuestionPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Attempt your questions here."})
}
