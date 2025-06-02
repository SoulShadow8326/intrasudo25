package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"intrasudo25/database"
	"intrasudo25/routes"
)

func init() {
	godotenv.Load()
}

func main() {
	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	database.InitDB()

	handler := routes.RegisterRoutes()

	address := fmt.Sprintf(":%s", *port)
	log.Printf("Server running on %s", address)
	log.Fatal(http.ListenAndServe(address, handler))
}
