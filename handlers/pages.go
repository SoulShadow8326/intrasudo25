package handlers

import (
	"encoding/json"
	"html/template"
	"intrasudo25/config"
	"intrasudo25/database"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type PageData struct {
	User           *database.Login
	Level          *database.Level
	Levels         []database.AdminLevel
	Leaderboard    []database.Sucker
	IsAdmin        bool
	CSRFToken      string
	ErrorMessage   string
	SuccessMessage string
}

type AdminPageData struct {
	User           *database.Login
	Levels         []database.AdminLevel
	Users          []database.Login
	Stats          AdminStats
	IsAdmin        bool
	CSRFToken      string
	ErrorMessage   string
	SuccessMessage string
}

type AdminStats struct {
	TotalUsers  int
	TotalLevels int
}

func isAdminEmail(email string) bool {
	adminEmails := config.GetAdminEmails()
	email = strings.ToLower(email)
	for _, adminEmail := range adminEmails {
		if email == strings.ToLower(adminEmail) {
			return true
		}
	}
	return false
}

func renderTemplate(w http.ResponseWriter, templateName string, data PageData) {
	templatePath := filepath.Join("frontend", templateName)

	tmpl := template.New(templateName).Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"isAdminEmail": isAdminEmail,
	})

	tmpl, err := tmpl.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./frontend/index.html")
}

func HintsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./frontend/hints.html")
}

func GuidelinesHandler(w http.ResponseWriter, r *http.Request) {
	_, err := GetUserFromSession(r)
	isLoggedIn := err == nil

	content, err := os.ReadFile("./frontend/guidelines.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	htmlContent := string(content)

	if !isLoggedIn {
		navStart := strings.Index(htmlContent, `<nav class="navbar">`)
		mobileNavStart := strings.Index(htmlContent, `<div class="mobile-nav-menu"`)

		if navStart != -1 && mobileNavStart != -1 {
			mobileNavEnd := strings.Index(htmlContent[mobileNavStart:], `</div>`)
			if mobileNavEnd != -1 {
				mobileNavEnd = mobileNavStart + mobileNavEnd + 6
				beforeNav := htmlContent[:navStart]
				afterMobileNav := htmlContent[mobileNavEnd:]
				htmlContent = beforeNav + afterMobileNav
			}
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlContent))
}

func LeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./frontend/leaderboard.html")
}

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	if !isAdminEmail(user.Gmail) {
		AdminRequiredHandler(w, r)
		return
	}

	levelsData, err := database.Get("levels", map[string]interface{}{})
	var levels []database.AdminLevel
	if err != nil {
		levels = []database.AdminLevel{}
	} else {
		if l, ok := levelsData.([]database.AdminLevel); ok {
			levels = l
		} else {
			levels = []database.AdminLevel{}
		}
	}

	data := PageData{
		User:    user,
		Levels:  levels,
		IsAdmin: true,
	}

	renderTemplate(w, "admin.html", data)
}

func AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromSession(r)
	if err != nil || user == nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	if !isAdminEmail(user.Gmail) {
		AdminRequiredHandler(w, r)
		return
	}

	levelsData, err := database.Get("levels", map[string]interface{}{})
	var levels []database.AdminLevel
	if err != nil {
		levels = []database.AdminLevel{}
	} else {
		if l, ok := levelsData.([]database.AdminLevel); ok {
			levels = l
		} else {
			levels = []database.AdminLevel{}
		}
	}

	usersData, err := database.Get("login", map[string]interface{}{"all": true})
	var users []database.Login
	if err != nil {
		users = []database.Login{}
	} else {
		if u, ok := usersData.([]database.Login); ok {
			users = u
		} else {
			users = []database.Login{}
		}
	}

	stats := AdminStats{
		TotalUsers:  len(users),
		TotalLevels: len(levels),
	}

	errorMsg := r.URL.Query().Get("error")
	successMsg := r.URL.Query().Get("success")

	data := AdminPageData{
		User:           user,
		Levels:         levels,
		Users:          users,
		Stats:          stats,
		IsAdmin:        true,
		ErrorMessage:   errorMsg,
		SuccessMessage: successMsg,
	}

	renderAdminTemplate(w, "admin.html", data)
}

func NewLevelFormHandler(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromSession(r)
	if err != nil || user == nil || !isAdminEmail(user.Gmail) {
		AdminRequiredHandler(w, r)
		return
	}

	errorMsg := r.URL.Query().Get("error")

	data := AdminPageData{
		User:         user,
		IsAdmin:      true,
		ErrorMessage: errorMsg,
	}

	renderAdminTemplate(w, "admin_new_level.html", data)
}

