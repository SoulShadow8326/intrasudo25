package database

import (
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

func CreateLevel(q Level) (int64, error) {
	result, err := db.Exec(`
        INSERT INTO levels (markdown, src_hint, console_hint, answer, active)
        VALUES (?, ?, ?, ?, ?)`,
		q.Markdown, q.SourceHint, q.ConsoleHint, q.Answer, q.Active)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
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
