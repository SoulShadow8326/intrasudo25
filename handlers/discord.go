package handlers

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"intrasudo25/config"
	"intrasudo25/database"
	"net/http"
	"sort"
	"strings"
)

type DiscordMessage struct {
	DiscordMsgID string `json:"discordMsgId"`
	UserEmail    string `json:"userEmail,omitempty"`
	Username     string `json:"username,omitempty"`
	Message      string `json:"message"`
	IsReply      bool   `json:"isReply,omitempty"`
	ParentMsgID  string `json:"parentMsgId,omitempty"`
	SentBy       string `json:"sentBy,omitempty"`
	Timestamp    string `json:"timestamp"`
}

type DiscordBotRequest struct {
	Type         string           `json:"type"`
	UserEmail    string           `json:"userEmail"`
	Username     string           `json:"username"`
	Message      string           `json:"message"`
	LevelNumber  int              `json:"levelNumber"`
	DiscordMsgID string           `json:"discordMsgId"`
	ParentMsgID  int              `json:"parentMsgId"`
	SentBy       string           `json:"sentBy"`
	MessageType  string           `json:"messageType,omitempty"`
	Messages     []DiscordMessage `json:"messages,omitempty"`
}

type DiscordBotResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ChatChecksumRequest struct {
	LeadsHash string `json:"leadsHash"`
	HintsHash string `json:"hintsHash"`
}

type ChatChecksumResponse struct {
	Changed   bool                   `json:"changed"`
	Leads     []database.LeadMessage `json:"leads,omitempty"`
	Hints     []database.HintMessage `json:"hints,omitempty"`
	LeadsHash string                 `json:"leadsHash"`
	HintsHash string                 `json:"hintsHash"`
}

func DiscordBotHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	botToken := r.Header.Get("Authorization")
	if botToken != "Bearer "+config.GetDiscordBotToken() {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req DiscordBotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch req.Type {
	case "lead_message":
		handleLeadMessage(w, req)
	case "lead_reply":
		handleLeadReply(w, req)
	case "hint_message":
		handleHintMessage(w, req)
	case "update_discord_msg_id":
		handleUpdateDiscordMsgID(w, req)
	case "lookup_discord_msg":
		handleLookupDiscordMsg(w, req)
	case "sync_messages":
		handleSyncMessages(w, req)
	default:
		json.NewEncoder(w).Encode(DiscordBotResponse{
			Success: false,
			Message: "Invalid request type",
		})
	}
}

func handleLeadMessage(w http.ResponseWriter, req DiscordBotRequest) {
	leadMsg := database.LeadMessage{
		UserEmail:    req.UserEmail,
		Username:     req.Username,
		Message:      req.Message,
		LevelNumber:  req.LevelNumber,
		DiscordMsgID: req.DiscordMsgID,
		IsReply:      false,
	}

	err := database.Create("lead_message", leadMsg)
	if err != nil {
		json.NewEncoder(w).Encode(DiscordBotResponse{
			Success: false,
			Message: "Failed to save lead message",
		})
		return
	}

	json.NewEncoder(w).Encode(DiscordBotResponse{
		Success: true,
		Message: "Lead message saved",
	})
}

func handleLeadReply(w http.ResponseWriter, req DiscordBotRequest) {
	leadMsg := database.LeadMessage{
		UserEmail:    req.UserEmail,
		Username:     req.SentBy,
		Message:      req.Message,
		LevelNumber:  req.LevelNumber,
		DiscordMsgID: req.DiscordMsgID,
		IsReply:      true,
		ParentMsgID:  req.ParentMsgID,
	}

	err := database.Create("lead_message", leadMsg)
	if err != nil {
		json.NewEncoder(w).Encode(DiscordBotResponse{
			Success: false,
			Message: "Failed to save lead reply",
		})
		return
	}

	json.NewEncoder(w).Encode(DiscordBotResponse{
		Success: true,
		Message: "Lead reply saved",
	})
}

