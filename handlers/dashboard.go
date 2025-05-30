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
	var id uint
	if is_there, login := Authorize(r); is_there {
		id = login.On
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Login pending..."})
		return
	}

	result, err := database.Get("levels", map[string]interface{}{"number": int(id), "admin": false})
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

	// Use database function to check answer instead of accessing level.Answer directly
	result, err := database.Get("check_answer", map[string]interface{}{"level": currentLevel, "answer": userAnswerTrimmed})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to check answer"})
		return
	}
	isCorrect := result.(bool)

	w.Header().Set("Content-Type", "application/json")
	if isCorrect {
		nextLevel := currentLevel + 1
		err = database.Update("login_field", map[string]interface{}{"gmail": login.Gmail, "field": "on"}, nextLevel)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update progress"})
			return
		}

		result, _ := database.Get("user_score", map[string]interface{}{"gmail": login.Gmail})
		currentScore := 0
		if result != nil {
			currentScore = result.(int)
		}
		newScore := currentScore + 10
		database.Update("score", map[string]interface{}{"gmail": login.Gmail}, newScore)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":    "Correct answer!",
			"correct":    true,
			"next_level": nextLevel,
			"score":      newScore,
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":       "Incorrect answer. Try again!",
			"correct":       false,
			"current_level": currentLevel,
		})
	}
}
