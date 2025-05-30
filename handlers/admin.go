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

	// Check admin access
	user, err := GetUserFromSession(r)
	if err != nil || user == nil || !isAdminEmail(user.Gmail) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	r.ParseForm()

	levelNumber := r.FormValue("level_number")
	markdown := r.FormValue("markdown")
	sourceHint := r.FormValue("source_hint")
	consoleHint := r.FormValue("console_hint")
	answer := r.FormValue("answer")
	active := r.FormValue("active") == "true"

	if levelNumber == "" || markdown == "" || answer == "" {
		http.Redirect(w, r, "/admin/levels/new?error=Required fields missing", http.StatusSeeOther)
		return
	}

	levelNum, err := strconv.Atoi(levelNumber)
	if err != nil {
		http.Redirect(w, r, "/admin/levels/new?error=Invalid level number", http.StatusSeeOther)
		return
	}

	newLvl := database.AdminLevel{
		LevelNumber: levelNum,
		Markdown:    markdown,
		SourceHint:  sourceHint,
		ConsoleHint: consoleHint,
		Answer:      answer,
		Active:      active,
	}

	err = database.Create("levels", newLvl)
	if err != nil {
		http.Redirect(w, r, "/admin/levels/new?error=Failed to create level", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin?success=Level created successfully", http.StatusSeeOther)
}

func UpdateLvlHandler(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check admin access
	user, err := GetUserFromSession(r)
	if err != nil || user == nil || !isAdminEmail(user.Gmail) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Redirect(w, r, "/admin?error=Invalid level ID", http.StatusSeeOther)
		return
	}

	r.ParseForm()

	levelNumber := r.FormValue("level_number")
	markdown := r.FormValue("markdown")
	sourceHint := r.FormValue("source_hint")
	consoleHint := r.FormValue("console_hint")
	answer := r.FormValue("answer")
	active := r.FormValue("active") == "true"

	if levelNumber == "" || markdown == "" || answer == "" {
		http.Redirect(w, r, "/admin/levels/"+id+"/edit?error=Required fields missing", http.StatusSeeOther)
		return
	}

	levelNum, err := strconv.Atoi(levelNumber)
	if err != nil {
		http.Redirect(w, r, "/admin/levels/"+id+"/edit?error=Invalid level number", http.StatusSeeOther)
		return
	}

	updatedLvl := database.AdminLevel{
		LevelNumber: levelNum,
		Markdown:    markdown,
		SourceHint:  sourceHint,
		ConsoleHint: consoleHint,
		Answer:      answer,
		Active:      active,
	}

	err = database.Update("level", map[string]interface{}{"number": idInt}, updatedLvl)
	if err != nil {
		http.Redirect(w, r, "/admin/levels/"+id+"/edit?error=Failed to update level", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin?success=Level updated successfully", http.StatusSeeOther)
}

func DeleteLvlHandler(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check admin access
	user, err := GetUserFromSession(r)
	if err != nil || user == nil || !isAdminEmail(user.Gmail) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Redirect(w, r, "/admin?error=Invalid level ID", http.StatusSeeOther)
		return
	}

	err = database.Delete("level", map[string]interface{}{"number": idInt})
	if err != nil {
		http.Redirect(w, r, "/admin?error=Failed to delete level", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin?success=Level deleted successfully", http.StatusSeeOther)
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
	result, err := database.Get("levels", map[string]interface{}{"all": true})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve levels"})
		return
	}
	levels := result.([]database.AdminLevel)

	levelList := make([]map[string]interface{}, len(levels))
	for i, level := range levels {
		levelList[i] = map[string]interface{}{
			"id":       level.LevelNumber,
			"number":   level.LevelNumber,
			"title":    "Level " + strconv.Itoa(level.LevelNumber),
			"question": level.Markdown,
			"answer":   level.Answer,
			"hint":     level.SourceHint,
			"active":   level.Active,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(levelList)
}

func GetStatsHandler(w http.ResponseWriter, r *http.Request) {
	result, err := database.Get("levels", map[string]interface{}{"all": true})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve stats"})
		return
	}
	levels := result.([]database.AdminLevel)

	result, err = database.Get("leaderboard", map[string]interface{}{"limit": 0})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve stats"})
		return
	}
	leaderboard := result.([]database.Sucker)

	activeUsers := 0
	for _, user := range leaderboard {
		if user.Score > 0 {
			activeUsers++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"totalUsers":  len(leaderboard),
		"totalLevels": len(levels),
		"activeUsers": activeUsers,
	})
}

func GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	result, err := database.Get("login", map[string]interface{}{"all": true})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve users"})
		return
	}
	users := result.([]database.Login)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": users,
		"count": len(users),
	})
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request, email string) {
	err := database.Delete("login", map[string]interface{}{"gmail": email})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete user"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted successfully"})
}