func handleHintMessage(w http.ResponseWriter, req DiscordBotRequest) {
	hintMsg := database.HintMessage{
		Message:      req.Message,
		LevelNumber:  req.LevelNumber,
		DiscordMsgID: req.DiscordMsgID,
		SentBy:       req.SentBy,
	}

	err := database.Create("hint_message", hintMsg)
	if err != nil {
		json.NewEncoder(w).Encode(DiscordBotResponse{
			Success: false,
			Message: "Failed to save hint message",
		})
		return
	}

	usersAtLevel, err := database.Get("users_at_level", map[string]interface{}{
		"level": req.LevelNumber,
	})
	if err != nil {
		json.NewEncoder(w).Encode(DiscordBotResponse{
			Success: false,
			Message: "Failed to get users at level",
		})
		return
	}

	if users, ok := usersAtLevel.([]string); ok {
		for _, userEmail := range users {
			database.Create("notification", map[string]interface{}{
				"userEmail": userEmail,
				"message":   fmt.Sprintf("New hint for level %d: %s", req.LevelNumber, req.Message),
				"type":      "hint",
			})
		}
	}

	json.NewEncoder(w).Encode(DiscordBotResponse{
		Success: true,
		Message: "Hint message saved and notifications sent",
	})
}

func ChatChecksumHandler(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatChecksumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	currentLeadsHash := calculateLeadsHash(user.Gmail)
	currentHintsHash := calculateHintsHash(user.Gmail)

	if req.LeadsHash == currentLeadsHash && req.HintsHash == currentHintsHash {
		json.NewEncoder(w).Encode(ChatChecksumResponse{
			Changed:   false,
			LeadsHash: currentLeadsHash,
			HintsHash: currentHintsHash,
		})
		return
	}

	leads := []database.LeadMessage{}
	hints := []database.HintMessage{}

	userLevel, err := database.Get("user_level", map[string]interface{}{"email": user.Gmail})
	if err == nil {
		if level, ok := userLevel.(int); ok {
			result, err := database.Get("lead_messages", map[string]interface{}{
				"userEmail": user.Gmail,
				"level":     level,
			})
			if err == nil {
				if leadMsgs, ok := result.([]database.LeadMessage); ok {
					leads = leadMsgs
				}
			}

			result, err = database.Get("hint_messages", map[string]interface{}{
				"level": level,
			})
			if err == nil {
				if hintMsgs, ok := result.([]database.HintMessage); ok {
					hints = hintMsgs
				}
			}
		}
	}

	json.NewEncoder(w).Encode(ChatChecksumResponse{
		Changed:   true,
		Leads:     leads,
		Hints:     hints,
		LeadsHash: currentLeadsHash,
		HintsHash: currentHintsHash,
	})
}

func calculateHintsHash(userEmail string) string {
	userLevel, err := database.Get("user_level", map[string]interface{}{"email": userEmail})
	if err != nil {
		return ""
	}

	if level, ok := userLevel.(int); ok {
		result, err := database.Get("hint_messages", map[string]interface{}{
			"level": level,
		})
		if err != nil {
			return ""
		}

		if hints, ok := result.([]database.HintMessage); ok {
			var hintData []string
			for _, hint := range hints {
				hintData = append(hintData, fmt.Sprintf("%d:%d:%s", hint.ID, hint.LevelNumber, hint.Message))
			}
			sort.Strings(hintData)
			combined := strings.Join(hintData, "|")
			hash := md5.Sum([]byte(combined))
			return hex.EncodeToString(hash[:])
		}
	}
	return ""
}

func calculateLeadsHash(userEmail string) string {
	userLevel, err := database.Get("user_level", map[string]interface{}{"email": userEmail})
	if err != nil {
		return ""
	}

	if level, ok := userLevel.(int); ok {
		result, err := database.Get("lead_messages", map[string]interface{}{
			"userEmail": userEmail,
			"level":     level,
		})
		if err != nil {
			return ""
		}

		if leads, ok := result.([]database.LeadMessage); ok {
			var leadData []string
			for _, lead := range leads {
				leadData = append(leadData, fmt.Sprintf("%d:%s:%s", lead.ID, lead.UserEmail, lead.Message))
			}
			sort.Strings(leadData)
			combined := strings.Join(leadData, "|")
			hash := md5.Sum([]byte(combined))
			return hex.EncodeToString(hash[:])
		}
	}
	return ""
}

func LeadsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case "GET":
		handleGetLeads(w, user)
	case "POST":
		handleSendLead(w, r, user)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetLeads(w http.ResponseWriter, user *database.Login) {
	userLevel, err := database.Get("user_level", map[string]interface{}{"email": user.Gmail})
	if err != nil {
		http.Error(w, "Failed to get user level", http.StatusInternalServerError)
		return
	}

	if level, ok := userLevel.(int); ok {
		result, err := database.Get("lead_messages", map[string]interface{}{
			"userEmail": user.Gmail,
			"level":     level,
		})
		if err != nil {
			http.Error(w, "Failed to get lead messages", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	} else {
		http.Error(w, "Invalid user level", http.StatusInternalServerError)
	}
}

func handleSendLead(w http.ResponseWriter, r *http.Request, user *database.Login) {
	var req struct {
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Message) == "" {
		http.Error(w, "Message cannot be empty", http.StatusBadRequest)
		return
	}

	userLevel, err := database.Get("user_level", map[string]interface{}{"email": user.Gmail})
	if err != nil {
		http.Error(w, "Failed to get user level", http.StatusInternalServerError)
		return
	}

	if level, ok := userLevel.(int); ok {
		leadMsg := database.LeadMessage{
			UserEmail:   user.Gmail,
			Username:    user.Name,
			Message:     req.Message,
			LevelNumber: level,
			IsReply:     false,
		}

		err := database.Create("lead_message", leadMsg)
		if err != nil {
			http.Error(w, "Failed to save lead message", http.StatusInternalServerError)
			return
		}

		err = forwardToDiscord(user.Gmail, req.Message, level)
		if err != nil {
			fmt.Printf("Failed to forward message to Discord: %v\n", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Lead message sent",
		})
	} else {
		http.Error(w, "Invalid user level", http.StatusInternalServerError)
	}
}

func SubmitMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Message) == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Message cannot be empty",
		})
		return
	}

	userLevel, err := database.Get("user_level", map[string]interface{}{"email": user.Gmail})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to get user level",
		})
		return
	}

	if level, ok := userLevel.(int); ok {
		leadMsg := database.LeadMessage{
			UserEmail:   user.Gmail,
			Username:    user.Name,
			Message:     req.Message,
			LevelNumber: level,
			IsReply:     false,
		}

		err := database.Create("lead_message", leadMsg)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Failed to save message",
			})
			return
		}

		err = forwardToDiscord(user.Gmail, req.Message, level)
		if err != nil {
			fmt.Printf("Failed to forward message to Discord: %v\n", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Message sent successfully",
		})
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid user level",
		})
	}
}

type DiscordWebhookPayload struct {
	Content string `json:"content"`
}

type DiscordMessageRequest struct {
	UserEmail string `json:"userEmail"`
	Username  string `json:"username"`
	Message   string `json:"message"`
	Level     int    `json:"level"`
}

