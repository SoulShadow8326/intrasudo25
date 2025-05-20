package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	
)

type Login struct {
	Hashed string
	SeshTok string
	CSRFtok string

	Gmail string
	Verified bool
	VerificationNumber uint
}

var db *sql.DB

func InitDB() {
	var err error
	db, err = sql.Open("sqlite3", "./logins.db")
	if err != nil {
		log.Fatal(err)
	}
	createTable := `
	CREATE TABLE IF NOT EXISTS logins (
		hashed TEXT PRIMARY KEY,
		seshTok TEXT,
		CSRFtok TEXT,
		gmail TEXT,
		verified BOOLEAN,
		verificationNumber INTEGER
	);`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}
}

func InsertLogin(l Login) error {
	_, err := db.Exec(`
		INSERT INTO logins (hashed, seshTok, CSRFtok, gmail, verified, verificationNumber)
		VALUES (?, ?, ?, ?, ?, ?)`,
		l.Hashed, l.SeshTok, l.CSRFtok, l.Gmail, l.Verified, l.VerificationNumber)
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

