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

	r.POST("/enter/login", handlers.LoginF)

	// Playground/Game routes
	r.GET("/api/question", handlers.GetQuestionHandler)
	r.POST("/api/submit", handlers.SubmitAnswer)

	// Dashboard API routes
	dashboard := r.Group("/api/dashboard")
	{
		//dashboard.GET("/questions", handlers.GetQuestionsHandler)
		dashboard.GET("/questions/:id", handlers.GetQuestionHandler)
	}

	// Admin Panel API routes - for level management
	admin := r.Group("/api/admin")
	{
		admin.GET("/", handlers.AdminPanelHandler)
		admin.GET("/levels", handlers.GetAllLevelsHandler)
		admin.POST("/levels", handlers.CreateLvlHandler)
		admin.PUT("/levels/:id", handlers.UpdateLvlHandler)
		admin.DELETE("/levels/:id", handlers.DeleteLvlHandler)
	}
}
