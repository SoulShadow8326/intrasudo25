package handlers

import (
	"encoding/json"
	"intrasudo25/database"
	"net/http"
	"strings"
)

func checkIfAdmin(userEmail string, adminEmails []string) bool {
	for _, adminEmail := range adminEmails {
		if strings.EqualFold(userEmail, adminEmail) {
			return true
		}
	}
	return false
}

func Authorize(r *http.Request) (bool, *database.Login) {
	cookie, err := r.Cookie("exun_sesh_cookie")
	if err != nil || cookie.Value == "" {
		return false, nil
	}
	result, err := database.Get("login", map[string]interface{}{"cookie": cookie.Value})
	if err != nil || result == nil {
		return false, nil
	}
	acc := result.(*database.Login)
	csrf := r.Header.Get("CSRFtok")

	if csrf == "" || csrf != acc.CSRFtok {
		return false, nil
	}

	return true, acc
}

func AdminAuth(w http.ResponseWriter, r *http.Request, users []string) bool {
	isAuth, user := Authorize(r)

	if !isAuth || user == nil {
		// Check if this is an API request
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "Forbidden"})
			return false
		}
		// Redirect to appropriate error page for HTML requests
		UnauthorizedHandler(w, r)
		return false
	}

	allowed := checkIfAdmin(user.Gmail, users)

	if !allowed {
		// Check if this is an API request
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "Forbidden"})
			return false
		}
		// Redirect to admin required error page for HTML requests
		AdminRequiredHandler(w, r)
		return false
	}

	return true
}

// RequireAuth middleware to check if user is authenticated
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAuth, user := Authorize(r)
		if !isAuth || user == nil {
			// Check if this is an API request
			if strings.HasPrefix(r.URL.Path, "/api/") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
				return
			}
			// Redirect to auth page for HTML requests
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAdmin middleware to check if user is admin
func RequireAdmin(adminEmails []string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isAuth, user := Authorize(r)
			if !isAuth || user == nil {
				if strings.HasPrefix(r.URL.Path, "/api/") {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
					return
				}
				UnauthorizedHandler(w, r)
				return
			}

			isAdmin := checkIfAdmin(user.Gmail, adminEmails)

			if !isAdmin {
				if strings.HasPrefix(r.URL.Path, "/api/") {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(map[string]string{"error": "Admin access required"})
					return
				}
				AdminRequiredHandler(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromSession extracts user from session cookie
func GetUserFromSession(r *http.Request) (*database.Login, error) {
	cookie, err := r.Cookie("exun_sesh_cookie")
	if err != nil {
		return nil, err
	}

	result, err := database.Get("login", map[string]interface{}{"cookie": cookie.Value})
	if err != nil || result == nil {
		return nil, err
	}
	user := result.(*database.Login)

	return user, nil
}

// UserSessionHandler provides user session information for frontend
func UserSessionHandler(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "No active session"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"userId":   user.Gmail,
		"email":    user.Gmail,
		"level":    user.On,
		"verified": user.Verified,
	})
}
