package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"intrasudo25/routes"
)

func init() {
	godotenv.Load()
}

func main() {
	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = gin.ReleaseMode
	}
	gin.SetMode(mode)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), cors.Default())

	routes.RegisterRoutes(router)

	router.Static("/static", "./static")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	address := fmt.Sprintf(":%s", port)
	log.Printf("Server running on %s", address)
	log.Fatal(router.Run(address))
}
