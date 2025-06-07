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
	On    uint
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

type LeadMessage struct {
	ID           int    `json:"id"`
	UserEmail    string `json:"userEmail"`
	Username     string `json:"username"`
	Message      string `json:"message"`
	LevelNumber  int    `json:"levelNumber"`
	Timestamp    string `json:"timestamp"`
	DiscordMsgID string `json:"discordMsgId"`
	IsReply      bool   `json:"isReply"`
	ParentMsgID  int    `json:"parentMsgId"`
}

type HintMessage struct {
	ID           int    `json:"id"`
	Message      string `json:"message"`
	LevelNumber  int    `json:"levelNumber"`
	Timestamp    string `json:"timestamp"`
	DiscordMsgID string `json:"discordMsgId"`
	SentBy       string `json:"sentBy"`
}

type ChatChecksum struct {
	MessagesHash string `json:"messagesHash"`
	LeadsHash    string `json:"leadsHash"`
}

type AdminStats struct {
	TotalUsers  int `json:"totalUsers"`
	TotalLevels int `json:"totalLevels"`
}

func GetAdminStats() (*AdminStats, error) {
	stats := &AdminStats{}

	adminEmails := config.GetAdminEmails()

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

	totalQuery := "SELECT COUNT(*) FROM logins" + whereClause
	err := db.QueryRow(totalQuery, args...).Scan(&stats.TotalUsers)
	if err != nil {
		return nil, err
	}

	err = db.QueryRow("SELECT COUNT(*) FROM levels").Scan(&stats.TotalLevels)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

type AdminLevelResponse struct {
	ID       int    `json:"id"`
	Number   int    `json:"number"`
	Title    string `json:"title"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Active   bool   `json:"active"`
	Enabled  bool   `json:"enabled"`
}

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

		level.ID = level.Number
		level.Title = fmt.Sprintf("Level %d", level.Number)
		level.Enabled = level.Active

		levels = append(levels, level)
	}

	return levels, nil
}

type AdminUserResponse struct {
	Gmail    string `json:"Gmail"`
	Name     string `json:"Name"`
	On       uint   `json:"On"`
	Verified bool   `json:"Verified"`
	IsAdmin  bool   `json:"IsAdmin"`
}

func GetAllUsersForAdmin() ([]AdminUserResponse, error) {
	rows, err := db.Query("SELECT gmail, name, \"on\", verified FROM logins ORDER BY gmail")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []AdminUserResponse
	adminEmails := config.GetAdminEmails()

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

func DeleteLevelSimple(levelNum int) error {
	return Delete("level", map[string]interface{}{"number": levelNum})
}

func DeleteUserSimple(email string) error {
	return Delete("login", map[string]interface{}{"gmail": email})
}

func ToggleLevelState(levelNum int, enabled bool) error {
	return Update("level_state", map[string]interface{}{"number": levelNum}, enabled)
}

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
		`CREATE TABLE IF NOT EXISTS announcements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			heading TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			active BOOLEAN DEFAULT TRUE
		);`,
		`CREATE TABLE IF NOT EXISTS lead_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_email TEXT NOT NULL,
			username TEXT NOT NULL,
			message TEXT NOT NULL,
			level_number INTEGER NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			discord_msg_id TEXT,
			is_reply BOOLEAN DEFAULT FALSE,
			parent_msg_id INTEGER,
			is_deleted BOOLEAN DEFAULT FALSE,
			FOREIGN KEY (parent_msg_id) REFERENCES lead_messages(id)
		);`,
		`CREATE TABLE IF NOT EXISTS hint_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			message TEXT NOT NULL,
			level_number INTEGER NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			discord_msg_id TEXT,
			sent_by TEXT NOT NULL,
			is_deleted BOOLEAN DEFAULT FALSE
		);`,
		`CREATE TABLE IF NOT EXISTS message_mappings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			db_message_id INTEGER NOT NULL,
			discord_msg_id TEXT NOT NULL,
			user_email TEXT NOT NULL,
			level_number INTEGER NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS system_settings (
			"key" TEXT PRIMARY KEY,
			"value" TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS level_completions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_email TEXT NOT NULL,
			level_number INTEGER NOT NULL,
			completed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_email, level_number)
		);`,
	}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			log.Fatal(err)
		}
	}

	runMigrations()
}

func runMigrations() {
	rows, err := db.Query("PRAGMA table_info(lead_messages)")
	if err != nil {
		return
	}
	defer rows.Close()

	hasIsDeleted := false
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue interface{}

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			continue
		}

		if name == "is_deleted" {
			hasIsDeleted = true
			break
		}
	}

	if !hasIsDeleted {
		db.Exec("ALTER TABLE lead_messages ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE")
	}

	rows2, err := db.Query("PRAGMA table_info(hint_messages)")
	if err != nil {
		return
	}
	defer rows2.Close()

	hasIsDeletedHints := false
	for rows2.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue interface{}

		err := rows2.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			continue
		}

		if name == "is_deleted" {
			hasIsDeletedHints = true
			break
		}
	}

	if !hasIsDeletedHints {
		db.Exec("ALTER TABLE hint_messages ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE")
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

		adminEmails := config.GetAdminEmails()

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

		baseQuery := `SELECT l.gmail, 0 as score, l."on" FROM logins l
			LEFT JOIN level_completions lc ON l.gmail = lc.user_email AND lc.level_number = l."on" - 1`

		if whereClause != "" {
			baseQuery += " " + strings.Replace(whereClause, "gmail", "l.gmail", -1)
		}

		query := baseQuery + ` ORDER BY l."on" DESC, lc.completed_at ASC`
		if limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", limit)
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var suckers []Sucker
		for rows.Next() {
			var s Sucker
			if err := rows.Scan(&s.Gmail, &s.Score, &s.On); err != nil {
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
	case "max_level":
		var maxLevel int
		err := db.QueryRow("SELECT MAX(level_number) FROM levels WHERE active = 1").Scan(&maxLevel)
		if err != nil {
			return 0, err
		}
		return maxLevel, nil
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
	case "lead_messages":
		if userEmail, ok := params["userEmail"].(string); ok {
			if level, ok := params["level"].(int); ok {
				rows, err := db.Query("SELECT id, user_email, username, message, level_number, timestamp, discord_msg_id, is_reply, parent_msg_id FROM lead_messages WHERE user_email = ? AND level_number = ? ORDER BY timestamp ASC", userEmail, level)
				if err != nil {
					return nil, err
				}
				defer rows.Close()
				var messages []LeadMessage
				for rows.Next() {
					var msg LeadMessage
					var discordMsgID sql.NullString
					var parentMsgIDInt sql.NullInt64
					if err := rows.Scan(&msg.ID, &msg.UserEmail, &msg.Username, &msg.Message, &msg.LevelNumber, &msg.Timestamp, &discordMsgID, &msg.IsReply, &parentMsgIDInt); err != nil {
						return nil, err
					}
					if discordMsgID.Valid {
						msg.DiscordMsgID = discordMsgID.String
					}
					if parentMsgIDInt.Valid {
						msg.ParentMsgID = int(parentMsgIDInt.Int64)
					}
					messages = append(messages, msg)
				}
				return messages, nil
			}
		}
	case "lead_message_by_discord_id":
		if discordMsgId, ok := params["discordMsgId"].(string); ok {
			var msg LeadMessage
			var discordMsgID sql.NullString
			var parentMsgIDInt sql.NullInt64
			err := db.QueryRow("SELECT id, user_email, username, message, level_number, timestamp, discord_msg_id, is_reply, parent_msg_id FROM lead_messages WHERE discord_msg_id = ?", discordMsgId).
				Scan(&msg.ID, &msg.UserEmail, &msg.Username, &msg.Message, &msg.LevelNumber, &msg.Timestamp, &discordMsgID, &msg.IsReply, &parentMsgIDInt)
			if err != nil {
				return nil, err
			}
			if discordMsgID.Valid {
				msg.DiscordMsgID = discordMsgID.String
			}
			if parentMsgIDInt.Valid {
				msg.ParentMsgID = int(parentMsgIDInt.Int64)
			}
			return msg, nil
		}
	case "hint_messages":
		if level, ok := params["level"].(int); ok {
			rows, err := db.Query("SELECT id, message, level_number, timestamp, discord_msg_id, sent_by FROM hint_messages WHERE level_number = ? ORDER BY timestamp ASC", level)
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			var messages []HintMessage
			for rows.Next() {
				var msg HintMessage
				var discordMsgID sql.NullString
				if err := rows.Scan(&msg.ID, &msg.Message, &msg.LevelNumber, &msg.Timestamp, &discordMsgID, &msg.SentBy); err != nil {
					return nil, err
				}
				if discordMsgID.Valid {
					msg.DiscordMsgID = discordMsgID.String
				}
				messages = append(messages, msg)
			}
			return messages, nil
		}
	case "user_level":
		if email, ok := params["email"].(string); ok {
			var level int
			err := db.QueryRow("SELECT \"on\" FROM logins WHERE gmail = ?", email).Scan(&level)
			if err != nil {
				return 1, err
			}
			return level, nil
		}
	case "users_at_level":
		if level, ok := params["level"].(int); ok {
			rows, err := db.Query("SELECT gmail FROM logins WHERE \"on\" = ?", level)
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			var users []string
			for rows.Next() {
				var email string
				if err := rows.Scan(&email); err != nil {
					return nil, err
				}
				users = append(users, email)
			}
			return users, nil
		}
	case "message_by_discord_id":
		discordMsgID := params["discordMsgId"].(string)
		var leadMsg LeadMessage
		err := db.QueryRow("SELECT id, user_email, username, message, level_number, timestamp, discord_msg_id, is_reply, parent_msg_id FROM lead_messages WHERE discord_msg_id = ?",
			discordMsgID).Scan(&leadMsg.ID, &leadMsg.UserEmail, &leadMsg.Username, &leadMsg.Message,
			&leadMsg.LevelNumber, &leadMsg.Timestamp, &leadMsg.DiscordMsgID, &leadMsg.IsReply, &leadMsg.ParentMsgID)
		if err != nil {
			return nil, err
		}
		return leadMsg, nil

	case "message_mapping":
		discordMsgID := params["discordMsgId"].(string)
		var dbMessageId int
		var userEmail string
		var levelNumber int
		var timestamp string
		err := db.QueryRow("SELECT db_message_id, user_email, level_number, timestamp FROM message_mappings WHERE discord_msg_id = ?",
			discordMsgID).Scan(&dbMessageId, &userEmail, &levelNumber, &timestamp)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"dbMessageId": dbMessageId,
			"userEmail":   userEmail,
			"levelNumber": levelNumber,
			"timestamp":   timestamp,
		}, nil

	case "lead_messages_by_content":
		userEmail := params["userEmail"].(string)
		level := params["level"].(int)
		content := params["content"].(string)

		query := "SELECT id, user_email, username, message, level_number, timestamp, discord_msg_id, is_reply, parent_msg_id FROM lead_messages WHERE user_email = ? AND level_number = ?"
		args := []interface{}{userEmail, level}

		if content != "" {
			query += " AND message LIKE ?"
			args = append(args, content+"%")
		}

		query += " ORDER BY timestamp DESC LIMIT 5"

		rows, err := db.Query(query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var leadMessages []LeadMessage
		for rows.Next() {
			var msg LeadMessage
			if err := rows.Scan(&msg.ID, &msg.UserEmail, &msg.Username, &msg.Message,
				&msg.LevelNumber, &msg.Timestamp, &msg.DiscordMsgID, &msg.IsReply, &msg.ParentMsgID); err != nil {
				return nil, err
			}
			leadMessages = append(leadMessages, msg)
		}

		return leadMessages, nil
	case "system_setting":
		if key, ok := params["key"].(string); ok {
			var setting SystemSetting
			err := db.QueryRow("SELECT \"key\", \"value\" FROM system_settings WHERE \"key\" = ?", key).Scan(&setting.Key, &setting.Value)
			if err != nil {
				return nil, err
			}
			return &setting, nil
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
	case "announcement":
		if announcement, ok := data.(Announcement); ok {
			_, err := db.Exec("INSERT INTO announcements (heading, active) VALUES (?, ?)", announcement.Heading, announcement.Active)
			return err
		}
	case "lead_message":
		if leadMsg, ok := data.(LeadMessage); ok {
			_, err := db.Exec("INSERT INTO lead_messages (user_email, username, message, level_number, discord_msg_id, is_reply, parent_msg_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
				leadMsg.UserEmail, leadMsg.Username, leadMsg.Message, leadMsg.LevelNumber, leadMsg.DiscordMsgID, leadMsg.IsReply, leadMsg.ParentMsgID)
			return err
		}
	case "hint_message":
		if hintMsg, ok := data.(HintMessage); ok {
			_, err := db.Exec("INSERT INTO hint_messages (message, level_number, discord_msg_id, sent_by) VALUES (?, ?, ?, ?)",
				hintMsg.Message, hintMsg.LevelNumber, hintMsg.DiscordMsgID, hintMsg.SentBy)
			return err
		}
	case "message_mapping":
		if params, ok := data.(map[string]interface{}); ok {
			dbMessageID := params["dbMessageId"].(int)
			discordMsgID := params["discordMsgId"].(string)
			userEmail := params["userEmail"].(string)
			levelNumber := params["levelNumber"].(int)

			_, err := db.Exec("INSERT INTO message_mappings (db_message_id, discord_msg_id, user_email, level_number) VALUES (?, ?, ?, ?)",
				dbMessageID, discordMsgID, userEmail, levelNumber)
			if err != nil {
				return err
			}
			return nil
		}
	case "system_setting":
		if params, ok := data.(map[string]interface{}); ok {
			key := params["key"].(string)
			value := params["value"].(string)
			_, err := db.Exec("INSERT INTO system_settings (\"key\", \"value\") VALUES (?, ?)", key, value)
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
	case "announcement":
		if id, ok := params["id"].(int); ok {
			if heading, ok := data.(string); ok {
				_, err := db.Exec("UPDATE announcements SET heading = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", heading, id)
				return err
			}
		}
	case "lead_message":
		if id, ok := params["id"].(int); ok {
			if updateData, ok := data.(map[string]interface{}); ok {
				if discordMsgID, exists := updateData["discordMsgID"]; exists {
					_, err := db.Exec("UPDATE lead_messages SET discord_msg_id = ? WHERE id = ?", discordMsgID, id)
					return err
				}
			}
		}
	case "system_setting":
		if key, ok := params["key"].(string); ok {
			if updateData, ok := data.(map[string]interface{}); ok {
				value := updateData["value"].(string)
				_, err := db.Exec("UPDATE system_settings SET \"value\" = ? WHERE \"key\" = ?", value, key)
				return err
			}
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
	case "announcement":
		if id, ok := params["id"].(int); ok {
			_, err := db.Exec("UPDATE announcements SET active = FALSE WHERE id = ?", id)
			return err
		}
	}
	return fmt.Errorf("invalid delete request")
}

type GameLevel struct {
	ID           int    `json:"id"`
	Number       int    `json:"number"`
	Description  string `json:"description"`
	MediaURL     string `json:"mediaUrl,omitempty"`
	MediaType    string `json:"mediaType,omitempty"`
	AllCompleted bool   `json:"allCompleted,omitempty"`
	MaxLevel     int    `json:"maxLevel,omitempty"`
}

func GetCurrentLevelForUser(userEmail string) (*GameLevel, error) {
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

	var user Login
	err = db.QueryRow("SELECT \"on\" FROM logins WHERE gmail = ?", userEmail).Scan(&user.On)
	if err != nil {
		log.Printf("ERROR: Failed to get user level for %s: %v", userEmail, err)
		return nil, fmt.Errorf("user not found or no level assigned")
	}

	log.Printf("DEBUG: User %s is on level %d", userEmail, user.On)

	var maxLevelNumber int
	err = db.QueryRow("SELECT MAX(level_number) FROM levels WHERE active = 1").Scan(&maxLevelNumber)
	if err != nil {
		log.Printf("ERROR: Failed to get max level number: %v", err)
		return nil, fmt.Errorf("error determining maximum level")
	}

	if user.On > 1 {
		var allPreviousExist bool
		err = db.QueryRow("SELECT COUNT(*) = ? FROM levels WHERE level_number BETWEEN 1 AND ? AND active = 1", user.On-1, user.On-1).Scan(&allPreviousExist)
		if err != nil {
			log.Printf("ERROR: Failed to check previous levels for user %s: %v", userEmail, err)
			return nil, fmt.Errorf("database error checking level progression")
		}

		if !allPreviousExist {
			log.Printf("ERROR: User %s is on level %d but not all previous levels (1-%d) exist or are active", userEmail, user.On, user.On-1)
			_, err = db.Exec("UPDATE logins SET \"on\" = 1 WHERE gmail = ?", userEmail)
			if err != nil {
				log.Printf("ERROR: Failed to reset user %s to level 1: %v", userEmail, err)
			}
			user.On = 1
			log.Printf("INFO: Reset user %s to level 1 due to broken progression", userEmail)
		}
	}

	if int(user.On) > maxLevelNumber {
		log.Printf("INFO: User %s has completed all available levels (current level: %d, max level: %d)", userEmail, user.On, maxLevelNumber)
		gameLevel := &GameLevel{
			ID:           0,
			Number:       maxLevelNumber + 1,
			Description:  "You have completed all available levels!",
			AllCompleted: true,
			MaxLevel:     maxLevelNumber,
		}
		return gameLevel, nil
	}

	var level AdminLevel
	err = db.QueryRow("SELECT level_number, markdown FROM levels WHERE level_number = ? AND active = 1", user.On).Scan(&level.LevelNumber, &level.Markdown)
	if err != nil {
		log.Printf("ERROR: Failed to get level %d for user %s: %v", user.On, userEmail, err)
		if user.On > 1 {
			_, resetErr := db.Exec("UPDATE logins SET \"on\" = 1 WHERE gmail = ?", userEmail)
			if resetErr != nil {
				log.Printf("ERROR: Failed to reset user %s to level 1: %v", userEmail, resetErr)
			} else {
				log.Printf("INFO: Reset user %s to level 1 because level %d doesn't exist", userEmail, user.On)
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

type SubmitAnswerResult struct {
	Correct    bool   `json:"correct"`
	Message    string `json:"message"`
	ReloadPage bool   `json:"reload_page"`
}

func CheckAnswer(userEmail string, levelID int, answer string) (*SubmitAnswerResult, error) {
	if strings.Contains(answer, " ") {
		return &SubmitAnswerResult{
			Correct: false,
			Message: "Answer cannot contain spaces. Please enter a valid answer without spaces.",
		}, nil
	}

	for _, char := range answer {
		if char >= 'A' && char <= 'Z' {
			return &SubmitAnswerResult{
				Correct: false,
				Message: "Answer must be lowercase only. Please enter the answer in lowercase.",
			}, nil
		}
	}

	answer = strings.TrimSpace(answer)

	var currentLevel uint
	err := db.QueryRow("SELECT \"on\" FROM logins WHERE gmail = ?", userEmail).Scan(&currentLevel)
	if err != nil {
		return nil, err
	}

	log.Printf("DEBUG CheckAnswer: User %s, currentLevel=%d, submittedLevelID=%d", userEmail, currentLevel, levelID)

	if int(currentLevel) != levelID {
		log.Printf("DEBUG CheckAnswer: Level mismatch for user %s - user is on level %d but submitted answer for level %d", userEmail, currentLevel, levelID)
		return &SubmitAnswerResult{
			Correct:    true,
			Message:    "Validating...",
			ReloadPage: true,
		}, nil
	}

	if int(currentLevel) > levelID {
		log.Printf("DEBUG CheckAnswer: User %s already completed level %d (currently on %d), ignoring duplicate submission", userEmail, levelID, currentLevel)
		return &SubmitAnswerResult{
			Correct:    true,
			Message:    "Validating...",
			ReloadPage: true,
		}, nil
	}

	var correctAnswer string
	err = db.QueryRow("SELECT answer FROM levels WHERE level_number = ? AND active = 1", levelID).Scan(&correctAnswer)
	if err != nil {
		return &SubmitAnswerResult{
			Correct: false,
			Message: "Level not found",
		}, nil
	}

	if answer == strings.TrimSpace(correctAnswer) {
		var maxLevelNumber int
		err = db.QueryRow("SELECT MAX(level_number) FROM levels WHERE active = 1").Scan(&maxLevelNumber)
		if err != nil {
			log.Printf("ERROR: Failed to get max level number: %v", err)
			return nil, err
		}

		// Record level completion time
		_, err = db.Exec("INSERT OR IGNORE INTO level_completions (user_email, level_number) VALUES (?, ?)", userEmail, levelID)
		if err != nil {
			log.Printf("ERROR: Failed to record level completion time: %v", err)
		}

		if levelID == maxLevelNumber {
			_, err = db.Exec("UPDATE logins SET \"on\" = ? WHERE gmail = ?", maxLevelNumber+1, userEmail)
		} else {
			_, err = db.Exec("UPDATE logins SET \"on\" = \"on\" + 1 WHERE gmail = ?", userEmail)
		}

		if err != nil {
			return nil, err
		}

		log.Printf("DEBUG CheckAnswer: User %s answered correctly for level %d, promoted to level %d", userEmail, levelID, levelID+1)

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

type Announcement struct {
	ID        int    `json:"id" db:"id"`
	Heading   string `json:"heading" db:"heading"`
	CreatedAt string `json:"created_at" db:"created_at"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
	Active    bool   `json:"active" db:"active"`
}

