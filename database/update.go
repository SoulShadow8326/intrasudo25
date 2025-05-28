package database

import (
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func UpdateField(gmail string, field string, value interface{}) error {
	query := fmt.Sprintf("UPDATE logins SET %s = ? WHERE gmail = ?", field)
	_, err := db.Exec(query, value, gmail)
	return err
}

func UpdateLevel(number int, l Level) error {
	_, err := db.Exec(`
        UPDATE levels 
        SET markdown = ?, src_hint = ?, console_hint = ?, answer = ?, active = ?
        WHERE level_number = ?`,
		l.Markdown, l.SourceHint, l.ConsoleHint, l.Answer, l.Active, number)

	return err
}

func UpdateScore(gmail string, score int) error {
	_, err := db.Exec(`INSERT INTO leaderboard (gmail, score) VALUES (?, ?) ON CONFLICT(gmail) DO UPDATE SET score = excluded.score`, gmail, score)
	return err
}
