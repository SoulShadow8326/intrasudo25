package routes

import (
	"intrasudo25/config"
	"intrasudo25/database"
	"intrasudo25/handlers"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

type CustomHandler = handlers.CustomHandler

func customFileServer(root string) http.Handler {
	fs := http.FileServer(http.Dir(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ext := filepath.Ext(r.URL.Path)
		switch ext {
		case ".css":
			w.Header().Set("Content-Type", "text/css")
		case ".js":
			w.Header().Set("Content-Type", "application/javascript")
		case ".png":
			w.Header().Set("Content-Type", "image/png")
		case ".jpg", ".jpeg":
			w.Header().Set("Content-Type", "image/jpeg")
		case ".ico":
			w.Header().Set("Content-Type", "image/x-icon")
		default:
			if mt := mime.TypeByExtension(ext); mt != "" {
				w.Header().Set("Content-Type", mt)
			}
		}
		fs.ServeHTTP(w, r)
	})
}

func RegisterRoutes() http.Handler {
	Mux := http.NewServeMux()

	// Time-gated routes (protected by countdown)
	Mux.HandleFunc("/landing", handlers.TimeGateMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/landing.html")
	}))
	Mux.HandleFunc("/playground", handlers.TimeGateMiddleware(handlers.RequireAuth(handlers.IndexHandler)))
	Mux.HandleFunc("/home", handlers.TimeGateMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/playground", http.StatusMovedPermanently)
	}))
	Mux.HandleFunc("/", handlers.TimeGateMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/landing", http.StatusSeeOther)
	}))
	Mux.HandleFunc("/auth", handlers.TimeGateMiddleware(handlers.AuthPageHandler))
	Mux.HandleFunc("/leaderboard", handlers.TimeGateMiddleware(handlers.RequireAuth(handlers.LeaderboardHandler)))
	Mux.HandleFunc("/announcements", handlers.TimeGateMiddleware(handlers.AnnouncementsHandler))
	Mux.HandleFunc("/hints", handlers.TimeGateMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/announcements", http.StatusMovedPermanently)
	}))
	Mux.HandleFunc("/guidelines", handlers.TimeGateMiddleware(handlers.GuidelinesHandler))

	// Status page (not time-gated)
	Mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		http.ServeFile(w, r, "./frontend/status.html")
	})

	// 404 page (not time-gated)
	Mux.HandleFunc("/404", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/404.html")
	})

	Mux.HandleFunc("/admin", handlers.RequireAdmin(config.GetAdminEmails())(handlers.AdminDashboardHandler))
	Mux.HandleFunc("/admin/levels/new", handlers.RequireAdmin(config.GetAdminEmails())(handlers.NewLevelFormHandler))
	Mux.HandleFunc("/submit", handlers.RequireAuth(handlers.SubmitAnswerFormHandler))
	Mux.HandleFunc("/admin/levels/create", handlers.RequireAdmin(config.GetAdminEmails())(handlers.CreateLvlHandler))
	Mux.HandleFunc("/admin/levels/", handlers.RequireAdmin(config.GetAdminEmails())(func(w http.ResponseWriter, r *http.Request) {
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

	Mux.HandleFunc("/admin/users/", handlers.RequireAdmin(config.GetAdminEmails())(func(w http.ResponseWriter, r *http.Request) {
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
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	}))

	Mux.HandleFunc("/enter/New", handlers.New)
	Mux.HandleFunc("/enter", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/enter/New", http.StatusPermanentRedirect)
	})
	Mux.HandleFunc("/enter/verify", handlers.CORS(handlers.Verify))
	Mux.HandleFunc("/enter/login", handlers.CORS(handlers.LoginF))
	Mux.HandleFunc("/api/auth/logout", handlers.CORS(handlers.Logout))

	Mux.HandleFunc("/enter/email", handlers.CORS(handlers.EmailOnly))
	Mux.HandleFunc("/enter/email-verify", handlers.CORS(handlers.EmailVerify))

	Mux.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"test": "success"}`))
	})
	Mux.HandleFunc("/api/countdown-status", handlers.CountdownStatusHandler)
	Mux.HandleFunc("/api/question", handlers.GetQuestionHandler)
	Mux.HandleFunc("/api/announcements", handlers.GetAnnouncementsForPublicHandler)
	Mux.HandleFunc("/api/submit", handlers.SubmitAnswer)
	Mux.HandleFunc("/dashboard", handlers.DashboardPage)

	Mux.HandleFunc("/api/user/session", handlers.UserSessionHandler)
	Mux.HandleFunc("/api/user/current-level", handlers.RequireAuth(handlers.GetCurrentLevelHandler))

	Mux.HandleFunc("/api/submit-answer", handlers.RequireAuth(handlers.SubmitAnswerHandler))
	Mux.HandleFunc("/api/notifications/unread-count", handlers.RequireAuth(handlers.GetNotificationCountHandler))
	Mux.HandleFunc("/api/leaderboard", handlers.RequireAuth(handlers.LeaderboardPage))

	Mux.HandleFunc("/api/admin/", func(w http.ResponseWriter, r *http.Request) {
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
			} else if userPath == "/reset-my-level" && r.Method == "POST" {
				handlers.ResetMyLevelHandler(w, r)
			} else {
				parts := strings.Split(strings.TrimPrefix(userPath, "/"), "/")
				if len(parts) >= 1 {
					email := parts[0]
					if len(parts) >= 2 && parts[1] == "reset-level" {
						if r.Method == "POST" {
							handlers.ResetUserLevelHandler(w, r, email)
						}
					} else if len(parts) >= 2 && parts[1] == "ban" {
						if r.Method == "POST" {
							handlers.BanUserEmailHandler(w, r, email)
						}
					} else if r.Method == "DELETE" {
						handlers.DeleteUserHandler(w, r, email)
					}
				}
			}
		}

		if strings.HasPrefix(path, "/announcements") {
			announcementPath := strings.TrimPrefix(path, "/announcements")
			if announcementPath == "" || announcementPath == "/" {
				if r.Method == "GET" {
					handlers.GetAllAnnouncementsHandler(w, r)
				} else if r.Method == "POST" {
					handlers.CreateAnnouncementHandler(w, r)
				}
			} else {
				parts := strings.Split(strings.TrimPrefix(announcementPath, "/"), "/")
				if len(parts) >= 1 {
					id := parts[0]
					if r.Method == "PUT" {
						handlers.UpdateAnnouncementHandler(w, r, id)
					} else if r.Method == "DELETE" {
						handlers.DeleteAnnouncementHandler(w, r, id)
					}
				}
			}
		}
	})

	Mux.HandleFunc("/api/secret", handlers.CORS(handlers.GetSecretHandler))

	// Discord bot specific routes (with Discord bot token authentication)
	Mux.HandleFunc("/api/discord/chat/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			handlers.GetChatStatusHandler(w, r)
		} else if r.Method == "POST" {
			handlers.ToggleChatStatusHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	Mux.HandleFunc("/api/discord/chat/level/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handlers.ToggleLevelChatStatusHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	Mux.HandleFunc("/api/discord-bot", handlers.DiscordBotHandler)
	Mux.HandleFunc("/api/levels", handlers.GetLevelsHandler)
	Mux.HandleFunc("/api/chat/checksum", handlers.RequireAuth(handlers.ChatChecksumHandler))
	Mux.HandleFunc("/api/check-messages", handlers.RequireAuth(handlers.CheckMessagesHandler))
	Mux.HandleFunc("/api/leads", handlers.RequireAuth(handlers.LeadsHandler))
	Mux.HandleFunc("/submit_message", handlers.RequireAuth(handlers.SubmitMessageHandler))

	Mux.Handle("/static/", http.StripPrefix("/static/", customFileServer("./frontend/")))
	Mux.Handle("/assets/", http.StripPrefix("/assets/", customFileServer("./frontend/assets/")))
	Mux.Handle("/css/", http.StripPrefix("/css/", customFileServer("./frontend/css/")))
	Mux.Handle("/js/", http.StripPrefix("/js/", customFileServer("./frontend/js/")))
	Mux.HandleFunc("/styles.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, "./frontend/styles.css")
	})

	ret_h := &CustomHandler{Mux: Mux}
	return handlers.CheckHeadersMiddleware(ret_h)
}
