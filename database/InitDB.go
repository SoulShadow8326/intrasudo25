package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Level struct {
	Markdown    string
	//LevelNumber int || auto increment
	SourceHint  string
	ConsoleHint string

	Answer string
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


