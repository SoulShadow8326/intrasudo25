package routes

import (
	"intrasudo25/handlers"
	"net/http"
	"strings"
)

func RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/leaderboard", handlers.LeaderboardPage)
	mux.HandleFunc("/questions/attempt", handlers.AttemptQuestionPage)
	mux.HandleFunc("/dashboard", handlers.DashboardPage)
	mux.HandleFunc("/chat", handlers.ChatHandler)

	mux.HandleFunc("/enter/New", handlers.New)
	mux.HandleFunc("/enter", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/enter/New", http.StatusPermanentRedirect)
	})

	mux.HandleFunc("/enter/verify", handlers.Verify)
	mux.HandleFunc("/enter/login", handlers.LoginF)

	mux.HandleFunc("/api/question", handlers.GetQuestionHandler)
	mux.HandleFunc("/api/submit", handlers.SubmitAnswer)

	mux.HandleFunc("/api/admin/", func(w http.ResponseWriter, r *http.Request) {
		if !handlers.AdminAuth(w, r, []string{}) {
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/api/admin")
		if path == "/" || path == "" {
			handlers.AdminPanelHandler(w, r)
			return
		}

		if strings.HasPrefix(path, "/levels") {
			levelPath := strings.TrimPrefix(path, "/levels")
			if levelPath == "" || levelPath == "/" {
				if r.Method == "GET" {
					handlers.GetAllLevelsHandler(w, r)
				} else if r.Method == "POST" {
					handlers.CreateLvlHandler(w, r)
				}
			} else {
				id := strings.TrimPrefix(levelPath, "/")
				if r.Method == "PUT" {
					handlers.UpdateLvlHandler(w, r, id)
				} else if r.Method == "DELETE" {
					handlers.DeleteLvlHandler(w, r, id)
				}
			}
		}
	})

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	return mux
}
