package handlers

import (
	"intrasudo25/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetQuestionHandler(c *gin.Context) {
	var id uint
	if is_there, login := Authorize(c); is_there {
		id = login.On
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login pending..."})
		return
	}

	question, err := database.GetLevel(int(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"question": question,
	})
}
