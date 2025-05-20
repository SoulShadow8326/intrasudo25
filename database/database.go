package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Login struct {
	Hashed             string
	SeshTok            string
	CSRFtok            string
	Gmail              string
	Verified           bool
	VerificationNumber uint
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
		hashed TEXT PRIMARY KEY,
		seshTok TEXT,
		CSRFtok TEXT,
		gmail TEXT,
		verified BOOLEAN,
		verificationNumber INTEGER
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
}

func InsertLogin(l Login) error {
	_, err := db.Exec(`
		INSERT INTO logins (hashed, seshTok, CSRFtok, gmail, verified, verificationNumber)
		VALUES (?, ?, ?, ?, ?, ?)`,
		l.Hashed, l.SeshTok, l.CSRFtok, l.Gmail, l.Verified, l.VerificationNumber)
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT OR IGNORE INTO leaderboard (gmail, score) VALUES (?, 0)`, l.Gmail)
	return err
}

func GetLogin(gmail string) (*Login, error) {
	row := db.QueryRow(`SELECT hashed, seshTok, CSRFtok, gmail, verified, verificationNumber FROM logins WHERE gmail = ?`, gmail)
	var l Login
	err := row.Scan(&l.Hashed, &l.SeshTok, &l.CSRFtok, &l.Gmail, &l.Verified, &l.VerificationNumber)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func GetLoginFromCookie(tok string) (*Login, error) {
	row := db.QueryRow(`SELECT hashed, seshTok, CSRFtok, gmail, verified, verificationNumber FROM logins WHERE seshTok = ?`, tok)
	var l Login
	err := row.Scan(&l.Hashed, &l.SeshTok, &l.CSRFtok, &l.Gmail, &l.Verified, &l.VerificationNumber)
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
	rows, err := db.Query(`SELECT gmail, score FROM leaderboard ORDER BY score DESC LIMIT ?`, n)
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
