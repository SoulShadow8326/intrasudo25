package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"sync"
	"time"

	"intrasudo25/config"
	"intrasudo25/database"

	"golang.org/x/crypto/bcrypt"
)

type CodeCooldown struct {
	mu       sync.RWMutex
	lastSent map[string]time.Time
}

var codeCooldown = &CodeCooldown{
	lastSent: make(map[string]time.Time),
}

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

	result, err := database.Get("login", map[string]interface{}{"gmail": gmail})
	if err == nil && result != nil {
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

	err = database.Create("login", Login{Gmail: gmail, Hashed: hashedPass, SeshTok: "", CSRFtok: "", Verified: false, VerificationNumber: fullVerificationCodeHash})

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
	emailConfig := config.GetEmailConfig()

	from := emailConfig.From
	pass := emailConfig.Password
	smtpHost := emailConfig.SMTPHost + ":" + emailConfig.SMTPPort

	msg := []byte("To: " + email + "\r\n" +
		"Subject: Exun Elite - Verification Code\r\n" +
		"\r\n" +
		"Your verification code (last 4 digits) is: " + codeToSend + "\r\n")

	err := smtp.SendMail(smtpHost,
		smtp.PlainAuth("", from, pass, emailConfig.SMTPHost),
		from, []string{email}, msg)

	if err != nil {
		return err
	}

	fmt.Println("Sent code (last 4 digits):", codeToSend, "to:", email)
	return nil
}

func sendLoginCodeEmail(email string, name string, loginCode string) error {
	userName := name
	if userName == "" {
		userName = "user"
	}

	emailConfig := config.GetEmailConfig()

	from := emailConfig.From
	pass := emailConfig.Password
	smtpHost := emailConfig.SMTPHost + ":" + emailConfig.SMTPPort

	greeting := "Hello,"
	if name != "" {
		greeting = "Hello " + name + ","
	}

	msg := []byte("To: " + email + "\r\n" +
		"Subject: Intra Sudo 2025 - Your Login Code\r\n" +
		"\r\n" +
		greeting + "\r\n" +
		"\r\n" +
		"Your permanent 4-digit login code for Intra Sudo 2025 is: " + loginCode + "\r\n" +
		"\r\n" +
		"Keep this code safe - you'll use it every time you log in.\r\n" +
		"\r\n" +
		"Good luck with the challenge!\r\n" +
		"- Exun Team\r\n")

	err := smtp.SendMail(smtpHost,
		smtp.PlainAuth("", from, pass, emailConfig.SMTPHost),
		from, []string{email}, msg)

	if err != nil {
		return err
	}

	fmt.Println("Sent login code:", loginCode, "to:", email, "for user:", userName)
	return nil
}

func Verify(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	gmail := r.FormValue("gmail")
	userProvidedCode := r.FormValue("vnum")

	result, err := database.Get("login", map[string]interface{}{"gmail": gmail})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Account not found or not registered"})
		return
	}
	acc := result.(*database.Login)

	storedFullVerificationHash := acc.VerificationNumber
	if len(storedFullVerificationHash) < 4 || storedFullVerificationHash[len(storedFullVerificationHash)-4:] != userProvidedCode {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Incorrect Verification Number;"})
		return
	}

	database.Update("login_field", map[string]interface{}{"gmail": gmail, "field": "verified"}, true)

	database.Create("leaderboard", Sucker{Gmail: gmail, Score: 0})
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

	result, err := database.Get("login", map[string]interface{}{"gmail": gmail})
	if err != nil || result == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Either; Gmail incorrect ; not verified ; password incorrect"})
		return
	}
	acc := result.(*database.Login)
	if !acc.Verified || !checkHash(acc.Hashed, password) {
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
	database.Update("login_field", map[string]interface{}{"gmail": gmail, "field": "seshTok"}, seshT)
	database.Update("login_field", map[string]interface{}{"gmail": gmail, "field": "CSRFtok"}, csrf)

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

func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "please use POST"})
		return
	}

	cookie, err := r.Cookie("exun_sesh_cookie")
	if err == nil && cookie.Value != "" {
		database.Update("login", map[string]interface{}{"seshTok": cookie.Value}, map[string]interface{}{
			"seshTok": "",
			"CSRFtok": "",
		})
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "exun_sesh_cookie",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "X-CSRF_COOKIE",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: false,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "logged out successfully"})
}

func EmailOnly(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	r.ParseForm()
	gmail := r.FormValue("gmail")

	if gmail == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email is required"})
		return
	}

	codeCooldown.mu.RLock()
	lastSent, exists := codeCooldown.lastSent[gmail]
	codeCooldown.mu.RUnlock()

	if exists && time.Since(lastSent) < 60*time.Second {
		remainingTime := 60 - int(time.Since(lastSent).Seconds())
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{
			"error":    fmt.Sprintf("Please wait %d seconds before requesting another code", remainingTime),
			"cooldown": "true",
		})
		return
	}

	_, err := database.Get("login", map[string]interface{}{"gmail": gmail})
	if err != nil {
		salt := generateSalt(16)
		codeSource := gmail + salt
		h := sha256.New()
		h.Write([]byte(codeSource))
		fullHash := fmt.Sprintf("%x", h.Sum(nil))
		permanentLoginCode := fullHash[len(fullHash)-4:]

		hashedPass, err := hash("email_verified_user")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Error creating account"})
			return
		}

		err = sendLoginCodeEmail(gmail, "", permanentLoginCode)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to send login code email"})
			return
		}

		codeCooldown.mu.Lock()
		codeCooldown.lastSent[gmail] = time.Now()
		codeCooldown.mu.Unlock()

		err = database.Create("login", Login{
			Gmail:              gmail,
			Name:               "",
			Hashed:             hashedPass,
			SeshTok:            "",
			CSRFtok:            "",
			Verified:           false,
			VerificationNumber: fullHash,
			LoginCode:          permanentLoginCode,
			On:                 1,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unable to create account"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Account created! Check your email for your permanent 4-digit login code."})
	} else {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message":       "Account found! Use your existing 4-digit login code.",
			"existing_user": "true",
		})
	}
}

func EmailVerify(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		r.ParseForm()
	}
	gmail := r.FormValue("gmail")
	userProvidedCode := r.FormValue("vnum")

	result, err := database.Get("login", map[string]interface{}{"gmail": gmail})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Account not found"})
		return
	}
	acc := result.(*database.Login)

	if acc.LoginCode != userProvidedCode {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Incorrect login code"})
		return
	}

	database.Update("login_field", map[string]interface{}{"gmail": gmail, "field": "verified"}, true)

	leaderboardResult, _ := database.Get("leaderboard", map[string]interface{}{"gmail": gmail})
	if leaderboardResult == nil {
		database.Create("leaderboard", Sucker{Gmail: gmail, Score: 0})
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
	database.Update("login_field", map[string]interface{}{"gmail": gmail, "field": "seshTok"}, seshT)
	database.Update("login_field", map[string]interface{}{"gmail": gmail, "field": "CSRFtok"}, csrf)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Verified and logged in successfully"})
}
