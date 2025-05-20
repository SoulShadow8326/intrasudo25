package routes

import (
	"github.com/gin-gonic/gin"
	"main/handlers"
)

func RegisterRoutes(r *gin.Engine) {
	r.GET("/", handlers.LandingPage)
	r.GET("/leaderboard", handlers.LeaderboardPage)
	r.GET("/questions/attempt", handlers.AttemptQuestionPage)
	r.GET("/dashboard", handlers.DashboardPage)

	// Placeholder for future chat route
	// r.GET("/chat", handlers.ChatHandler)
}
