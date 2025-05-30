package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"intrasudo25/config"
	_ "github.com/mattn/go-sqlite3"
)

type Level struct {
	LevelNumber int    `json:"levelNumber"`
	Markdown    string `json:"markdown"`
	SourceHint  string `json:"sourceHint"`
	ConsoleHint string `json:"consoleHint"`
}

type AdminLevel struct {
	LevelNumber int    `json:"levelNumber"`
	Markdown    string `json:"markdown"`
	SourceHint  string `json:"sourceHint"`
	ConsoleHint string `json:"consoleHint"`
	Answer      string `json:"answer"`
	Active      bool   `json:"active"`
}

type Login struct {
	Hashed             string
	SeshTok            string
	CSRFtok            string
	Gmail              string
	Name               string
	Verified           bool
	VerificationNumber string
	LoginCode          string
	On                 uint
}

type Sucker struct {
	Gmail string
	Score int
}

type ChatMessage struct {
	ID            int    `json:"id"`
	UserEmail     string `json:"userEmail"`
	Username      string `json:"username"`
	Message       string `json:"message"`
	Timestamp     string `json:"timestamp"`
	FormattedTime string `json:"formattedTime"`
	IsAdmin       bool   `json:"isAdmin"`
}

type ChatParticipant struct {
	Email    string `json:"email"`
	IsOnline bool   `json:"isOnline"`
	LastSeen string `json:"lastSeen"`
	IsAdmin  bool   `json:"isAdmin"`
}

// AdminStats represents admin dashboard statistics
type AdminStats struct {
	TotalUsers  int `json:"totalUsers"`
	TotalLevels int `json:"totalLevels"`
	ActiveUsers int `json:"activeUsers"`
}

