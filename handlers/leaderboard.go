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
		w.Write([]byte("Error fetching leaderboard"))
		return
	}
	top := result.([]database.Sucker)

	type Entry struct {
		Gmail string
		Score string
	}

	var entries []Entry
	for _, e := range top {
		entries = append(entries, Entry{
			Gmail: e.Gmail,
			Score: strconv.Itoa(e.Score),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries": entries,
	})
}
