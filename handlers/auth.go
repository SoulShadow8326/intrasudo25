package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"os"

	//"strings"

	"intrasudo25/database"

	"golang.org/x/crypto/bcrypt"
)

type Login = database.Login
type Sucker = database.Sucker

func New(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "please use POST"})
		return
	}

	r.ParseForm()
	gmail := r.FormValue("gmail")
	password := r.FormValue("password")

	if _, err := database.GetLogin(gmail); err == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Gmail Taken"})
		return
	}

	salt := generateSalt(16)
	verificationCodeSource := password + salt
	h := sha256.New()
	h.Write([]byte(verificationCodeSource))
	fullVerificationCodeHash := fmt.Sprintf("%x", h.Sum(nil))

	verificationCodeForUser := fullVerificationCodeHash[len(fullVerificationCodeHash)-4:]

	hashedPass, err := hash(password)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error hashing password"})
		return
	}

	err = sendVerificationEmail(gmail, verificationCodeForUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to send verification email"})
		return
	}

	err = database.InsertLogin(Login{Gmail: gmail, Hashed: hashedPass, SeshTok: "", CSRFtok: "", Verified: false, VerificationNumber: fullVerificationCodeHash})

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Unable to add user..."})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Verification email sent. Please check your inbox for the last 4 digits of your verification code."})
}

func hash(pass string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), 2)
	return string(bytes), err
}

func sendVerificationEmail(email string, codeToSend string) error {
	// if !strings.HasSuffix(email, "@dpsrkp.net") {
	//     return fmt.Errorf("email must end with @dpsrkp.net")
	// }

	from := "e11383hursh@dpsrkp.net"
	pass := os.Getenv("pass")

	msg := []byte("To: " + email + "\r\n" +
		"Subject: Exun Elite - Verification Code\r\n" +
		"\r\n" +
		"Your verification code (last 4 digits) is: " + codeToSend + "\r\n")

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{email}, msg)

	if err != nil {
		return err
	}

	fmt.Println("Sent code (last 4 digits):", codeToSend, "to:", email)
	return nil
}

func Verify(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	gmail := r.FormValue("gmail")
	userProvidedCode := r.FormValue("vnum")

	acc, err := database.GetLogin(gmail)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Account not found or not registered"})
		return
	}

	storedFullVerificationHash := acc.VerificationNumber
	if len(storedFullVerificationHash) < 4 || storedFullVerificationHash[len(storedFullVerificationHash)-4:] != userProvidedCode {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Incorrect Verification Number;"})
		return
	}

	database.UpdateField(gmail, "Verified", true)

	database.InsertSucker(Sucker{Gmail: gmail, Score: 0})
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Welcome..."})
}

func LoginF(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "please use POST"})
		return
	}

	r.ParseForm()
	gmail := r.FormValue("gmail")
	password := r.FormValue("password")

	acc, err := database.GetLogin(gmail)
	if err != nil || !acc.Verified || !checkHash(acc.Hashed, password) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Either; Gmail incorrect ; not verified ; password incorrect"})
		return
	}

	seshT := generateTok(32)
	csrf := generateTok(32)
	http.SetCookie(w, &http.Cookie{
		Name:     "exun_sesh_cookie",
		Value:    seshT,
		MaxAge:   172800,
		Path:     "/",
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:   "X-CSRF_COOKIE",
		Value:  csrf,
		MaxAge: 172800,
		Path:   "/",
	})
	database.UpdateField(gmail, "SeshTok", seshT)
	database.UpdateField(gmail, "CSRFtok", csrf)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged in..."})
}

func checkHash(hash string, pass string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	return err == nil
}

func generateTok(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(bytes)
}

func generateSalt(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}
