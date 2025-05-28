package handlers

import (
	"intrasudo25/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Authorize(c *gin.Context) (bool, Login) {
	cookie, err := c.Cookie("exun_sesh_cookie")
	if err != nil || cookie == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login Pending..."})
		return false, Login{}
	}
	acc, err := database.GetLoginFromCookie(cookie)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login Pending..."})
		return false, Login{}
	}
	csrf := c.GetHeader("CSRFtok")

	if csrf == "" || csrf != acc.CSRFtok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login Pending..."})
		return false, Login{}
	}

	return true, *acc
}

func AdminPriv(users []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		isAuth, user := Authorize(c)

		if !isAuth {
			c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden"})
			return
		}

		// Check if user is in allowed list
		allowed := false
		for _, u := range users {
			if u == user.Gmail {
				allowed = true
				break
			}
		}

		if !allowed {
			c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden"})
			return
		}

		c.Next() // allow request to continue
	}
}

