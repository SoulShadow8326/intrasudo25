package handlers

import (
	"intrasudo25/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)



func CreateLvlHandler(c *gin.Context) {
	var newLvl database.Level //NOTE
	if err := c.ShouldBindJSON(&newLvl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := database.CreateLevel(newLvl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create lvl"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Lvl created successfully",
		"lvl":     newLvl,
	})
}

func UpdateLvlHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lvl ID"})
		return
	}

	var updatedLvl database.Level
	if err := c.ShouldBindJSON(&updatedLvl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = database.UpdateLevel(id, updatedLvl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update lvl"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lvl updated successfully",
		"lvl":     updatedLvl,
	})
}

func DeleteLvlHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lvl ID"})
		return
	}

	err = database.DeleteLevel(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete lvl"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lvl deleted successfully",
	})
}

// Admin Panel Handlers - for managing all levels/questions
func AdminPanelHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Admin Panel - Level Management",
		"endpoints": gin.H{
			"GET /api/admin/levels":        "Get all levels",
			"POST /api/admin/levels":       "Create new level",
			"PUT /api/admin/levels/:id":    "Update level",
			"DELETE /api/admin/levels/:id": "Delete level",
		},
	})
}

func GetAllLevelsHandler(c *gin.Context) {
	levels, err := database.GetLevels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve levels"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"levels": levels,
		"count":  len(levels),
	})
}
