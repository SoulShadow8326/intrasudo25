package handlers

import (
	"encoding/json"
	"intrasudo25/database"
	"net/http"
	"strings"
)

// ChatHandler serves the chat page
func ChatHandler(w http.ResponseWriter, r *http.Request) {
	// This is handled by ChatPageHandler in pages.go for serving HTML
	http.Redirect(w, r, "/chat", http.StatusSeeOther)
}

// ChatAPIHandler handles chat API endpoints
func ChatAPIHandler(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Update participant status
	adminEmails := []string{"admin@intrasudo.com", "lead@intrasudo.com", "organizer@intrasudo.com"}
	isAdmin := false
	email := strings.ToLower(user.Gmail)
	for _, adminEmail := range adminEmails {
		if email == adminEmail {
			isAdmin = true
			break
		}
	}
	database.Update("chat_participant", map[string]interface{}{}, map[string]interface{}{
		"email":    user.Gmail,
		"isOnline": true,
		"isAdmin":  isAdmin,
	})

	switch r.Method {
	case "GET":
		handleGetChatData(w)
	case "POST":
		handleSendMessage(w, r, user)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetChatData(w http.ResponseWriter) {
	result, err := database.Get("chat_messages", map[string]interface{}{"limit": 50})
	var messages []database.ChatMessage
	if err != nil {
		messages = []database.ChatMessage{}
	} else {
		messages = result.([]database.ChatMessage)
	}

	result, err = database.Get("chat_participants", map[string]interface{}{})
	var participants []database.ChatParticipant
	if err != nil {
		participants = []database.ChatParticipant{}
	} else {
		participants = result.([]database.ChatParticipant)
	}

	response := map[string]interface{}{
		"messages":     messages,
		"participants": participants,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleSendMessage(w http.ResponseWriter, r *http.Request, user *database.Login) {
	r.ParseForm()
	message := strings.TrimSpace(r.FormValue("message"))

	if message == "" {
		http.Error(w, "Message cannot be empty", http.StatusBadRequest)
		return
	}

	// Check if user is admin
	adminEmails := []string{"admin@intrasudo.com", "lead@intrasudo.com", "organizer@intrasudo.com"}
	isAdmin := false
	email := strings.ToLower(user.Gmail)
	for _, adminEmail := range adminEmails {
		if email == adminEmail {
			isAdmin = true
			break
		}
	}

	err := database.Create("chat_message", map[string]interface{}{
		"email":   user.Gmail,
		"message": message,
		"isAdmin": isAdmin,
	})
	if err != nil {
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// ChatLeaveHandler handles when users leave chat
func ChatLeaveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := GetUserFromSession(r)
	if err != nil {
		return // User not logged in, nothing to do
	}

	// Check if user is admin
	adminEmails := []string{"admin@intrasudo.com", "lead@intrasudo.com", "organizer@intrasudo.com"}
	isAdmin := false
	email := strings.ToLower(user.Gmail)
	for _, adminEmail := range adminEmails {
		if email == adminEmail {
			isAdmin = true
			break
		}
	}

	database.Update("chat_participant", map[string]interface{}{}, map[string]interface{}{
		"email":    user.Gmail,
		"isOnline": false,
		"isAdmin":  isAdmin,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
