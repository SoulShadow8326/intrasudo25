package handlers

import (
	"intrasudo25/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func LeaderboardPage(c *gin.Context) {
	top, err := database.GetLeaderboardTop(10)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching leaderboard")
		return
	}

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

	c.HTML(http.StatusOK, "leaderboard.html", gin.H{
		"entries": entries,
	})
}
