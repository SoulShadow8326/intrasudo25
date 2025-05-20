package handlers

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"time"

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

type Login = database.Login;

var users = map[string]Login{}

func New(c *gin.Context) {
	if c.Request.Method != "POST" {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "please use POST"})
		return
	}
	gmail := c.PostForm("gmail");
	password := c.PostForm("password");

	if _, err := database.GetLogin(gmail); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gmail Taken"})
		return
	}

	hashedPass, err := hash(password);

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "please use POST"})
		return
	}

	vnum, err := validate(gmail)

	database.InsertLogin(Login{Hashed: hashedPass, SeshTok: "", CSRFtok: "", Gmail: gmail, Verified: false, VerificationNumber: uint(vnum)})

	c.JSON(http.StatusOK, gin.H{})
	return
}

func hash(pass string) (string, error){
	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), 2)
	return string(bytes), err
}

func validate(email string) (int, error) {
	// Uncomment during production to only allow DPS RKP emails
    // if !strings.HasSuffix(email, "@dpsrkp.net") {
    //     return 0, fmt.Errorf("email must end with @dpsrkp.net")
    //
	rand.Seed(time.Now().UnixNano())
	code := strconv.Itoa(100000 + rand.Intn(900000))

	from := "e11383hursh@dpsrkp.net"
	
	pass := os.Getenv("pass")

	msg := []byte("To: " + email + "\r\n" +
		"Subject: Exun Elite - Verification Code\r\n" +
		"\r\n" +
		"Your verification code is: " + code + "\r\n")

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{email}, msg)

	if err != nil {
		return 0, err
	}

	fmt.Println("Sent code:", code)
	num, err := strconv.Atoi(code) // returns (int, error)
    if err != nil {
        fmt.Println("Conversion error:", err)
        return 0, err
    }
	return num, nil
}

func Verify(c *gin.Context) {
	gmail := c.PostForm("gmail")
	Vnum := c.PostForm("vnum")
	vnum, err := strconv.ParseFloat(Vnum, 64)

	acc, err := database.GetLogin(gmail); 
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account not found"})
		return
	}
	if acc.VerificationNumber == uint(vnum) {
		database.UpdateField(gmail, "Verified", true)
	} else{
		database.DeleteLogin(gmail)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect Verification Number; Login Denied;"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Welcome..."})
	return

}

func LoginF(c *gin.Context) {
	if c.Request.Method != "POST" {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "please use POST"})
		return
	}

	gmail := c.PostForm("gmail");
	password := c.PostForm("password");

	acc, err := database.GetLogin(gmail)
	if err != nil || acc.Verified || checkHash(acc.Hashed, password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Either; Gmail incorrect ; not verified ; password incorrect"})
		return		
	}

	seshT := generateTok(32)
	c.SetCookie("exun_sesh_cookie", seshT, 172800, "/", "", false, true)
	database.UpdateField(gmail, "SeshTok", seshT)

	c.JSON(http.StatusOK, gin.H{"message": "Logged in..."})
	return
}

func checkHash(hash string, pass string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	return err == nil
}

func generateTok(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes);
	
	return base64.URLEncoding.EncodeToString(bytes)
}
