package handlers

import (
	"encoding/json"
	"html/template"
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
	ActiveUsers int
}

type Config struct {
	Email struct {
		From     string `json:"from"`
		Password string `json:"password"`
		SMTPHost string `json:"smtp_host"`
		SMTPPort string `json:"smtp_port"`
	} `json:"email"`
	AdminEmails []string `json:"admin_emails"`
}

var AdminEmails = []string{"dsiddhant460@gmail.com", "lead@intrasudo.com", "organizer@intrasudo.com"}

func loadConfig() (*Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func isAdminEmail(email string) bool {
	email = strings.ToLower(email)
	for _, adminEmail := range AdminEmails {
		if email == strings.ToLower(adminEmail) {
			return true
		}
	}
	return false
}

func renderTemplate(w http.ResponseWriter, templateName string, data PageData) {
	templatePath := filepath.Join("frontend", templateName)

	// Create template with custom functions
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
	user, err := GetUserFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	if isAdminEmail(user.Gmail) {
		// Admins can always access the page, even if they have no level/data
		// Try to get a level, but don't block if missing
		var level *database.Level
		loginData, err := database.Get("login", map[string]interface{}{"gmail": user.Gmail})
		if err == nil {
			if login, ok := loginData.(*database.Login); ok {
				levelData, err := database.Get("levels", map[string]interface{}{"level_number": login.On})
				if err == nil {
					if l, ok := levelData.(*database.Level); ok {
						level = l
					}
				}
			}
		}
		// Get error/success messages from URL parameters
		errorMsg := r.URL.Query().Get("error")
		successMsg := r.URL.Query().Get("success")
		data := PageData{
			User:           user,
			Level:          level,
			IsAdmin:        true,
			ErrorMessage:   errorMsg,
			SuccessMessage: successMsg,
		}
		renderTemplate(w, "index.html", data)
		return
	}

	// Get current user level
	loginData, err := database.Get("login", map[string]interface{}{"gmail": user.Gmail})
	if err != nil {
		http.Error(w, "Error getting user data", http.StatusInternalServerError)
		return
	}
	login, ok := loginData.(*database.Login)
	if !ok {
		http.Error(w, "Error parsing user data", http.StatusInternalServerError)
		return
	}
	currentLevelNum := login.On

	// Get level data
	levelData, err := database.Get("levels", map[string]interface{}{"level_number": currentLevelNum})
	if err != nil {
		// If level doesn't exist, redirect to 404 page with level_not_found error
		LevelNotFoundHandler(w, r)
		return
	}
	level, ok := levelData.(*database.Level)
	if !ok {
		http.Error(w, "Error parsing level data", http.StatusInternalServerError)
		return
	}

	// Get error/success messages from URL parameters
	errorMsg := r.URL.Query().Get("error")
	successMsg := r.URL.Query().Get("success")

	data := PageData{
		User:           user,
		Level:          level,
		IsAdmin:        isAdminEmail(user.Gmail),
		ErrorMessage:   errorMsg,
		SuccessMessage: successMsg,
	}

	renderTemplate(w, "index.html", data)
}

func HintsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./frontend/hints.html")
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

	// Get admin levels
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

	// Load admin data
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

	usersData, err := database.Get("all_logins", map[string]interface{}{})
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
		ActiveUsers: countActiveUsers(users),
	}

	// Get error/success messages from query parameters
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

	// Get error messages from query parameters
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

	// Get error messages from query parameters
	errorMsg := r.URL.Query().Get("error")

	data := AdminPageData{
		User:         user,
		Levels:       []database.AdminLevel{*level},
		IsAdmin:      true,
		ErrorMessage: errorMsg,
	}

	renderAdminTemplate(w, "admin_edit_level.html", data)
}

func countActiveUsers(users []database.Login) int {
	// For now, count all users as active
	// This could be enhanced to check last login time
	return len(users)
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

func SubmitAnswerHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get current user level
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

	// Check answer using AdminLevel to get the answer field
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
		// Update user's current level
		newLevel := currentLevelNum + 1
		err = database.Update("login_field",
			map[string]interface{}{"gmail": user.Gmail, "field": "on"},
			map[string]interface{}{"value": newLevel})
		if err != nil {
			http.Redirect(w, r, "/?error=update_error", http.StatusSeeOther)
			return
		}

		// Update score (could be level number or points system)
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
	// If user is already logged in, redirect to main page
	if _, err := GetUserFromSession(r); err == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := PageData{}
	renderTemplate(w, "auth.html", data)
}

func ChatPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./frontend/chat.html")
}

func init() {
	config, err := loadConfig()
	if err == nil {
		AdminEmails = config.AdminEmails
	} else {
		AdminEmails = []string{"admin@intrasudo.com", "lead@intrasudo.com", "organizer@intrasudo.com"}
	}
}