func EditLevelFormHandler(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromSession(r)
	if err != nil || user == nil || !isAdminEmail(user.Gmail) {
		AdminRequiredHandler(w, r)
		return
	}

	levelNumberStr := r.URL.Path[len("/admin/levels/"):]
	levelNumberStr = strings.TrimSuffix(levelNumberStr, "/edit")

	levelNumber, err := strconv.Atoi(levelNumberStr)
	if err != nil {
		http.Error(w, "Invalid level number", http.StatusBadRequest)
		return
	}

	levelData, err := database.Get("levels", map[string]interface{}{"level_number": levelNumber})
	if err != nil {
		LevelNotFoundHandler(w, r)
		return
	}
	level, ok := levelData.(*database.AdminLevel)
	if !ok {
		http.Error(w, "Error parsing level data", http.StatusInternalServerError)
		return
	}

	errorMsg := r.URL.Query().Get("error")

	data := AdminPageData{
		User:         user,
		Levels:       []database.AdminLevel{*level},
		IsAdmin:      true,
		ErrorMessage: errorMsg,
	}

	renderAdminTemplate(w, "admin_edit_level.html", data)
}

func renderAdminTemplate(w http.ResponseWriter, templateName string, data AdminPageData) {
	templatePath := filepath.Join("frontend", templateName)

	tmpl := template.New(templateName).Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"slice": func(s string, start, end int) string {
			if start < 0 {
				start = 0
			}
			if end > len(s) {
				end = len(s)
			}
			if start >= end {
				return ""
			}
			return s[start:end]
		},
		"gt": func(a, b int) bool {
			return a > b
		},
		"len": func(s string) int {
			return len(s)
		},
		"isAdminEmail": isAdminEmail,
	})

	tmpl, err := tmpl.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func SubmitAnswerFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := GetUserFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	answer := r.FormValue("answer")
	if answer == "" {
		http.Redirect(w, r, "/?error=empty_answer", http.StatusSeeOther)
		return
	}

	answer = strings.TrimSpace(answer)

	loginData, err := database.Get("login", map[string]interface{}{"gmail": user.Gmail})
	if err != nil {
		http.Redirect(w, r, "/?error=level_error", http.StatusSeeOther)
		return
	}
	login, ok := loginData.(*database.Login)
	if !ok {
		http.Redirect(w, r, "/?error=level_error", http.StatusSeeOther)
		return
	}
	currentLevelNum := login.On

	levelData, err := database.Get("levels", map[string]interface{}{"level_number": currentLevelNum})
	if err != nil {
		http.Redirect(w, r, "/?error=check_error", http.StatusSeeOther)
		return
	}
	adminLevel, ok := levelData.(*database.AdminLevel)
	if !ok {
		http.Redirect(w, r, "/?error=check_error", http.StatusSeeOther)
		return
	}

	correct := (answer == adminLevel.Answer)

	if correct {
		newLevel := currentLevelNum + 1
		err = database.Update("login_field",
			map[string]interface{}{"gmail": user.Gmail, "field": "on"},
			map[string]interface{}{"value": newLevel})
		if err != nil {
			http.Redirect(w, r, "/?error=update_error", http.StatusSeeOther)
			return
		}

		err = database.Update("login_field",
			map[string]interface{}{"gmail": user.Gmail, "field": "score"},
			map[string]interface{}{"value": newLevel - 1})
		if err != nil {
			http.Redirect(w, r, "/?error=score_error", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/?success=correct", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/?error=incorrect", http.StatusSeeOther)
	}
}

func AuthPageHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := GetUserFromSession(r); err == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := PageData{}
	renderTemplate(w, "auth.html", data)
}

func GetSecretHandler(w http.ResponseWriter, r *http.Request) {
	secret := config.GetXSecretValue()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"secret": secret})
}

// Announcement handlers
func GetAllAnnouncementsHandler(w http.ResponseWriter, r *http.Request) {
	announcements, err := database.GetAllAnnouncements()
	if err != nil {
		http.Error(w, "Failed to get announcements", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(announcements)
}

func CreateAnnouncementHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Heading string `json:"heading"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Heading == "" {
		http.Error(w, "Heading is required", http.StatusBadRequest)
		return
	}

	if err := database.CreateAnnouncement(req.Heading); err != nil {
		http.Error(w, "Failed to create announcement", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Announcement created successfully"})
}

func UpdateAnnouncementHandler(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid announcement ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Heading string `json:"heading"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Heading == "" {
		http.Error(w, "Heading is required", http.StatusBadRequest)
		return
	}

	if err := database.UpdateAnnouncement(id, req.Heading); err != nil {
		http.Error(w, "Failed to update announcement", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Announcement updated successfully"})
}

func DeleteAnnouncementHandler(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid announcement ID", http.StatusBadRequest)
		return
	}

	if err := database.DeleteAnnouncement(id); err != nil {
		http.Error(w, "Failed to delete announcement", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Announcement deleted successfully"})
}

// Public handler for getting announcements (no auth required)
func GetAnnouncementsForPublicHandler(w http.ResponseWriter, r *http.Request) {
	announcements, err := database.GetAllAnnouncements()
	if err != nil {
		http.Error(w, "Failed to get announcements", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(announcements)
}
