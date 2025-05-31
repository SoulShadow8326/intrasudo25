package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"intrasudo25/config"
	"intrasudo25/database"
	"net/http"
	"os"
	"strings"
)

type CustomHandler struct {
	Mux *http.ServeMux
}
func (h *CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, pattern := h.Mux.Handler(r)
	if pattern == "" {
		NotFoundHandler(w, r)
		return
	}

	h.Mux.ServeHTTP(w, r)
}

func checkIfAdmin(userEmail string, adminEmails []string) bool {
	emailsToCheck := adminEmails
	if emailsToCheck == nil {
		emailsToCheck = config.GetAdminEmails()
	}

	for _, adminEmail := range emailsToCheck {
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

func AdminAuth(w http.ResponseWriter, r *http.Request, adminEmails []string) bool {
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
	allowed := checkIfAdmin(user.Gmail, adminEmails)
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
	for _, adminEmail := range config.GetAdminEmails() {
		if strings.EqualFold(user.Gmail, adminEmail) {
			isAdmin = true
			break
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"userId":    user.Gmail,
		"isAdmin":   isAdmin,
		"csrfToken": user.CSRFtok,
	})
}

func CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-secret, X-CSRF-Token, Accept")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func HasXForwardedFor(r *http.Request) bool {
	return r.Header.Get("X-Forwarded-For") != ""
}
func IsSecretValid(r *http.Request) bool {
    xSecret := r.Header.Get("X-secret")
    if xSecret == "" {
        fmt.Println("NO SECRET?")
        return false
    }
    secret := os.Getenv("salt")
    method := r.Method

    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(method))
    expectedMAC := mac.Sum(nil)
    expectedB64 := base64.StdEncoding.EncodeToString(expectedMAC)

    fmt.Println("Expected:", expectedB64)
    fmt.Println("Actual:  ", xSecret)

    return hmac.Equal([]byte(expectedB64), []byte(xSecret))
}

func CheckHeadersMiddleware(Next *CustomHandler) http.Handler {
	next := Next.Mux;
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !HasXForwardedFor(r) {
			http.Error(w, "Missing X-Forwarded-For header", http.StatusForbidden)
			return
		}
		if !IsSecretValid(r) {
			http.Error(w, "Invalid or missing X-Secret header", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
