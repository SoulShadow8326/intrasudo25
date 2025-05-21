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

type Question struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Answer   string `json:"answer"`
	ImageURL string `json:"imageUrl,omitempty"`
	TextClue string `json:"textClue"`
	Order    int    `json:"order"`
	Active   bool   `json:"active"`
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

	createQuestionsTable := `
    CREATE TABLE IF NOT EXISTS questions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        answer TEXT NOT NULL,
        image_url TEXT,
        text_clue TEXT NOT NULL,
        question_order INTEGER NOT NULL,
        active BOOLEAN DEFAULT TRUE
    );`
	_, err = db.Exec(createQuestionsTable)
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

// CreateQuestion adds a new question to the database
func CreateQuestion(q Question) (int64, error) {
	result, err := db.Exec(`
        INSERT INTO questions (title, answer, image_url, text_clue, question_order, active)
        VALUES (?, ?, ?, ?, ?, ?)`,
		q.Title, q.Answer, q.ImageURL, q.TextClue, q.Order, q.Active)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetQuestions retrieves all questions from the database
func GetQuestions() ([]Question, error) {
	rows, err := db.Query(`SELECT id, title, answer, image_url, text_clue, question_order, active FROM questions ORDER BY question_order ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []Question
	for rows.Next() {
		var q Question
		err := rows.Scan(&q.ID, &q.Title, &q.Answer, &q.ImageURL, &q.TextClue, &q.Order, &q.Active)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	return questions, nil
}

// GetQuestion retrieves a specific question by ID
func GetQuestion(id int) (*Question, error) {
	row := db.QueryRow(`SELECT id, title, answer, image_url, text_clue, question_order, active FROM questions WHERE id = ?`, id)

	var q Question
	err := row.Scan(&q.ID, &q.Title, &q.Answer, &q.ImageURL, &q.TextClue, &q.Order, &q.Active)
	if err != nil {
		return nil, err
	}

	return &q, nil
}

// UpdateQuestion updates an existing question
func UpdateQuestion(q Question) error {
	_, err := db.Exec(`
        UPDATE questions 
        SET title = ?, answer = ?, image_url = ?, text_clue = ?, question_order = ?, active = ?
        WHERE id = ?`,
		q.Title, q.Answer, q.ImageURL, q.TextClue, q.Order, q.Active, q.ID)

	return err
}

// DeleteQuestion removes a question by ID
func DeleteQuestion(id int) error {
	_, err := db.Exec(`DELETE FROM questions WHERE id = ?`, id)
	return err
}

// GetActiveQuestion retrieves the question at the specified order position that is active
func GetActiveQuestion(order int) (*Question, error) {
	row := db.QueryRow(`SELECT id, title, answer, image_url, text_clue, question_order, active FROM questions WHERE question_order = ? AND active = true`, order)

	var q Question
	err := row.Scan(&q.ID, &q.Title, &q.Answer, &q.ImageURL, &q.TextClue, &q.Order, &q.Active)
	if err != nil {
		return nil, err
	}

	return &q, nil
}
