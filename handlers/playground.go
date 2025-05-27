package handlers

import (
	"intrasudo25/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetQuestionHandler(c *gin.Context) {
	var id uint;
	if is_there, login := Authorize(c); is_there {
		id = login.On
	} else{
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

func submitAnswer(c *gin.Context) bool {
	ans := c.PostForm("answer")
	var User Login
	if isFound, user := Authorize(c); isFound {
		User = user
	} else {
		return false
	}

	levels, err := database.GetLevels()
	if err != nil {
		return false
	}

	if int(User.On) >= len(levels) {
		return false
	}

	level, err := database.GetLevel(int(User.On))
	if err != nil {
		return false
	}

	if level.Answer == ans {
		User.On += 1

		err = database.UpdateField(User.Gmail, "On", User.On)
		if err != nil {
			return false
		}
		return true
	}

	return false
}

