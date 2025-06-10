package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	ports := []string{"9001", "9002", "9003"}
	delays := []int{50, 100, 150}
	var wg sync.WaitGroup

	for i, p := range ports {
		port := p
		delay := delays[i]
		wg.Add(1)

		go func() {
			defer wg.Done()
			mux := http.NewServeMux()

			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Duration(delay) * time.Millisecond)
				fmt.Fprintf(w, "Response from backend server on port %s\n", port)
			})

			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(2 * time.Second)
				fmt.Fprintf(w, "Slow response from backend on port %s\n", port)
			})

			mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			})

			addr := ":" + port
			log.Printf("Starting backend server on %s\n", addr)
			if err := http.ListenAndServe(addr, mux); err != nil {
				log.Fatalf("Backend server %s failed: %v", port, err)
			}
		}()
	}

	wg.Wait()
}
