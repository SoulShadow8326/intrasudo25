package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/smtp"
	"os"

	"intrasudo25/database"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	//"strings"
)

/*
type Login struct {
	Hashed string
	SeshTok string
	CSRFtok string

	Gmail string
	Verified bool
	VerificationNumber uint
}
*/

type Login = database.Login
type Sucker = database.Sucker

func New(c *gin.Context) {
	if c.Request.Method != "POST" {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "please use POST"})
		return
	}
	gmail := c.PostForm("gmail")
	password := c.PostForm("password")

	if _, err := database.GetLogin(gmail); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gmail Taken"})
		return
	}

	// Generate a salted hash for the verification code
	salt := generateSalt(16)                  // Generate a 16-byte salt
	verificationCodeSource := password + salt // Combine password and salt
	h := sha256.New()
	h.Write([]byte(verificationCodeSource))
	fullVerificationCodeHash := fmt.Sprintf("%x", h.Sum(nil))

	// The code sent to the user is the last 4 digits of the hash
	verificationCodeForUser := fullVerificationCodeHash[len(fullVerificationCodeHash)-4:]

	hashedPass, err := hash(password) // This is for storing the login password, not the verification code

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	err = sendVerificationEmail(gmail, verificationCodeForUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email"})
		return
	}

	err = database.InsertLogin(Login{Gmail: gmail, Hashed: hashedPass, SeshTok: "", CSRFtok: "", Verified: false, VerificationNumber: fullVerificationCodeHash}) // Store the full hash

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to add user..."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification email sent. Please check your inbox for the last 4 digits of your verification code."})
}

func hash(pass string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), 2)
	return string(bytes), err
}

// Renamed from validate to sendVerificationEmail and modified
func sendVerificationEmail(email string, codeToSend string) error {
	// Uncomment during production to only allow DPS RKP emails
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

func Verify(c *gin.Context) {
	gmail := c.PostForm("gmail")
	userProvidedCode := c.PostForm("vnum") // This is the 4-digit code from the user

	acc, err := database.GetLogin(gmail)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account not found or not registered"})
		return
	}

	// Check if the last 4 digits of the stored hash match the user-provided code
	storedFullVerificationHash := acc.VerificationNumber
	if len(storedFullVerificationHash) < 4 || storedFullVerificationHash[len(storedFullVerificationHash)-4:] != userProvidedCode {
		// Potentially delete the login attempt or implement a retry limit
		// database.DeleteLogin(gmail) // Decided against auto-deletion for now, could lock out users.
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect Verification Number;"})
		return
	}

	database.UpdateField(gmail, "Verified", true)

	database.InsertSucker(Sucker{Gmail: gmail, Score: 0})
	c.JSON(http.StatusOK, gin.H{"message": "Welcome..."})
}

func LoginF(c *gin.Context) {
	if c.Request.Method != "POST" {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "please use POST"})
		return
	}

	gmail := c.PostForm("gmail")
	password := c.PostForm("password")

	acc, err := database.GetLogin(gmail)
	if err != nil || !acc.Verified || !checkHash(acc.Hashed, password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Either; Gmail incorrect ; not verified ; password incorrect"})
		return
	}

	seshT := generateTok(32)
	csrf := generateTok(32)
	c.SetCookie("exun_sesh_cookie", seshT, 172800, "/", "", false, true)
	c.SetCookie("X-CSRF_COOKIE", csrf, 172800, "/", "", false, false)
	database.UpdateField(gmail, "SeshTok", seshT)
	database.UpdateField(gmail, "CSRFtok", csrf)

	c.JSON(http.StatusOK, gin.H{"message": "Logged in..."})
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


