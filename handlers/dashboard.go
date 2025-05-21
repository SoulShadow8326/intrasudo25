package handlers

import (
	"intrasudo25/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func DashboardPage(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Dashboard"})
}

func GetQuestionHandler(c *gin.Context) {
	var id uint;
	if is_there, login := Authorize(c); is_there {
		id = login.On
	} else{
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login pending..."})
		return
	}

	question, err := database.GetQuestion(int(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"question": question,
	})
}

func CreateQuestionHandler(c *gin.Context) {
	var newQuestion database.Question
	if err := c.ShouldBindJSON(&newQuestion); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := database.CreateQuestion(newQuestion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create question"})
		return
	}

	newQuestion.ID = int(id)

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Question created successfully",
		"question": newQuestion,
	})
}

func UpdateQuestionHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
		return
	}

	var updatedQuestion database.Question
	if err := c.ShouldBindJSON(&updatedQuestion); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedQuestion.ID = id

	err = database.UpdateQuestion(updatedQuestion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update question"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Question updated successfully",
		"question": updatedQuestion,
	})
}

func DeleteQuestionHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
		return
	}

	err = database.DeleteQuestion(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete question"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Question deleted successfully",
	})
}
