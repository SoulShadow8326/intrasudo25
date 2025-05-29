package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
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
	// Read the 404.html template
	content, err := ioutil.ReadFile(filepath.Join("frontend", "404.html"))
	if err != nil {
		// Fallback to plain text if template is missing
		w.WriteHeader(statusCode)
		w.Write([]byte(fmt.Sprintf("%d - Error occurred", statusCode)))
		return
	}

	// Build the redirect URL with error parameters
	var redirectURL string
	if errorType != "" {
		redirectURL = fmt.Sprintf("/404?type=%s&code=%d", errorType, statusCode)
		if customMessage != "" {
			redirectURL += fmt.Sprintf("&message=%s", customMessage)
		}
	} else {
		redirectURL = "/404"
	}

	// Check if this is already a request to /404 to avoid redirect loops
	if r.URL.Path == "/404" {
		// Serve the 404 page directly
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(statusCode)
		w.Write(content)
		return
	}

	// Redirect to the 404 page with parameters
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