func forwardToDiscord(userEmail, message string, level int) error {
	req := DiscordMessageRequest{
		UserEmail: userEmail,
		Username:  getUsernameFromEmail(userEmail),
		Message:   message,
		Level:     level,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	botURL := config.GetDiscordBotURL()
	if !strings.HasPrefix(botURL, "http") {
		botURL = "http://" + botURL
	}

	fmt.Printf("Forwarding message to Discord at %s/discord/forward\n", botURL)
	resp, err := http.Post(
		fmt.Sprintf("%s/discord/forward", botURL),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discord bot returned status %d", resp.StatusCode)
	}

	// Parse response to get Discord message ID
	var result struct {
		Success      bool   `json:"success"`
		Message      string `json:"message"`
		DiscordMsgID string `json:"discordMsgId"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// If we got a Discord message ID, update it in our database
	if result.Success && result.DiscordMsgID != "" {
		// Find the latest lead message for this user at this level
		findResult, err := database.Get("lead_messages", map[string]interface{}{
			"userEmail": userEmail,
			"level":     level,
		})
		if err != nil {
			return fmt.Errorf("failed to get lead messages: %v", err)
		}

		if leadMessages, ok := findResult.([]database.LeadMessage); ok && len(leadMessages) > 0 {
			// Get the last message (most recent)
			lastMsg := leadMessages[len(leadMessages)-1]

			// Update the DiscordMsgID
			err = database.Update("lead_message", map[string]interface{}{
				"id": lastMsg.ID,
			}, map[string]interface{}{
				"discordMsgID": result.DiscordMsgID,
			})

			if err != nil {
				return fmt.Errorf("failed to update Discord message ID: %v", err)
			}

			// Store a mapping for quicker lookup later
			database.Create("message_mapping", map[string]interface{}{"userEmail": userEmail,
				"dbMessageId":  lastMsg.ID,
				"discordMsgId": result.DiscordMsgID,
				"levelNumber":  level,
				"timestamp":    lastMsg.Timestamp,
			})
		}
	}

	return nil
}

func getUsernameFromEmail(email string) string {
	result, err := database.Get("user_by_email", map[string]interface{}{"email": email})
	if err != nil {
		return email
	}
	if user, ok := result.(*database.Login); ok {
		return user.Name
	}
	return email
}

func GetLevelsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	botToken := r.Header.Get("Authorization")
	if botToken != "Bearer "+config.GetDiscordBotToken() {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	result, err := database.Get("level", map[string]interface{}{"all": true})
	if err != nil {
		http.Error(w, "Failed to get levels", http.StatusInternalServerError)
		return
	}

	levels := []int{}
	if levelData, ok := result.([]database.AdminLevel); ok {
		for _, level := range levelData {
			levels = append(levels, level.LevelNumber)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"levels": levels,
	})
}

func RefreshDiscordChannels() error {
	resp, err := http.Post(
		fmt.Sprintf("%s/discord/refresh", config.GetDiscordBotURL()),
		"application/json",
		bytes.NewBuffer([]byte("{}")),
	)
	if err != nil {
		return fmt.Errorf("failed to refresh discord channels: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discord bot returned status %d", resp.StatusCode)
	}

	return nil
}

func GetUserHintsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userLevel, err := database.Get("user_level", map[string]interface{}{"email": user.Gmail})
	if err != nil {
		http.Error(w, "Failed to get user level", http.StatusInternalServerError)
		return
	}

	if level, ok := userLevel.(int); ok {
		result, err := database.Get("hint_messages", map[string]interface{}{
			"level": level,
		})
		if err != nil {
			http.Error(w, "Failed to get hint messages", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	} else {
		http.Error(w, "Invalid user level", http.StatusInternalServerError)
	}
}

func handleUpdateDiscordMsgID(w http.ResponseWriter, req DiscordBotRequest) {
	fmt.Printf("Received update_discord_msg_id request for user: %s, level: %d, message: %s, discordMsgId: %s\n",
		req.UserEmail, req.LevelNumber, req.Message, req.DiscordMsgID)

	// First try to find the message by content (exact match on first 50 chars)
	result, err := database.Get("lead_messages_by_content", map[string]interface{}{
		"userEmail": req.UserEmail,
		"level":     req.LevelNumber,
		"content":   req.Message,
	})

	if err != nil || result == nil {
		// If exact content match fails, try to find the most recent message from this user at this level
		result, err = database.Get("lead_messages", map[string]interface{}{
			"userEmail": req.UserEmail,
			"level":     req.LevelNumber,
			"limit":     1,
			"orderBy":   "created_at DESC",
		})

		if err != nil {
			json.NewEncoder(w).Encode(DiscordBotResponse{
				Success: false,
				Message: "Failed to find lead message",
			})
			return
		}
	}

	if leadMsgs, ok := result.([]database.LeadMessage); ok && len(leadMsgs) > 0 {
		// Update the DiscordMsgID for the most recent matching message
		latestMsg := leadMsgs[len(leadMsgs)-1]

		fmt.Printf("Found lead message with ID %d for user %s\n", latestMsg.ID, req.UserEmail)

		err := database.Update("lead_message", map[string]interface{}{
			"id": latestMsg.ID,
		}, map[string]interface{}{
			"discordMsgID": req.DiscordMsgID,
		})

		if err != nil {
			json.NewEncoder(w).Encode(DiscordBotResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to update Discord message ID: %v", err),
			})
			return
		}

		// Store a mapping for quicker lookup later
		database.Create("message_mapping", map[string]interface{}{
			"userEmail":    req.UserEmail,
			"dbMessageId":  latestMsg.ID,
			"discordMsgId": req.DiscordMsgID,
			"levelNumber":  req.LevelNumber,
			"timestamp":    latestMsg.Timestamp,
		})

		json.NewEncoder(w).Encode(DiscordBotResponse{
			Success: true,
			Message: "Discord message ID updated",
			Data: map[string]interface{}{
				"id": latestMsg.ID,
			},
		})
		return
	}

	json.NewEncoder(w).Encode(DiscordBotResponse{
		Success: false,
		Message: "No matching lead message found",
	})
}

func handleLookupDiscordMsg(w http.ResponseWriter, req DiscordBotRequest) {
	if req.DiscordMsgID == "" {
		json.NewEncoder(w).Encode(DiscordBotResponse{
			Success: false,
			Message: "Missing Discord message ID",
		})
		return
	}

	// Try to find the message in the database
	result, err := database.Get("message_by_discord_id", map[string]interface{}{
		"discordMsgId": req.DiscordMsgID,
	})

	if err != nil {
		json.NewEncoder(w).Encode(DiscordBotResponse{
			Success: false,
			Message: fmt.Sprintf("Error looking up message: %v", err),
		})
		return
	}

	if msg, ok := result.(database.LeadMessage); ok {
		json.NewEncoder(w).Encode(DiscordBotResponse{
			Success: true,
			Message: "Message found",
			Data: map[string]interface{}{
				"dbMessageId": msg.ID,
				"userEmail":   msg.UserEmail,
				"levelNumber": msg.LevelNumber,
			},
		})
		return
	}

	// If we couldn't find it in the LeadMessage table directly, try the message_mapping table
	mappingResult, err := database.Get("message_mapping", map[string]interface{}{
		"discordMsgId": req.DiscordMsgID,
	})

	if err != nil {
		json.NewEncoder(w).Encode(DiscordBotResponse{
			Success: false,
			Message: "Message not found",
		})
		return
	}

	if mapping, ok := mappingResult.(map[string]interface{}); ok {
		json.NewEncoder(w).Encode(DiscordBotResponse{
			Success: true,
			Message: "Message mapping found",
			Data: map[string]interface{}{
				"dbMessageId": mapping["dbMessageId"],
				"userEmail":   mapping["userEmail"],
				"levelNumber": mapping["levelNumber"],
			},
		})
		return
	}

	json.NewEncoder(w).Encode(DiscordBotResponse{
		Success: false,
		Message: "No mapping found for Discord message",
	})
}

func handleSyncMessages(w http.ResponseWriter, req DiscordBotRequest) {
	if req.MessageType == "" {
		json.NewEncoder(w).Encode(DiscordBotResponse{
			Success: false,
			Message: "Missing message type",
		})
		return
	}

	if req.LevelNumber <= 0 {
		json.NewEncoder(w).Encode(DiscordBotResponse{
			Success: false,
			Message: "Invalid level number",
		})
		return
	}

	if req.Messages == nil {
		json.NewEncoder(w).Encode(DiscordBotResponse{
			Success: false,
			Message: "No messages provided",
		})
		return
	}

	var successCount int
	var errorCount int

	if req.MessageType == "lead" {
		err := database.DeleteAllMessagesForLevel(req.LevelNumber, "lead")
		if err != nil {
			json.NewEncoder(w).Encode(DiscordBotResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to delete existing lead messages: %v", err),
			})
			return
		}

		for _, msg := range req.Messages {
			leadMsg := database.LeadMessage{
				UserEmail:    msg.UserEmail,
				Username:     msg.Username,
				Message:      msg.Message,
				LevelNumber:  req.LevelNumber,
				DiscordMsgID: msg.DiscordMsgID,
				IsReply:      msg.IsReply,
			}

			if msg.IsReply && msg.ParentMsgID != "" {
				parentResult, err := database.Get("lead_message_by_discord_id", map[string]interface{}{
					"discordMsgId": msg.ParentMsgID,
				})
				if err == nil {
					if parentMsg, ok := parentResult.(database.LeadMessage); ok {
						leadMsg.ParentMsgID = parentMsg.ID
					}
				}
			}

			err := database.Create("lead_message", leadMsg)
			if err != nil {
				errorCount++
			} else {
				successCount++
			}
		}
	} else if req.MessageType == "hint" {
		err := database.DeleteAllMessagesForLevel(req.LevelNumber, "hint")
		if err != nil {
			json.NewEncoder(w).Encode(DiscordBotResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to delete existing hint messages: %v", err),
			})
			return
		}

		for _, msg := range req.Messages {
			hintMsg := database.HintMessage{
				Message:      msg.Message,
				LevelNumber:  req.LevelNumber,
				DiscordMsgID: msg.DiscordMsgID,
				SentBy:       msg.SentBy,
			}

			err := database.Create("hint_message", hintMsg)
			if err != nil {
				errorCount++
			} else {
				successCount++
			}
		}
	} else {
		json.NewEncoder(w).Encode(DiscordBotResponse{
			Success: false,
			Message: "Invalid message type. Must be 'lead' or 'hint'",
		})
		return
	}

	json.NewEncoder(w).Encode(DiscordBotResponse{
		Success: true,
		Message: fmt.Sprintf("Sync completed: %d successful, %d errors", successCount, errorCount),
		Data: map[string]interface{}{
			"successCount":  successCount,
			"errorCount":    errorCount,
			"totalMessages": len(req.Messages),
		},
	})
}
