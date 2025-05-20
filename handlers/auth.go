package handlers

import (
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
