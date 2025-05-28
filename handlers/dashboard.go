package handlers

import (
	"encoding/json"
	"intrasudo25/database"
	"net/http"
	"strings"
)

func DashboardPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Dashboard"})
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

	question, err := database.GetLevel(int(id))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Question not found"})
		return
	}

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
	level, err := database.GetLevel(currentLevel)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Level not found"})
		return
	}

	userAnswerTrimmed := strings.TrimSpace(strings.ToLower(userAnswer))
	correctAnswerTrimmed := strings.TrimSpace(strings.ToLower(level.Answer))

	w.Header().Set("Content-Type", "application/json")
	if userAnswerTrimmed == correctAnswerTrimmed {
		nextLevel := currentLevel + 1
		err = database.UpdateField(login.Gmail, "On", nextLevel)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update progress"})
			return
		}

		currentScore, _ := database.GetUserScore(login.Gmail)
		newScore := currentScore + 10
		database.UpdateScore(login.Gmail, newScore)

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
