package handlers

import (
	"encoding/json"
	"fmt"
	"intrasudo25/database"
	"net/http"
	"strings"
)

func checkIfAdmin(userEmail string, _ []string) bool {
	for _, adminEmail := range AdminEmails {
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
	result, err := database.Get("login", map[string]interface{}{"seshTok": cookie.Value})
	if err != nil || result == nil {
		return false, nil
	}
	acc := result.(*database.Login)

	if acc.SeshTok == "" || acc.SeshTok != cookie.Value {
		return false, nil
	}

	if r.Method != "GET" {
		csrf := r.Header.Get("CSRFtok")
		if csrf == "" || csrf != acc.CSRFtok {
			return false, nil
		}
	}

	return true, acc
}

func AdminAuth(w http.ResponseWriter, r *http.Request, _ []string) bool {
	isAuth, user := Authorize(r)

	if !isAuth || user == nil {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "Forbidden"})
			return false
		}
		UnauthorizedHandler(w, r)
		return false
	}
	allowed := checkIfAdmin(user.Gmail, nil)
	if !allowed {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "Forbidden"})
			return false
		}
		AdminRequiredHandler(w, r)
		return false
	}
	return true
}

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAuth, user := Authorize(r)
		if !isAuth || user == nil {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
				return
			}
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

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

func GetUserFromSession(r *http.Request) (*database.Login, error) {
	cookie, err := r.Cookie("exun_sesh_cookie")
	if err != nil || cookie.Value == "" {
		return nil, err
	}
	result, err := database.Get("login", map[string]interface{}{"seshTok": cookie.Value})
	if err != nil || result == nil {
		return nil, err
	}
	acc := result.(*database.Login)
	if acc.SeshTok == "" || acc.SeshTok != cookie.Value {
		return nil, fmt.Errorf("invalid session")
	}
	return acc, nil
}

func UserSessionHandler(w http.ResponseWriter, r *http.Request) {
	isAuth, user := Authorize(r)
	if !isAuth || user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"isAdmin": false})
		return
	}
	isAdmin := false
	for _, adminEmail := range AdminEmails {
		if strings.EqualFold(user.Gmail, adminEmail) {
			isAdmin = true
			break
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"userId":  user.Gmail,
		"isAdmin": isAdmin,
	})
}
