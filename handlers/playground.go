package handlers

import (
	"intrasudo25/database"
	"net/http"
	"strings"

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

func SubmitAnswer(c *gin.Context) {
	// Authorize the user
	is_there, login := Authorize(c)
	if !is_there {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login pending..."})
		return
	}

	// Get the user's submitted answer
	userAnswer := c.PostForm("answer")
	if userAnswer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Answer is required"})
		return
	}

	// Check the level they are on
	currentLevel := int(login.On)
	level, err := database.GetLevel(currentLevel)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Level not found"})
		return
	}

	// Compare their answer with level answer
	userAnswerTrimmed := strings.TrimSpace(strings.ToLower(userAnswer))
	correctAnswerTrimmed := strings.TrimSpace(strings.ToLower(level.Answer))

	if userAnswerTrimmed == correctAnswerTrimmed {
		// Correct answer - advance to next level
		nextLevel := currentLevel + 1
		err = database.UpdateField(login.Gmail, "On", nextLevel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update progress"})
			return
		}

		// Update score
		currentScore, _ := database.GetUserScore(login.Gmail)
		newScore := currentScore + 10 // 10 points per correct answer
		database.UpdateScore(login.Gmail, newScore)

		c.JSON(http.StatusOK, gin.H{
			"message":    "Correct answer!",
			"correct":    true,
			"next_level": nextLevel,
			"score":      newScore,
		})
	} else {
		// Wrong answer
		c.JSON(http.StatusOK, gin.H{
			"message":       "Incorrect answer. Try again!",
			"correct":       false,
			"current_level": currentLevel,
		})
	}
}