func CreateAnnouncement(heading string) error {
	query := `INSERT INTO announcements (heading) VALUES (?)`
	_, err := db.Exec(query, heading)
	return err
}

func GetAllAnnouncements() ([]Announcement, error) {
	var announcements []Announcement
	query := `SELECT id, heading, created_at, updated_at, active FROM announcements WHERE active = TRUE ORDER BY created_at DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var announcement Announcement
		err := rows.Scan(&announcement.ID, &announcement.Heading, &announcement.CreatedAt, &announcement.UpdatedAt, &announcement.Active)
		if err != nil {
			return nil, err
		}
		announcements = append(announcements, announcement)
	}

	return announcements, nil
}

func GetAnnouncementByID(id int) (*Announcement, error) {
	var announcement Announcement
	query := `SELECT id, heading, created_at, updated_at, active FROM announcements WHERE id = ?`

	err := db.QueryRow(query, id).Scan(&announcement.ID, &announcement.Heading, &announcement.CreatedAt, &announcement.UpdatedAt, &announcement.Active)
	if err != nil {
		return nil, err
	}

	return &announcement, nil
}

func UpdateAnnouncement(id int, heading string) error {
	query := `UPDATE announcements SET heading = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := db.Exec(query, heading, id)
	return err
}

func DeleteAnnouncement(id int) error {
	query := `UPDATE announcements SET active = FALSE WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}

func ResetUserLevel(userEmail string) error {
	log.Printf("Resetting level for user %s", userEmail)
	result, err := db.Exec("UPDATE logins SET \"on\" = 1 WHERE gmail = ?", userEmail)
	if err != nil {
		log.Printf("ERROR: Failed to reset level for user %s: %v", userEmail, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user %s not found", userEmail)
	}

	notification := map[string]interface{}{
		"userEmail": userEmail,
		"message":   "Your level has been reset to Level 1 by an administrator",
		"type":      "info",
	}
	Create("notification", notification)

	log.Printf("Successfully reset level for user %s", userEmail)
	return nil
}

func GetUnreadNotificationCount(email string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM notifications WHERE user_email = ? AND read = FALSE`
	err := db.QueryRow(query, email).Scan(&count)
	return count, err
}

