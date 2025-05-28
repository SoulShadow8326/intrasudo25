package database

import (
	_ "github.com/mattn/go-sqlite3"
)

func DeleteLogin(gmail string) error {
	_, err := db.Exec(`DELETE FROM logins WHERE gmail = ?`, gmail)
	return err
}

func DeleteLevel(number int) error {
	_, err := db.Exec(`DELETE FROM levels WHERE level_number = ?`, number)
	return err
}
