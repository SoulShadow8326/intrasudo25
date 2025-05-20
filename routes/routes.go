package routes

import (
	"intrasudo25/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.GET("/leaderboard", handlers.LeaderboardPage)
	r.GET("/questions/attempt", handlers.AttemptQuestionPage)
	r.GET("/dashboard", handlers.DashboardPage)

	// Placeholder for future chat route
	// r.GET("/chat", handlers.ChatHandler)

	r.POST("/enter/New", handlers.New)
	r.POST("/enter", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/enter/New")
	})

	r.POST("/enter/verify", handlers.Verify)
}
