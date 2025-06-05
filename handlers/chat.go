package handlers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"

	"intrasudo25/database"
)

type CheckMessagesRequest struct {
	Checksum string `json:"checksum"`
}

type CheckMessagesResponse struct {
	HasUpdate bool                   `json:"hasUpdate"`
	Messages  map[string]interface{} `json:"messages,omitempty"`
}

func CheckMessagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CheckMessagesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	chatMessages, err := database.Get("chat_messages", map[string]interface{}{"limit": 50})
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	messages := map[string]interface{}{
		"chat_messages": chatMessages,
	}

	messagesJSON, _ := json.Marshal(messages)
	currentChecksum := fmt.Sprintf("%x", md5.Sum(messagesJSON))

	response := CheckMessagesResponse{
		HasUpdate: currentChecksum != req.Checksum,
	}

	if response.HasUpdate {
		response.Messages = messages
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
