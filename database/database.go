package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Level struct {
	Markdown string
	//LevelNumber int || auto increment
	SourceHint string
	ConsoleHint string

	Answer   string 
	Active bool
}

type Login struct {
	Hashed             string
	SeshTok            string
	CSRFtok            string
	Gmail              string
	Verified           bool
	VerificationNumber string // Changed from uint to string
	On                 uint
}

type Sucker struct {
	Gmail string
	Score int
}

var db *sql.DB

func InitDB() {
	var err error
	db, err = sql.Open("sqlite3", "./logins.db")
	if err != nil {
		log.Fatal(err)
	}
	createLoginsTable := `
    CREATE TABLE IF NOT EXISTS logins (
        gmail TEXT PRIMARY KEY,      -- Changed: gmail is now PRIMARY KEY
        hashed TEXT NOT NULL,        -- Changed: no longer PRIMARY KEY
        seshTok TEXT,
        CSRFtok TEXT,
        verified BOOLEAN,
        verificationNumber TEXT,     -- Changed: type to TEXT
		On INTEGER
    );`
	_, err = db.Exec(createLoginsTable)
	if err != nil {
		log.Fatal(err)
	}

	createLeaderboardTable := `
    CREATE TABLE IF NOT EXISTS leaderboard (
        gmail TEXT PRIMARY KEY,
        score INTEGER
    );`
	_, err = db.Exec(createLeaderboardTable)
	if err != nil {
		log.Fatal(err)
	}

	createLevelsTable := `
    CREATE TABLE IF NOT EXISTS levels (
        level_number INTEGER PRIMARY KEY AUTOINCREMENT,
        markdown TEXT,
		src_hint TEXT,
		console_hint TEXT,
        answer TEXT NOT NULL,
        active BOOLEAN DEFAULT TRUE
    );`
	_, err = db.Exec(createLevelsTable)
	if err != nil {
		log.Fatal(err)
	}
}

func InsertLogin(l Login) error {
	_, err := db.Exec(`
        INSERT INTO logins (gmail, hashed, seshTok, CSRFtok, verified, verificationNumber, On)
        VALUES (?, ?, ?, ?, ?, ?, ?)`, /* Removed l.Hashed from first value, added l.Gmail */
		l.Gmail, l.Hashed, l.SeshTok, l.CSRFtok, l.Verified, l.VerificationNumber, 1) // Added l.Gmail
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT OR IGNORE INTO leaderboard (gmail, score) VALUES (?, 0)`, l.Gmail)
	return err
}

func GetLogin(gmail string) (*Login, error) {
	row := db.QueryRow(`SELECT gmail, hashed, seshTok, CSRFtok, verified, verificationNumber, On FROM logins WHERE gmail = ?`, gmail) // Added gmail to SELECT
	var l Login
	err := row.Scan(&l.Gmail, &l.Hashed, &l.SeshTok, &l.CSRFtok, &l.Gmail, &l.Verified, &l.VerificationNumber, &l.On) // Added &l.Gmail to scan
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func GetLoginFromCookie(tok string) (*Login, error) {
	row := db.QueryRow(`SELECT gmail, hashed, seshTok, CSRFtok, verified, verificationNumber, On FROM logins WHERE seshTok = ?`, tok) // Added gmail to SELECT
	var l Login
	err := row.Scan(&l.Gmail, &l.Hashed, &l.SeshTok, &l.CSRFtok, &l.Gmail, &l.Verified, &l.VerificationNumber, &l.On) // Added &l.Gmail to scan
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func UpdateField(gmail string, field string, value interface{}) error {
	query := fmt.Sprintf("UPDATE logins SET %s = ? WHERE gmail = ?", field)
	_, err := db.Exec(query, value, gmail)
	return err
}

func DeleteLogin(gmail string) error {
	_, err := db.Exec(`DELETE FROM logins WHERE gmail = ?`, gmail)
	return err
}

func UpdateScore(gmail string, score int) error {
	_, err := db.Exec(`INSERT INTO leaderboard (gmail, score) VALUES (?, ?) ON CONFLICT(gmail) DO UPDATE SET score = excluded.score`, gmail, score)
	return err
}

func GetUserScore(gmail string) (int, error) {
	row := db.QueryRow(`SELECT score FROM leaderboard WHERE gmail = ?`, gmail)
	var score int
	err := row.Scan(&score)
	if err != nil {
		return 0, err
	}
	return score, nil
}

func GetLeaderboardTop(n int) ([]Sucker, error) {
	
	var rows *sql.Rows;
	var err error;

	if n == 0 {
		rows, err = db.Query(`SELECT gmail, score FROM leaderboard ORDER BY score DESC`) 

	} else {
		rows, err = db.Query(`SELECT gmail, score FROM leaderboard ORDER BY score DESC LIMIT ?`, n) 

	}	

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var entries []Sucker
	for rows.Next() {
		var entry Sucker
		err := rows.Scan(&entry.Gmail, &entry.Score)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func InsertSucker(s Sucker) error {
	_, err := db.Exec(`
        INSERT INTO leaderboard (gmail, score)
        VALUES (?, ?)`,
		s.Gmail, s.Score)
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT OR IGNORE INTO leaderboard (gmail, score) VALUES (?, 0)`, s.Gmail)
	return err
}

