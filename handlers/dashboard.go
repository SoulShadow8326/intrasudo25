package handlers

import (
	"encoding/json"
	"intrasudo25/database"
	"net/http"
	"strings"
)

func DashboardPage(w http.ResponseWriter, r *http.Request) {
	is_there, login := Authorize(r)
	if !is_there {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Login pending..."})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       login.Gmail,
		"email":    login.Gmail,
		"level":    login.On,
		"verified": login.Verified,
	})
}

func GetQuestionHandler(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromSession(r)
	if err != nil || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Login pending..."})
		return
	}

	result, err := database.Get("level", map[string]interface{}{"number": int(user.On), "admin": false})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Question not found"})
		return
	}
	question := result.(*database.Level)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"question": question,
	})
}

func SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	is_there, login := Authorize(r)
	if !is_there {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Login pending..."})
		return
	}

	r.ParseForm()
	userAnswer := r.FormValue("answer")
	if userAnswer == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Answer is required"})
		return
	}

	currentLevel := int(login.On)

	userAnswerTrimmed := strings.TrimSpace(userAnswer)

	// Use CheckAnswer function directly to get the complete result including ReloadPage flag
	result, err := database.CheckAnswer(login.Gmail, currentLevel, userAnswerTrimmed)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to check answer"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
