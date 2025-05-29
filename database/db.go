package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Database structures - exported for use by other packages
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
	LoginCode          string // Permanent 4-digit login code
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

var db *sql.DB

// Initialize database
func InitDB() {
	var err error
	db, err = sql.Open("sqlite3", "./data/logins.db")
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
	}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			log.Fatal(err)
		}
	}
}

// GET - Retrieve data from database
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
		if params["all"] == true {
			rows, err := db.Query("SELECT gmail, hashed, seshTok, CSRFtok, verified, verificationNumber, \"on\" FROM logins")
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			var logins []Login
			for rows.Next() {
				var l Login
				if err := rows.Scan(&l.Gmail, &l.Hashed, &l.SeshTok, &l.CSRFtok, &l.Verified, &l.VerificationNumber, &l.On); err != nil {
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
		// Reverse to get chronological order
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
	}
	return nil, fmt.Errorf("invalid get request")
}

// CREATE - Insert new data into database
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
	}
	return fmt.Errorf("invalid create request")
}

// UPDATE - Update existing data in database
func Update(entity string, params map[string]interface{}, data interface{}) error {
	switch entity {
	case "login_field":
		if gmail, ok := params["gmail"].(string); ok {
			if field, ok := params["field"].(string); ok {
				query := fmt.Sprintf("UPDATE logins SET %s = ? WHERE gmail = ?", field)
				_, err := db.Exec(query, data, gmail)
				return err
			}
		}
	case "level":
		if number, ok := params["number"].(int); ok {
			if level, ok := data.(AdminLevel); ok {
				_, err := db.Exec(`UPDATE levels SET markdown = ?, src_hint = ?, console_hint = ?, answer = ?, active = ? WHERE level_number = ?`,
					level.Markdown, level.SourceHint, level.ConsoleHint, level.Answer, level.Active, number)
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
	}
	return fmt.Errorf("invalid update request")
}

// DELETE - Remove data from database
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
