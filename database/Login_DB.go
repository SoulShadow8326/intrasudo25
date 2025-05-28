package database

import (
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func InsertLogin(l Login) error {
	_, err := db.Exec(`
        INSERT INTO logins (gmail, hashed, seshTok, CSRFtok, verified, verificationNumber, "on")
        VALUES (?, ?, ?, ?, ?, ?, ?)`,
		l.Gmail, l.Hashed, l.SeshTok, l.CSRFtok, l.Verified, l.VerificationNumber, l.On)
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT OR IGNORE INTO leaderboard (gmail, score) VALUES (?, 0)`, l.Gmail)
	return err
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

func UpdateField(gmail string, field string, value interface{}) error {
	query := fmt.Sprintf("UPDATE logins SET %s = ? WHERE gmail = ?", field)
	_, err := db.Exec(query, value, gmail)
	return err
}

func DeleteLogin(gmail string) error {
	_, err := db.Exec(`DELETE FROM logins WHERE gmail = ?`, gmail)
	return err
}

//Score Related

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



