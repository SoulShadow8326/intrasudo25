package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

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


