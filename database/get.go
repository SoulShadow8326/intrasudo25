package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Level struct {
	Markdown    string
	SourceHint  string
	ConsoleHint string
	Answer      string
	Active      bool
}

type Login struct {
	Hashed             string
	SeshTok            string
	CSRFtok            string
	Gmail              string
	Verified           bool
	VerificationNumber string
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
        gmail TEXT PRIMARY KEY,
        hashed TEXT NOT NULL,
        seshTok TEXT,
        CSRFtok TEXT,
        verified BOOLEAN,
        verificationNumber TEXT,
		"on" INTEGER
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

func GetLogin(gmail string) (*Login, error) {
	row := db.QueryRow(`SELECT gmail, hashed, seshTok, CSRFtok, verified, verificationNumber, "on" FROM logins WHERE gmail = ?`, gmail)
	var l Login
	err := row.Scan(&l.Gmail, &l.Hashed, &l.SeshTok, &l.CSRFtok, &l.Verified, &l.VerificationNumber, &l.On)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func GetLoginFromCookie(tok string) (*Login, error) {
	row := db.QueryRow(`SELECT gmail, hashed, seshTok, CSRFtok, verified, verificationNumber, "on" FROM logins WHERE seshTok = ?`, tok)
	var l Login
	err := row.Scan(&l.Gmail, &l.Hashed, &l.SeshTok, &l.CSRFtok, &l.Verified, &l.VerificationNumber, &l.On)
	if err != nil {
		return nil, err
	}
	return &l, nil
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
	var rows *sql.Rows
	var err error

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

func GetLevels() ([]Level, error) {
	rows, err := db.Query(`SELECT level_number, markdown, src_hint, console_hint, answer, active FROM levels ORDER BY level_number ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var levels []Level
	for rows.Next() {
		var l Level
		err := rows.Scan(new(int), &l.Markdown, &l.SourceHint, &l.ConsoleHint, &l.Answer, &l.Active)
		if err != nil {
			return nil, err
		}
		levels = append(levels, l)
	}
	return levels, nil
}

func GetLevel(number int) (*Level, error) {
	row := db.QueryRow(`SELECT level_number, markdown, src_hint, console_hint, answer, active FROM levels WHERE level_number = ?`, number)

	var l Level
	err := row.Scan(new(int), &l.Markdown, &l.SourceHint, &l.ConsoleHint, &l.Answer, &l.Active)
	if err != nil {
		return nil, err
	}

	return &l, nil
}
