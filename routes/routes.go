package routes

import (
	"intrasudo25/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.GET("/leaderboard", handlers.LeaderboardPage)
	r.GET("/questions/attempt", handlers.AttemptQuestionPage)
	r.GET("/dashboard", handlers.DashboardPage)

	// Placeholder for future chat route
	// r.GET("/chat", handlers.ChatHandler)

	r.GET("/enter/:value", func(c *gin.Context) {
		path := c.Param("value");
		if path == "" {
			c.Redirect(302, "/enter/new");
		}
		switch path {
			case "New":
				handlers.New(c)
		}
	})
}