// GetAdminStats returns comprehensive admin statistics
func GetAdminStats() (*AdminStats, error) {
	stats := &AdminStats{}

	// Get admin emails from config
	adminEmails := config.GetAdminEmails()

	// Build WHERE clause to exclude admin emails
	whereClause := ""
	args := []interface{}{}
	if len(adminEmails) > 0 {
		placeholders := make([]string, len(adminEmails))
		for i, email := range adminEmails {
			placeholders[i] = "?"
			args = append(args, email)
		}
		whereClause = " WHERE gmail NOT IN (" + strings.Join(placeholders, ",") + ")"
	}

	// Get total non-admin users
	totalQuery := "SELECT COUNT(*) FROM logins" + whereClause
	err := db.QueryRow(totalQuery, args...).Scan(&stats.TotalUsers)
	if err != nil {
		return nil, err
	}

	// Get total levels
	err = db.QueryRow("SELECT COUNT(*) FROM levels").Scan(&stats.TotalLevels)
	if err != nil {
		return nil, err
	}

	// Get active non-admin users (verified users excluding admins)
	activeQuery := "SELECT COUNT(*) FROM logins" + whereClause
	if whereClause != "" {
		activeQuery += " AND verified = 1"
	} else {
		activeQuery += " WHERE verified = 1"
	}
	err = db.QueryRow(activeQuery, args...).Scan(&stats.ActiveUsers)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// AdminLevelResponse represents a level with admin fields for frontend
type AdminLevelResponse struct {
	ID       int    `json:"id"`
	Number   int    `json:"number"`
	Title    string `json:"title"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Active   bool   `json:"active"`
	Enabled  bool   `json:"enabled"`
}

// GetAllLevelsForAdmin returns all levels with admin-specific data
func GetAllLevelsForAdmin() ([]AdminLevelResponse, error) {
	rows, err := db.Query("SELECT level_number, markdown, src_hint, console_hint, answer, active FROM levels ORDER BY level_number")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var levels []AdminLevelResponse
	for rows.Next() {
		var level AdminLevelResponse
		var srcHint, consoleHint string

		err := rows.Scan(&level.Number, &level.Question, &srcHint, &consoleHint, &level.Answer, &level.Active)
		if err != nil {
			continue
		}

		level.ID = level.Number // Use number as ID for frontend compatibility
		level.Title = fmt.Sprintf("Level %d", level.Number)
		level.Enabled = level.Active // Set enabled same as active for frontend

		levels = append(levels, level)
	}

	return levels, nil
}

// AdminUserResponse represents a user with admin-specific data
type AdminUserResponse struct {
	Gmail    string `json:"Gmail"`
	Name     string `json:"Name"`
	On       uint   `json:"On"`
	Verified bool   `json:"Verified"`
	IsAdmin  bool   `json:"IsAdmin"`
}

// GetAllUsersForAdmin returns all users with admin-specific data
func GetAllUsersForAdmin() ([]AdminUserResponse, error) {
	rows, err := db.Query("SELECT gmail, name, \"on\", verified FROM logins ORDER BY gmail")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []AdminUserResponse
	adminEmails := config.GetAdminEmails() // Use config instead of hardcoded emails

	for rows.Next() {
		var user AdminUserResponse
		var name sql.NullString

		err := rows.Scan(&user.Gmail, &name, &user.On, &user.Verified)
		if err != nil {
			continue
		}

		if name.Valid {
			user.Name = name.String
		} else {
			user.Name = user.Gmail
		}

		// Check if user is admin
		for _, adminEmail := range adminEmails {
			if strings.EqualFold(user.Gmail, adminEmail) {
				user.IsAdmin = true
				break
			}
		}

		users = append(users, user)
	}

	return users, nil
}

// CreateLevelSimple creates a level with minimal validation (backend handles logic)
func CreateLevelSimple(levelNum int, question, answer string, active bool) error {
	level := AdminLevel{
		LevelNumber: levelNum,
		Markdown:    question,
		SourceHint:  question,
		ConsoleHint: question,
		Answer:      answer,
		Active:      active,
	}
	return Create("level", level)
}

// UpdateLevelSimple updates a level with minimal validation (backend handles logic)
func UpdateLevelSimple(levelNum int, question, answer string, active bool) error {
	level := AdminLevel{
		LevelNumber: levelNum,
		Markdown:    question,
		SourceHint:  question,
		ConsoleHint: question,
		Answer:      answer,
		Active:      active,
	}
	return Update("level", map[string]interface{}{"number": levelNum}, level)
}

// DeleteLevelSimple deletes a level by number
func DeleteLevelSimple(levelNum int) error {
	return Delete("level", map[string]interface{}{"number": levelNum})
}

// DeleteUserSimple deletes a user by email
func DeleteUserSimple(email string) error {
	return Delete("login", map[string]interface{}{"gmail": email})
}

// ToggleLevelState toggles the active state of a specific level
func ToggleLevelState(levelNum int, enabled bool) error {
	return Update("level_state", map[string]interface{}{"number": levelNum}, enabled)
}

// ToggleAllLevelsState toggles the active state of all levels
func ToggleAllLevelsState(enabled bool) error {
	return Update("bulk_level_state", map[string]interface{}{}, enabled)
}

var db *sql.DB

func InitDB() {
	var err error
	db, err = sql.Open("sqlite3", "./data/data.db")
	if err != nil {
		log.Fatal(err)
	}
	createTables()
}

func createTables() {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS logins (
			gmail TEXT PRIMARY KEY,
			hashed TEXT NOT NULL,
			seshTok TEXT,
			CSRFtok TEXT,
			name TEXT,
			verified BOOLEAN,
			verificationNumber TEXT,
			loginCode TEXT,
			"on" INTEGER DEFAULT 1
		);`,
		`CREATE TABLE IF NOT EXISTS leaderboard (
			gmail TEXT PRIMARY KEY,
			score INTEGER
		);`,
		`CREATE TABLE IF NOT EXISTS levels (
			level_number INTEGER PRIMARY KEY,
			markdown TEXT,
			src_hint TEXT,
			console_hint TEXT,
			answer TEXT NOT NULL,
			active BOOLEAN DEFAULT TRUE
		);`,
		`CREATE TABLE IF NOT EXISTS chat_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_email TEXT NOT NULL,
			message TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			is_admin BOOLEAN DEFAULT FALSE
		);`,
		`CREATE TABLE IF NOT EXISTS chat_participants (
			email TEXT PRIMARY KEY,
			is_online BOOLEAN DEFAULT FALSE,
			last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
			is_admin BOOLEAN DEFAULT FALSE
		);`,
		`CREATE TABLE IF NOT EXISTS notifications (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_email TEXT NOT NULL,
			message TEXT NOT NULL,
			type TEXT DEFAULT 'info',
			read BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			log.Fatal(err)
		}
	}
}

func Get(entity string, params map[string]interface{}) (interface{}, error) {
	switch entity {
	case "login":
		if gmail, ok := params["gmail"].(string); ok {
			var l Login
			err := db.QueryRow("SELECT gmail, hashed, seshTok, CSRFtok, name, verified, verificationNumber, loginCode, \"on\" FROM logins WHERE gmail = ?", gmail).
				Scan(&l.Gmail, &l.Hashed, &l.SeshTok, &l.CSRFtok, &l.Name, &l.Verified, &l.VerificationNumber, &l.LoginCode, &l.On)
			if err != nil {
				return nil, err
			}
			return &l, nil
		}
		if cookie, ok := params["cookie"].(string); ok {
			var l Login
			err := db.QueryRow("SELECT gmail, hashed, seshTok, CSRFtok, name, verified, verificationNumber, loginCode, \"on\" FROM logins WHERE seshTok = ?", cookie).
				Scan(&l.Gmail, &l.Hashed, &l.SeshTok, &l.CSRFtok, &l.Name, &l.Verified, &l.VerificationNumber, &l.LoginCode, &l.On)
			if err != nil {
				return nil, err
			}
			return &l, nil
		}
		if seshTok, ok := params["seshTok"].(string); ok {
			var l Login
			err := db.QueryRow("SELECT gmail, hashed, seshTok, CSRFtok, name, verified, verificationNumber, loginCode, \"on\" FROM logins WHERE seshTok = ?", seshTok).
				Scan(&l.Gmail, &l.Hashed, &l.SeshTok, &l.CSRFtok, &l.Name, &l.Verified, &l.VerificationNumber, &l.LoginCode, &l.On)
			if err != nil {
				return nil, err
			}
			return &l, nil
		}
		if params["all"] == true {
			rows, err := db.Query("SELECT gmail, hashed, seshTok, CSRFtok, name, verified, verificationNumber, \"on\" FROM logins")
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			var logins []Login
			for rows.Next() {
				var l Login
				if err := rows.Scan(&l.Gmail, &l.Hashed, &l.SeshTok, &l.CSRFtok, &l.Name, &l.Verified, &l.VerificationNumber, &l.On); err != nil {
					return nil, err
				}
				logins = append(logins, l)
			}
			return logins, nil
		}
	case "level":
		if number, ok := params["number"].(int); ok {
			if params["admin"] == true {
				var l AdminLevel
				err := db.QueryRow("SELECT level_number, markdown, src_hint, console_hint, answer, active FROM levels WHERE level_number = ?", number).
					Scan(&l.LevelNumber, &l.Markdown, &l.SourceHint, &l.ConsoleHint, &l.Answer, &l.Active)
				if err != nil {
					return nil, err
				}
				return &l, nil
			}
			var l Level
			err := db.QueryRow("SELECT level_number, markdown, src_hint, console_hint FROM levels WHERE level_number = ? AND active = 1", number).
				Scan(&l.LevelNumber, &l.Markdown, &l.SourceHint, &l.ConsoleHint)
			if err != nil {
				return nil, err
			}
			return &l, nil
		}
		if params["all"] == true {
			rows, err := db.Query("SELECT level_number, markdown, src_hint, console_hint, answer, active FROM levels ORDER BY level_number")
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			var levels []AdminLevel
			for rows.Next() {
				var l AdminLevel
				if err := rows.Scan(&l.LevelNumber, &l.Markdown, &l.SourceHint, &l.ConsoleHint, &l.Answer, &l.Active); err != nil {
					return nil, err
				}
				levels = append(levels, l)
			}
			return levels, nil
		}
	case "leaderboard":
		limit := 0
		if l, ok := params["limit"].(int); ok {
			limit = l
		}
		query := "SELECT gmail, score FROM leaderboard ORDER BY score DESC"
		if limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", limit)
		}
		rows, err := db.Query(query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var suckers []Sucker
		for rows.Next() {
			var s Sucker
			if err := rows.Scan(&s.Gmail, &s.Score); err != nil {
				return nil, err
			}
			suckers = append(suckers, s)
		}
		return suckers, nil
	case "chat_messages":
		limit := 50
		if l, ok := params["limit"].(int); ok {
			limit = l
		}
		query := `SELECT cm.id, cm.user_email, cm.user_email, cm.message, 
			cm.timestamp, cm.timestamp, cm.is_admin 
			FROM chat_messages cm ORDER BY cm.timestamp DESC LIMIT ?`
		rows, err := db.Query(query, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var messages []ChatMessage
		for rows.Next() {
			var msg ChatMessage
			if err := rows.Scan(&msg.ID, &msg.UserEmail, &msg.Username, &msg.Message,
				&msg.Timestamp, &msg.FormattedTime, &msg.IsAdmin); err != nil {
				return nil, err
			}
			messages = append(messages, msg)
		}
		for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
			messages[i], messages[j] = messages[j], messages[i]
		}
		return messages, nil
	case "chat_participants":
		rows, err := db.Query("SELECT email, is_online, last_seen, is_admin FROM chat_participants WHERE is_online = 1")
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var participants []ChatParticipant
		for rows.Next() {
			var p ChatParticipant
			if err := rows.Scan(&p.Email, &p.IsOnline, &p.LastSeen, &p.IsAdmin); err != nil {
				return nil, err
			}
			participants = append(participants, p)
		}
		return participants, nil
	case "user_score":
		if gmail, ok := params["gmail"].(string); ok {
			var score int
			err := db.QueryRow("SELECT score FROM leaderboard WHERE gmail = ?", gmail).Scan(&score)
			if err != nil {
				return 0, err
			}
			return score, nil
		}
	case "current_level":
		if gmail, ok := params["gmail"].(string); ok {
			var level int
			err := db.QueryRow("SELECT \"on\" FROM logins WHERE gmail = ?", gmail).Scan(&level)
			if err != nil {
				return 1, err
			}
			return level, nil
		}
	case "check_answer":
		if level, ok := params["level"].(int); ok {
			if answer, ok := params["answer"].(string); ok {
				var correctAnswer string
				err := db.QueryRow("SELECT answer FROM levels WHERE level_number = ?", level).Scan(&correctAnswer)
				if err != nil {
					return false, err
				}
				return strings.TrimSpace(strings.ToLower(answer)) == strings.TrimSpace(strings.ToLower(correctAnswer)), nil
			}
		}
	case "notification_count":
		if gmail, ok := params["gmail"].(string); ok {
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM notifications WHERE user_email = ? AND read = 0", gmail).Scan(&count)
			if err != nil {
				return 0, err
			}
			return count, nil
		}
	case "level_hints":
		if number, ok := params["number"].(int); ok {
			var l Level
			err := db.QueryRow("SELECT level_number, src_hint, console_hint FROM levels WHERE level_number = ? AND active = 1", number).
				Scan(&l.LevelNumber, &l.SourceHint, &l.ConsoleHint)
			if err != nil {
				return nil, err
			}
			return &l, nil
		}
	case "user_current_level_data":
		if gmail, ok := params["gmail"].(string); ok {
			var userLevel int
			err := db.QueryRow("SELECT \"on\" FROM logins WHERE gmail = ?", gmail).Scan(&userLevel)
			if err != nil {
				return nil, err
			}
			var l Level
			err = db.QueryRow("SELECT level_number, markdown, src_hint, console_hint FROM levels WHERE level_number = ? AND active = 1", userLevel).
				Scan(&l.LevelNumber, &l.Markdown, &l.SourceHint, &l.ConsoleHint)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"id":          l.LevelNumber,
				"number":      l.LevelNumber,
				"description": l.Markdown,
				"sourceHint":  l.SourceHint,
				"consoleHint": l.ConsoleHint,
			}, nil
		}
	}
	return nil, fmt.Errorf("invalid get request")
}

func Create(entity string, data interface{}) error {
	switch entity {
	case "login":
		if login, ok := data.(Login); ok {
			_, err := db.Exec("INSERT INTO logins (gmail, hashed, seshTok, CSRFtok, name, verified, verificationNumber, loginCode, \"on\") VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
				login.Gmail, login.Hashed, login.SeshTok, login.CSRFtok, login.Name, login.Verified, login.VerificationNumber, login.LoginCode, login.On)
			return err
		}
	case "level":
		if level, ok := data.(AdminLevel); ok {
			_, err := db.Exec("INSERT INTO levels (level_number, markdown, src_hint, console_hint, answer, active) VALUES (?, ?, ?, ?, ?, ?)",
				level.LevelNumber, level.Markdown, level.SourceHint, level.ConsoleHint, level.Answer, level.Active)
			return err
		}
	case "leaderboard":
		if sucker, ok := data.(Sucker); ok {
			_, err := db.Exec("INSERT INTO leaderboard (gmail, score) VALUES (?, ?)", sucker.Gmail, sucker.Score)
			return err
		}
	case "chat_message":
		if params, ok := data.(map[string]interface{}); ok {
			email := params["email"].(string)
			message := params["message"].(string)
			isAdmin := params["isAdmin"].(bool)
			_, err := db.Exec("INSERT INTO chat_messages (user_email, message, is_admin) VALUES (?, ?, ?)", email, message, isAdmin)
			return err
		}
	case "notification":
		if params, ok := data.(map[string]interface{}); ok {
			userEmail := params["userEmail"].(string)
			message := params["message"].(string)
			notifType := params["type"].(string)
			_, err := db.Exec("INSERT INTO notifications (user_email, message, type, read) VALUES (?, ?, ?, 0)", userEmail, message, notifType)
			return err
		}
	}
	return fmt.Errorf("invalid create request")
}

func Update(entity string, params map[string]interface{}, data interface{}) error {
	switch entity {
	case "login_field":
		if gmail, ok := params["gmail"].(string); ok {
			if field, ok := params["field"].(string); ok {
				if field == "on" {
					field = `"on"`
				}
				query := fmt.Sprintf("UPDATE logins SET %s = ? WHERE gmail = ?", field)
				_, err := db.Exec(query, data, gmail)
				return err
			}
		}
	case "login":
		if seshTok, ok := params["seshTok"].(string); ok {
			if updateData, ok := data.(map[string]interface{}); ok {
				_, err := db.Exec("UPDATE logins SET seshTok = ?, CSRFtok = ? WHERE seshTok = ?",
					updateData["seshTok"], updateData["CSRFtok"], seshTok)
				return err
			}
		}
	case "level":
		if number, ok := params["number"].(int); ok {
			if level, ok := data.(AdminLevel); ok {
				_, err := db.Exec(`UPDATE levels SET level_number = ?, markdown = ?, src_hint = ?, console_hint = ?, answer = ?, active = ? WHERE level_number = ?`,
					level.LevelNumber, level.Markdown, level.SourceHint, level.ConsoleHint, level.Answer, level.Active, number)
				return err
			}
		}
	case "score":
		if gmail, ok := params["gmail"].(string); ok {
			if score, ok := data.(int); ok {
				_, err := db.Exec(`INSERT INTO leaderboard (gmail, score) VALUES (?, ?) ON CONFLICT(gmail) DO UPDATE SET score = excluded.score`, gmail, score)
				return err
			}
		}
	case "chat_participant":
		if params, ok := data.(map[string]interface{}); ok {
			email := params["email"].(string)
			isOnline := params["isOnline"].(bool)
			isAdmin := params["isAdmin"].(bool)
			_, err := db.Exec(`INSERT INTO chat_participants (email, is_online, is_admin) VALUES (?, ?, ?) 
				ON CONFLICT(email) DO UPDATE SET is_online = excluded.is_online, last_seen = CURRENT_TIMESTAMP, is_admin = excluded.is_admin`,
				email, isOnline, isAdmin)
			return err
		}
	case "level_state":
		if number, ok := params["number"].(int); ok {
			if state, ok := data.(bool); ok {
				_, err := db.Exec("UPDATE levels SET active = ? WHERE level_number = ?", state, number)
				return err
			}
		}
	case "bulk_level_state":
		if state, ok := data.(bool); ok {
			_, err := db.Exec("UPDATE levels SET active = ?", state)
			return err
		}
	case "notification_read":
		if id, ok := params["id"].(int); ok {
			_, err := db.Exec("UPDATE notifications SET read = 1 WHERE id = ?", id)
			return err
		}
		if gmail, ok := params["gmail"].(string); ok {
			_, err := db.Exec("UPDATE notifications SET read = 1 WHERE user_email = ?", gmail)
			return err
		}
	}
	return fmt.Errorf("invalid update request")
}

func Delete(entity string, params map[string]interface{}) error {
	switch entity {
	case "login":
		if gmail, ok := params["gmail"].(string); ok {
			_, err := db.Exec("DELETE FROM logins WHERE gmail = ?", gmail)
			return err
		}
	case "level":
		if number, ok := params["number"].(int); ok {
			_, err := db.Exec("DELETE FROM levels WHERE level_number = ?", number)
			return err
		}
	}
	return fmt.Errorf("invalid delete request")
}

// GameLevel represents a level for the game interface
type GameLevel struct {
	ID          int    `json:"id"`
	Number      int    `json:"number"`
	Description string `json:"description"`
	MediaURL    string `json:"mediaUrl,omitempty"`
	MediaType   string `json:"mediaType,omitempty"`
}

// GetCurrentLevelForUser returns the current level data for a user
func GetCurrentLevelForUser(userEmail string) (*GameLevel, error) {
	// First, check if level 1 exists and is active - this is required for the game to work
	var level1Exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM levels WHERE level_number = 1 AND active = 1)").Scan(&level1Exists)
	if err != nil {
		log.Printf("ERROR: Failed to check if level 1 exists: %v", err)
		return nil, fmt.Errorf("database error checking level 1")
	}

	if !level1Exists {
		log.Printf("ERROR: Level 1 does not exist or is not active")
		return nil, fmt.Errorf("level 1 must exist and be active for the game to function")
	}

	// Get user's current level
	var user Login
	err = db.QueryRow("SELECT \"on\" FROM logins WHERE gmail = ?", userEmail).Scan(&user.On)
	if err != nil {
		log.Printf("ERROR: Failed to get user level for %s: %v", userEmail, err)
		return nil, fmt.Errorf("user not found or no level assigned")
	}

	log.Printf("DEBUG: User %s is on level %d", userEmail, user.On)

	// Ensure progressive access: user can only be on level N if they've completed levels 1 through N-1
	// But if user is on level 1, they should always be able to access it
	if user.On > 1 {
		// Check that all previous levels exist and are active
		var allPreviousExist bool
		err = db.QueryRow("SELECT COUNT(*) = ? FROM levels WHERE level_number BETWEEN 1 AND ? AND active = 1", user.On-1, user.On-1).Scan(&allPreviousExist)
		if err != nil {
			log.Printf("ERROR: Failed to check previous levels for user %s: %v", userEmail, err)
			return nil, fmt.Errorf("database error checking level progression")
		}

		if !allPreviousExist {
			log.Printf("ERROR: User %s is on level %d but not all previous levels (1-%d) exist or are active", userEmail, user.On, user.On-1)
			// Reset user to level 1 since progression is broken
			_, err = db.Exec("UPDATE logins SET \"on\" = 1 WHERE gmail = ?", userEmail)
			if err != nil {
				log.Printf("ERROR: Failed to reset user %s to level 1: %v", userEmail, err)
			}
			user.On = 1
			log.Printf("INFO: Reset user %s to level 1 due to broken progression", userEmail)
		}
	}

	// Now get the level data for the user's current level
	var level AdminLevel
	err = db.QueryRow("SELECT level_number, markdown FROM levels WHERE level_number = ? AND active = 1", user.On).Scan(&level.LevelNumber, &level.Markdown)
	if err != nil {
		log.Printf("ERROR: Failed to get level %d for user %s: %v", user.On, userEmail, err)
		// If the user's current level doesn't exist, reset them to level 1
		if user.On > 1 {
			_, resetErr := db.Exec("UPDATE logins SET \"on\" = 1 WHERE gmail = ?", userEmail)
			if resetErr != nil {
				log.Printf("ERROR: Failed to reset user %s to level 1: %v", userEmail, resetErr)
			} else {
				log.Printf("INFO: Reset user %s to level 1 because level %d doesn't exist", userEmail, user.On)
				// Try to get level 1
				err = db.QueryRow("SELECT level_number, markdown FROM levels WHERE level_number = 1 AND active = 1").Scan(&level.LevelNumber, &level.Markdown)
				if err != nil {
					log.Printf("ERROR: Failed to get level 1 after reset: %v", err)
					return nil, fmt.Errorf("level 1 not found after reset")
				}
			}
		} else {
			return nil, fmt.Errorf("level not found or not active")
		}
	}

	log.Printf("DEBUG: Found level %d for user %s: %s", level.LevelNumber, userEmail, level.Markdown)

	gameLevel := &GameLevel{
		ID:          level.LevelNumber,
		Number:      level.LevelNumber,
		Description: level.Markdown,
	}

	return gameLevel, nil
}

// SubmitAnswerResult represents the result of an answer submission
type SubmitAnswerResult struct {
	Correct bool   `json:"correct"`
	Message string `json:"message"`
}

// CheckAnswer validates a user's answer and updates their progress
func CheckAnswer(userEmail string, levelID int, answer string) (*SubmitAnswerResult, error) {
	// Get current user level
	var currentLevel uint
	err := db.QueryRow("SELECT \"on\" FROM logins WHERE gmail = ?", userEmail).Scan(&currentLevel)
	if err != nil {
		return nil, err
	}

	// Verify user is on the correct level
	if int(currentLevel) != levelID {
		return &SubmitAnswerResult{
			Correct: false,
			Message: "You're not on this level",
		}, nil
	}

	// Get correct answer
	var correctAnswer string
	err = db.QueryRow("SELECT answer FROM levels WHERE level_number = ? AND active = 1", levelID).Scan(&correctAnswer)
	if err != nil {
		return &SubmitAnswerResult{
			Correct: false,
			Message: "Level not found",
		}, nil
	}

	// Check if answer is correct (case-insensitive)
	if strings.EqualFold(strings.TrimSpace(answer), strings.TrimSpace(correctAnswer)) {
		// Update user progress
		_, err = db.Exec("UPDATE logins SET \"on\" = \"on\" + 1 WHERE gmail = ?", userEmail)
		if err != nil {
			return nil, err
		}

		// Create notification for level completion
		notification := map[string]interface{}{
			"user_email": userEmail,
			"message":    fmt.Sprintf("Congratulations! You completed Level %d", levelID),
			"type":       "success",
		}
		Create("notification", notification)

		return &SubmitAnswerResult{
			Correct: true,
			Message: "Correct! Moving to next level...",
		}, nil
	}

	return &SubmitAnswerResult{
		Correct: false,
		Message: "Incorrect answer. Try again!",
	}, nil
}

// GetUnreadNotificationCount returns the count of unread notifications for a user
func GetUnreadNotificationCount(userEmail string) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM notifications WHERE user_email = ? AND read = 0", userEmail).Scan(&count)
	return count, err
}
