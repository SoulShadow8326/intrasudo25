package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	serveErrorPage(w, r, http.StatusNotFound, "not_found", "")
}

func AccessDeniedHandler(w http.ResponseWriter, r *http.Request) {
	serveErrorPage(w, r, http.StatusForbidden, "access_denied", "")
}

func UnauthorizedHandler(w http.ResponseWriter, r *http.Request) {
	serveErrorPage(w, r, http.StatusUnauthorized, "unauthorized", "")
}

func AdminRequiredHandler(w http.ResponseWriter, r *http.Request) {
	serveErrorPage(w, r, http.StatusForbidden, "admin_required", "")
}

func LevelNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	serveErrorPage(w, r, http.StatusNotFound, "level_not_found", "")
}

func SessionExpiredHandler(w http.ResponseWriter, r *http.Request) {
	serveErrorPage(w, r, http.StatusUnauthorized, "session_expired", "")
}

func serveErrorPage(w http.ResponseWriter, r *http.Request, statusCode int, errorType string, customMessage string) {
	content, err := os.ReadFile(filepath.Join("frontend", "404.html"))
	if err != nil {
		w.WriteHeader(statusCode)
		w.Write([]byte(fmt.Sprintf("%d - Error occurred", statusCode)))
		return
	}

	var redirectURL string
	if errorType != "" {
		redirectURL = fmt.Sprintf("/404?type=%s&code=%d", errorType, statusCode)
		if customMessage != "" {
			redirectURL += fmt.Sprintf("&message=%s", customMessage)
		}
	} else {
		redirectURL = "/404"
	}

	if r.URL.Path == "/404" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(statusCode)
		w.Write(content)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
