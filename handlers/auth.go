package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/resend/resend-go/v2"

	"intrasudo25/config"
	"intrasudo25/database"

	"golang.org/x/crypto/bcrypt"
)

const td = `<td style="border-radius: 0.25rem; width: 42px; height: 42px; background-color: rgb(241, 245, 249); text-align: center; vertical-align: middle; font-size: 1.5rem; line-height: 2rem; font-weight: 800; color: rgb(15, 23, 42);">{digit}</td>`

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
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request method"})
		return
	}

	r.ParseForm()
	gmail := r.FormValue("gmail")
	password := r.FormValue("password")

	adminEmails := config.GetAdminEmails()
	isAdmin := false
	for _, adminEmail := range adminEmails {
		if strings.EqualFold(gmail, adminEmail) {
			isAdmin = true
			break
		}
	}

	if !isAdmin && !strings.HasSuffix(gmail, "@dpsrkp.net") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Please use your DPS RK Puram email address (@dpsrkp.net)"})
		return
	}

	result, err := database.Get("login", map[string]interface{}{"gmail": gmail})
	if err == nil && result != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "This email address is already registered"})
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
		json.NewEncoder(w).Encode(map[string]string{"error": "Registration failed. Please try again"})
		return
	}

	err = sendVerificationEmail(gmail, verificationCodeForUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unable to send verification email. Please try again"})
		return
	}

	err = database.Create("login", Login{Gmail: gmail, Hashed: hashedPass, SeshTok: "", CSRFtok: "", Verified: false, VerificationNumber: fullVerificationCodeHash})

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Registration failed. Please try again"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Verification email sent! Please check your inbox for the verification code"})
}

func hash(pass string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), 2)
	return string(bytes), err
}

func sendVerificationEmail(email string, codeToSend string) error {
	otp, _ := os.ReadFile("./frontend/otp.html")
	otp_str := string(otp)
	digits := ""
	for i := 0; i < 8; i++ {
		digits += strings.ReplaceAll(td, "{digit}", string(codeToSend[i]))
	}
	otp_str = strings.ReplaceAll(otp_str, "{otp}", digits)
	from := config.GetEmailConfig().From
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		fmt.Println("RESEND_API_KEY not set")
		return fmt.Errorf("RESEND_API_KEY not set")
	}
	client := resend.NewClient(apiKey)
	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{email},
		Subject: "Intra Sudo v6.0 - Verification Code",
		Html:    otp_str,
	}
	_, err := client.Emails.Send(params)
	if err != nil {
		fmt.Println("Resend sendVerificationEmail error:", err)
	}
	return err
}

func Verify(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	gmail := r.FormValue("gmail")
	userProvidedCode := r.FormValue("vnum")

	result, err := database.Get("login", map[string]interface{}{"gmail": gmail})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Account not found. Please register first"})
		return
	}
	acc := result.(*database.Login)

	storedFullVerificationHash := acc.VerificationNumber
	if len(storedFullVerificationHash) < 4 || storedFullVerificationHash[len(storedFullVerificationHash)-4:] != userProvidedCode {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid verification code. Please check your email and try again"})
		return
	}

	database.Update("login_field", map[string]interface{}{"gmail": gmail, "field": "verified"}, true)

	database.Create("leaderboard", Sucker{Gmail: gmail, Score: 0})
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Account verified successfully! You can now log in"})
}

func LoginF(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request method"})
		return
	}

	r.ParseForm()
	gmail := r.FormValue("gmail")
	password := r.FormValue("password")

	result, err := database.Get("login", map[string]interface{}{"gmail": gmail})
	if err != nil || result == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid email or password"})
		return
	}
	acc := result.(*database.Login)
	if !acc.Verified || !checkHash(acc.Hashed, password) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid email or password"})
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
	json.NewEncoder(w).Encode(map[string]string{"message": "Successfully logged in"})
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
	json.NewEncoder(w).Encode(map[string]string{"message": "Successfully logged out"})
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

	isBanned, banErr := database.IsEmailBanned(gmail)
	if banErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}
	if isBanned {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "This email has been banned from the platform"})
		return
	}

	adminEmails := config.GetAdminEmails()
	isAdmin := false
	for _, adminEmail := range adminEmails {
		if strings.EqualFold(gmail, adminEmail) {
			isAdmin = true
			break
		}
	}

	if !isAdmin && !strings.HasSuffix(gmail, "@dpsrkp.net") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Please use your DPS RK Puram email address (@dpsrkp.net)"})
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
		permanentLoginCode := fullHash[len(fullHash)-8:]

		hashedPass, err := hash("email_verified_user")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Registration failed. Please try again"})
			return
		}

		err = sendVerificationEmail(gmail, permanentLoginCode)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unable to send login code. Please try again"})
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
			json.NewEncoder(w).Encode(map[string]string{"error": "Registration failed. Please try again"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Account created successfully! Check your email for your login code"})
	} else {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message":       "Welcome back! Please enter your login code",
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
		json.NewEncoder(w).Encode(map[string]string{"error": "Account not found. Please register first"})
		return
	}
	acc := result.(*database.Login)

	if acc.LoginCode != userProvidedCode {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid login code. Please check your email and try again"})
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
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful! Welcome to Intra Sudo"})
}
