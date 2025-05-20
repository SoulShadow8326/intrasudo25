package handlers

import "github.com/gin-gonic/gin"

func DashboardPage(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Dashboard"})
}
