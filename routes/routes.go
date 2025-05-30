package routes

import (
	"intrasudo25/config"
	"intrasudo25/database"
	"intrasudo25/handlers"
	"net/http"
	"strings"
)

type CustomHandler struct {
	mux *http.ServeMux
}

func (h *CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if the request path has a registered handler
	_, pattern := h.mux.Handler(r)
	if pattern == "" {
		// No handler found, serve 404 page
		handlers.NotFoundHandler(w, r)
		return
	}

	// Serve the request normally
	h.mux.ServeHTTP(w, r)
}

func RegisterRoutes() *CustomHandler {
	mux := http.NewServeMux()

	// Use dynamic adminEmails from handlers package
	// Remove hardcoded adminEmails

	mux.HandleFunc("/landing", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/landing.html")
	})
	mux.HandleFunc("/home", handlers.IndexHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/landing", http.StatusSeeOther)
	})
	mux.HandleFunc("/auth", handlers.AuthPageHandler)
	mux.HandleFunc("/404", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/404.html")
	})

	mux.HandleFunc("/leaderboard", handlers.RequireAuth(handlers.LeaderboardHandler))
	mux.HandleFunc("/hints", handlers.RequireAuth(handlers.HintsHandler))
	mux.HandleFunc("/chat", handlers.RequireAuth(handlers.ChatPageHandler))

	// Use config.GetAdminEmails() for all admin routes
	mux.HandleFunc("/admin", handlers.RequireAdmin(config.GetAdminEmails())(handlers.AdminDashboardHandler))
	mux.HandleFunc("/admin/levels/new", handlers.RequireAdmin(config.GetAdminEmails())(handlers.NewLevelFormHandler))
	mux.HandleFunc("/submit", handlers.RequireAuth(handlers.SubmitAnswerFormHandler))
	mux.HandleFunc("/admin/levels/create", handlers.RequireAdmin(config.GetAdminEmails())(handlers.CreateLvlHandler))
	mux.HandleFunc("/admin/levels/", handlers.RequireAdmin(config.GetAdminEmails())(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/edit") {
			handlers.EditLevelFormHandler(w, r)
		} else if strings.HasSuffix(path, "/update") {
			levelID := strings.TrimSuffix(strings.TrimPrefix(path, "/admin/levels/"), "/update")
			handlers.UpdateLvlHandler(w, r, levelID)
		} else if strings.HasSuffix(path, "/delete") {
			levelID := strings.TrimSuffix(strings.TrimPrefix(path, "/admin/levels/"), "/delete")
			handlers.DeleteLvlHandler(w, r, levelID)
		} else {
			handlers.AdminHandler(w, r)
		}
	}))

	mux.HandleFunc("/admin/users/", handlers.RequireAdmin(config.GetAdminEmails())(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/delete") {
			userEmail := strings.TrimSuffix(strings.TrimPrefix(path, "/admin/users/"), "/delete")
			if r.Method == http.MethodPost {
				user, err := handlers.GetUserFromSession(r)
				if err != nil || user == nil {
					http.Error(w, "Access denied", http.StatusForbidden)
					return
				}
				if user.Gmail == userEmail {
					http.Redirect(w, r, "/admin?error=Cannot delete your own account", http.StatusSeeOther)
					return
				}
				err = database.Delete("login", map[string]interface{}{"gmail": userEmail})
				if err != nil {
					http.Redirect(w, r, "/admin?error=Failed to delete user", http.StatusSeeOther)
					return
				}
				http.Redirect(w, r, "/admin?success=User deleted successfully", http.StatusSeeOther)
				return
			}
		}
		// fallback
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	}))

	// Chat API endpoints
	mux.HandleFunc("/api/chat", handlers.ChatAPIHandler)
	mux.HandleFunc("/api/chat/leave", handlers.ChatLeaveHandler)

	// Authentication handlers
	mux.HandleFunc("/enter/New", handlers.New)
	mux.HandleFunc("/enter", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/enter/New", http.StatusPermanentRedirect)
	})
	mux.HandleFunc("/enter/verify", handlers.CORS(handlers.Verify))
	mux.HandleFunc("/enter/login", handlers.CORS(handlers.LoginF))
	mux.HandleFunc("/api/auth/logout", handlers.CORS(handlers.Logout))

	mux.HandleFunc("/enter/email", handlers.CORS(handlers.EmailOnly))
	mux.HandleFunc("/enter/email-verify", handlers.CORS(handlers.EmailVerify))

	// Legacy API endpoints (for backward compatibility)
	mux.HandleFunc("/api/question", handlers.GetQuestionHandler)
	mux.HandleFunc("/api/submit", handlers.SubmitAnswer)
	mux.HandleFunc("/dashboard", handlers.DashboardPage)

	// User session API
	mux.HandleFunc("/api/user/session", handlers.UserSessionHandler)
	mux.HandleFunc("/api/user/current-level", handlers.RequireAuth(handlers.GetCurrentLevelHandler))

	// Game API endpoints
	mux.HandleFunc("/api/submit-answer", handlers.RequireAuth(handlers.SubmitAnswerHandler))
	mux.HandleFunc("/api/notifications/unread-count", handlers.RequireAuth(handlers.GetNotificationCountHandler))

	// Admin API endpoints
	mux.HandleFunc("/api/admin/", func(w http.ResponseWriter, r *http.Request) {
		if !handlers.AdminAuth(w, r, config.GetAdminEmails()) {
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/api/admin")
		if path == "/" || path == "" {
			handlers.AdminPanelHandler(w, r)
			return
		}

		if path == "/stats" {
			handlers.GetStatsHandler(w, r)
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
			} else if levelPath == "/bulk-state" {
				if r.Method == "PATCH" {
					handlers.ToggleAllLevelsStateHandler(w, r)
				}
			} else {
				parts := strings.Split(strings.TrimPrefix(levelPath, "/"), "/")
				if len(parts) >= 1 {
					id := parts[0]
					if len(parts) >= 2 && parts[1] == "state" {
						if r.Method == "PATCH" {
							handlers.ToggleLevelStateHandler(w, r, id)
						}
					} else {
						if r.Method == "POST" {
							handlers.UpdateLvlHandler(w, r, id)
						} else if r.Method == "DELETE" {
							handlers.DeleteLvlHandler(w, r, id)
						}
					}
				}
			}
		}

		if strings.HasPrefix(path, "/users") {
			userPath := strings.TrimPrefix(path, "/users")
			if userPath == "" || userPath == "/" {
				if r.Method == "GET" {
					handlers.GetAllUsersHandler(w, r)
				}
			} else {
				email := strings.TrimPrefix(userPath, "/")
				if r.Method == "DELETE" {
					handlers.DeleteUserHandler(w, r, email)
				}
			}
		}
	})

	// API endpoints
	mux.HandleFunc("/api/secret", handlers.CORS(handlers.GetSecretHandler))

	// Static file serving
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./frontend/"))))
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./frontend/assets/"))))
	mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./frontend/css/"))))
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./frontend/js/"))))
	mux.HandleFunc("/styles.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/styles.css")
	})

	return &CustomHandler{mux: mux}
}
