package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"intrasudo25/database"
	"intrasudo25/routes"
)

func init() {
	godotenv.Load()
}

func main() {
	database.InitDB()

	handler := routes.RegisterRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	address := fmt.Sprintf(":%s", port)
	log.Printf("Server running on %s", address)
	log.Fatal(http.ListenAndServe(address, handler))
}
