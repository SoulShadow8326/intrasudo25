package handlers

import "github.com/gin-gonic/gin"

func AttemptQuestionPage(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Attempt your questions here."})
}
