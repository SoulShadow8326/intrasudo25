package handlers

import (
	"encoding/json"
	"intrasudo25/database"
	"net/http"
	"strconv"
)

func CreateLvlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := GetUserFromSession(r)
	if err != nil || user == nil || !isAdminEmail(user.Gmail) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	var requestData struct {
		LevelNumber string `json:"level_number"`
		Markdown    string `json:"markdown"`
		Answer      string `json:"answer"`
		Active      string `json:"active"`
	}

	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request data"})
		return
	}

	if requestData.LevelNumber == "" || requestData.Markdown == "" || requestData.Answer == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Required fields missing"})
		return
	}

	levelNum, err := strconv.Atoi(requestData.LevelNumber)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid level number"})
		return
	}

	active := requestData.Active == "true"
	err = database.CreateLevelSimple(levelNum, requestData.Markdown, requestData.Answer, active)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create level"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Level created successfully"})
}

func UpdateLvlHandler(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	user, err := GetUserFromSession(r)
	if err != nil || user == nil || !isAdminEmail(user.Gmail) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied"})
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid level ID"})
		return
	}

	var requestData struct {
		Markdown string `json:"markdown"`
		Answer   string `json:"answer"`
		Active   string `json:"active"`
	}

	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request data"})
		return
	}

	if requestData.Markdown == "" || requestData.Answer == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Required fields missing"})
		return
	}

	active := requestData.Active == "true"
	err = database.UpdateLevelSimple(idInt, requestData.Markdown, requestData.Answer, active)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update level"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Level updated successfully"})
}

func DeleteLvlHandler(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	user, err := GetUserFromSession(r)
	if err != nil || user == nil || !isAdminEmail(user.Gmail) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied"})
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid level ID"})
		return
	}

	err = database.DeleteLevelSimple(idInt)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete level"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Level deleted successfully"})
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
	levels, err := database.GetAllLevelsForAdmin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve levels"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(levels)
}

func GetStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := database.GetAdminStats()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve stats"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := database.GetAllUsersForAdmin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve users"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": users,
		"count": len(users),
	})
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request, email string) {
	err := database.DeleteUserSimple(email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete user"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted successfully"})
}

func ToggleLevelStateHandler(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPatch {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	user, err := GetUserFromSession(r)
	if err != nil || user == nil || !isAdminEmail(user.Gmail) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied"})
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid level ID"})
		return
	}

	var requestData struct {
		Enabled bool `json:"enabled"`
	}

	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request data"})
		return
	}

	err = database.ToggleLevelState(idInt, requestData.Enabled)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update level state"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Level state updated successfully"})
}

func ToggleAllLevelsStateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	user, err := GetUserFromSession(r)
	if err != nil || user == nil || !isAdminEmail(user.Gmail) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied"})
		return
	}

	var requestData struct {
		Enabled bool `json:"enabled"`
	}

	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request data"})
		return
	}

	err = database.ToggleAllLevelsState(requestData.Enabled)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update all level states"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "All level states updated successfully"})
}
