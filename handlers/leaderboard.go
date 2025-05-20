package handlers

import (
	"net/http"
	"strconv"

	"intrasudo25/database"

	"github.com/gin-gonic/gin"
)

func LeaderboardPage(c *gin.Context) {
	// Optional: support query param ?n=20
	limitStr := c.DefaultQuery("n", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return
	}

	entries, err := database.GetLeaderboardTop(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch leaderboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"leaderboard": entries})
}