// CreateQuestion adds a new question to the database
func CreateLevel(q Level) (int64, error) {
	result, err := db.Exec(`
        INSERT INTO levels (markdown, src_hint, console_hint, answer, active)
        VALUES (?, ?, ?, ?, ?, ?)`,
		q.Markdown, q.SourceHint, q.ConsoleHint, q.Answer, q.Active)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetLevels retrieves all levels from the database
func GetLevels() ([]Level, error) {
	rows, err := db.Query(`SELECT level_number, markdown, src_hint, console_hint, answer, active FROM levels ORDER BY level_number ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var levels []Level
	for rows.Next() {
		var l Level
		err := rows.Scan(new(int), &l.Markdown, &l.SourceHint, &l.ConsoleHint, &l.Answer, &l.Active) // level_number ignored
		if err != nil {
			return nil, err
		}
		levels = append(levels, l)
	}
	return levels, nil
}

// GetLevel retrieves a specific level by number
func GetLevel(number int) (*Level, error) {
	row := db.QueryRow(`SELECT level_number, markdown, src_hint, console_hint, answer, active FROM levels WHERE level_number = ?`, number)

	var l Level
	err := row.Scan(new(int), &l.Markdown, &l.SourceHint, &l.ConsoleHint, &l.Answer, &l.Active) // level_number ignored
	if err != nil {
		return nil, err
	}

	return &l, nil
}

// UpdateLevel updates an existing level
func UpdateLevel(number int, l Level) error {
	_, err := db.Exec(`
        UPDATE levels 
        SET markdown = ?, src_hint = ?, console_hint = ?, answer = ?, active = ?
        WHERE level_number = ?`,
		l.Markdown, l.SourceHint, l.ConsoleHint, l.Answer, l.Active, number)

	return err
}

// DeleteLevel removes a level by number
func DeleteLevel(number int) error {
	_, err := db.Exec(`DELETE FROM levels WHERE level_number = ?`, number)
	return err
}

// GetActiveLevel retrieves the active level by number

/*
func GetActiveLevel(number int) (*Level, error) {
	row := db.QueryRow(`SELECT level_number, markdown, src_hint, console_hint, answer, active FROM levels WHERE level_number = ? AND active = true`, number)

	var l Level
	err := row.Scan(new(int), &l.Markdown, &l.SourceHint, &l.ConsoleHint, &l.Answer, &l.Active) // level_number ignored
	if err != nil {
		return nil, err
	}

	return &l, nil
}
*/ //recreate in service; fetch logged in user -> get his current q -> fetch that q; ++ ans submit logic in service as well
