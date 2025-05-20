package handlers

import "github.com/gin-gonic/gin"

func LeaderboardPage(c *gin.Context) {
	c.JSON(200, gin.H{"message": "This is the leaderboard."})
}