func MarkMessageDeletedByDiscordID(discordMsgID string) error {
	result, err := db.Exec("UPDATE lead_messages SET is_deleted = TRUE WHERE discord_msg_id = ?", discordMsgID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		_, err = db.Exec("UPDATE hint_messages SET is_deleted = TRUE WHERE discord_msg_id = ?", discordMsgID)
		if err != nil {
			return err
		}
	}

	return nil
}

func MarkAllMessagesDeletedForLevel(levelNumber int, messageType string) error {
	var query string

	if messageType == "lead" {
		query = "UPDATE lead_messages SET is_deleted = TRUE WHERE level_number = ?"
	} else if messageType == "hint" {
		query = "UPDATE hint_messages SET is_deleted = TRUE WHERE level_number = ?"
	} else {
		return fmt.Errorf("invalid message type: %s", messageType)
	}

	_, err := db.Exec(query, levelNumber)
	return err
}

func DeleteAllMessagesForLevel(levelNumber int, messageType string) error {
	var query string

	if messageType == "lead" {
		query = "DELETE FROM lead_messages WHERE level_number = ?"
	} else if messageType == "hint" {
		query = "DELETE FROM hint_messages WHERE level_number = ?"
	} else {
		return fmt.Errorf("invalid message type: %s", messageType)
	}

	_, err := db.Exec(query, levelNumber)
	return err
}

