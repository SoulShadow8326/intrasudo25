package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	os.Chmod(*socketPath, 0666)

	server := &http.Server{Handler: handler}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		listener.Close()
		os.Remove(*socketPath)
	}()

	log.Printf("Server running on unix socket %s", *socketPath)
	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
