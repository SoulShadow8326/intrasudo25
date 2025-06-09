package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"intrasudo25/database"
	"intrasudo25/routes"
)

func init() {
	godotenv.Load()
}

func main() {
	socketPath := flag.String("socket", "/tmp/intrasudo25.sock", "Unix socket path")
	flag.Parse()

	database.InitDB()

	handler := routes.RegisterRoutes()

	os.Remove(*socketPath)

	listener, err := net.Listen("unix", *socketPath)
	if err != nil {
		log.Fatalf("Failed to create unix socket: %v", err)
	}
	defer listener.Close()
	defer os.Remove(*socketPath)

	os.Chmod(*socketPath, 0666)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		listener.Close()
		os.Remove(*socketPath)
		os.Exit(0)
	}()

	log.Printf("Server running on unix socket %s", *socketPath)
	log.Fatal(http.Serve(listener, handler))
}