func GetLeadMessageByDiscordID(discordMsgID string) (*LeadMessage, error) {
	params := map[string]interface{}{
		"discordMsgId": discordMsgID,
	}
	result, err := Get("message_by_discord_id", params)
	if err != nil {
		return nil, err
	}
	if leadMsg, ok := result.(LeadMessage); ok {
		return &leadMsg, nil
	}
	return nil, fmt.Errorf("message not found")
}

func MarkHintMessageDeleted(discordMsgID string) error {
	_, err := db.Exec("DELETE FROM hint_messages WHERE discord_msg_id = ?", discordMsgID)
	return err
}

type SystemSetting struct {
	Key   string `json:"key" db:"key"`
	Value string `json:"value" db:"value"`
}

func GetChatStatus() string {
	result, err := Get("system_setting", map[string]interface{}{"key": "chat_status"})
	if err != nil {
		return "active" // default to active if not found
	}
	if setting, ok := result.(*SystemSetting); ok {
		return setting.Value
	}
	return "active"
}

func SetChatStatus(status string) error {
	// Check if setting exists
	_, err := Get("system_setting", map[string]interface{}{"key": "chat_status"})
	if err != nil {
		// Setting doesn't exist, create it
		return Create("system_setting", map[string]interface{}{
			"key":   "chat_status",
			"value": status,
		})
	} else {
		// Setting exists, update it
		return Update("system_setting", map[string]interface{}{"key": "chat_status"}, map[string]interface{}{
			"value": status,
		})
	}
}
