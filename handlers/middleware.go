package handlers

import (
	"encoding/json"
	"intrasudo25/database"
	"net/http"
)

func Authorize(r *http.Request) (bool, Login) {
	cookie, err := r.Cookie("exun_sesh_cookie")
	if err != nil || cookie.Value == "" {
		return false, Login{}
	}
	acc, err := database.GetLoginFromCookie(cookie.Value)
	if err != nil {
		return false, Login{}
	}
	csrf := r.Header.Get("CSRFtok")

	if csrf == "" || csrf != acc.CSRFtok {
		return false, Login{}
	}

	return true, *acc
}

func AdminAuth(w http.ResponseWriter, r *http.Request, users []string) bool {
	isAuth, user := Authorize(r)

	if !isAuth {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Forbidden"})
		return false
	}

	allowed := false
	for _, u := range users {
		if u == user.Gmail {
			allowed = true
			break
		}
	}

	if !allowed {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Forbidden"})
		return false
	}

	return true
}
