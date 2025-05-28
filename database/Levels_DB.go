package database

import (
	_ "github.com/mattn/go-sqlite3"
)

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

func GetLevel(number int) (*Level, error) {
	row := db.QueryRow(`SELECT level_number, markdown, src_hint, console_hint, answer, active FROM levels WHERE level_number = ?`, number)

	var l Level
	err := row.Scan(new(int), &l.Markdown, &l.SourceHint, &l.ConsoleHint, &l.Answer, &l.Active) // level_number ignored
	if err != nil {
		return nil, err
	}

	return &l, nil
}

func UpdateLevel(number int, l Level) error {
	_, err := db.Exec(`
        UPDATE levels 
        SET markdown = ?, src_hint = ?, console_hint = ?, answer = ?, active = ?
        WHERE level_number = ?`,
		l.Markdown, l.SourceHint, l.ConsoleHint, l.Answer, l.Active, number)

	return err
}

func DeleteLevel(number int) error {
	_, err := db.Exec(`DELETE FROM levels WHERE level_number = ?`, number)
	return err
}


