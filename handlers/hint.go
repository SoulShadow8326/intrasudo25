package handlers

import (
	"encoding/json"
	"intrasudo25/database"
	"net/http"
	"strconv"
	"strings"
)

func GetLevelHintHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	user, err := GetUserFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not authenticated"})
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/user/level-hint/")
	levelNumber, err := strconv.Atoi(path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid level number"})
		return
	}

	currentLevel, err := database.GetCurrentLevelForUser(user.Gmail)
	if err != nil || currentLevel.Number != levelNumber {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied"})
		return
	}

	hint, err := database.GetLevelHint(levelNumber)
	if err != nil {
		hint = ""
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"hint": hint})
}
