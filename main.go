package main

import (
	"github.com/gin-gonic/gin"
	"main/routes"
)

func main() {
	router := gin.Default()

	// Register all routes
	routes.RegisterRoutes(router)

	// Start server
	router.Run(":8080")
}
