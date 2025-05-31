package handlers

import (
	"encoding/json"
	"intrasudo25/database"
	"net/http"
	"strconv"
)

func LeaderboardPage(w http.ResponseWriter, r *http.Request) {
	result, err := database.Get("leaderboard", map[string]interface{}{"limit": 0})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Error fetching leaderboard: " + err.Error()})
		return
	}
	top := result.([]database.Sucker)

	type Entry struct {
		Gmail string
		Score string
		On    uint
	}

	var entries []Entry
	for _, e := range top {
		entries = append(entries, Entry{
			Gmail: e.Gmail,
			Score: strconv.Itoa(e.Score),
			On:    e.On,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"leaderboard": entries,
		"count":       len(entries),
	})
}
