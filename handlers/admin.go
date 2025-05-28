package handlers

import (
	"encoding/json"
	"intrasudo25/database"
	"net/http"
	"strconv"
)

func CreateLvlHandler(w http.ResponseWriter, r *http.Request) {
	var newLvl database.Level
	if err := json.NewDecoder(r.Body).Decode(&newLvl); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	_, err := database.CreateLevel(newLvl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create lvl"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Lvl created successfully",
		"lvl":     newLvl,
	})
}

func UpdateLvlHandler(w http.ResponseWriter, r *http.Request, id string) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid lvl ID"})
		return
	}

	var updatedLvl database.Level
	if err := json.NewDecoder(r.Body).Decode(&updatedLvl); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	err = database.UpdateLevel(idInt, updatedLvl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update lvl"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Lvl updated successfully",
		"lvl":     updatedLvl,
	})
}

func DeleteLvlHandler(w http.ResponseWriter, r *http.Request, id string) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid lvl ID"})
		return
	}

	err = database.DeleteLevel(idInt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete lvl"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Lvl deleted successfully",
	})
}

func AdminPanelHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Admin Panel - Level Management",
		"endpoints": map[string]string{
			"GET /api/admin/levels":        "Get all levels",
			"POST /api/admin/levels":       "Create new level",
			"PUT /api/admin/levels/:id":    "Update level",
			"DELETE /api/admin/levels/:id": "Delete level",
		},
	})
}

func GetAllLevelsHandler(w http.ResponseWriter, r *http.Request) {
	levels, err := database.GetLevels()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve levels"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"levels": levels,
		"count":  len(levels),
	})
}
