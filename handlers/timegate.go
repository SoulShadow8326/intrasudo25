package handlers

import (
	"encoding/json"
	"fmt"
	"intrasudo25/config"
	"net/http"
	"strings"
	"time"
)

func TimeGateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("TimeGate: Path=%s\n", r.URL.Path)

		if !config.IsCountdownEnabled() {
			fmt.Printf("TimeGate: Countdown disabled\n")
			next(w, r)
			return
		}

		user, err := GetUserFromSession(r)
		if err == nil && user != nil {
			adminEmails := config.GetAdminEmails()
			for _, adminEmail := range adminEmails {
				if strings.EqualFold(user.Gmail, adminEmail) {
					fmt.Printf("TimeGate: Admin bypass for %s\n", user.Gmail)
					next(w, r)
					return
				}
			}
		}

		location, _ := time.LoadLocation("Asia/Kolkata")
		now := time.Now().In(location)
		startTime := config.GetCompetitionStartTime()
		endTime := config.GetCompetitionEndTime()

		fmt.Printf("TimeGate: Now=%s Start=%s End=%s\n", now, startTime, endTime)
		fmt.Printf("TimeGate: Before=%t After=%t\n", now.Before(startTime), now.After(endTime))

		if now.Before(startTime) || now.After(endTime) {
			fmt.Printf("TimeGate: Redirecting to /status\n")
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, private")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			w.Header().Set("Location", "/status")
			w.WriteHeader(http.StatusSeeOther)
			return
		}

		fmt.Printf("TimeGate: Allowing access\n")
		next(w, r)
	}
}

type CountdownStatus struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Details string `json:"details"`
}

func CountdownStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if !config.IsCountdownEnabled() {
		json.NewEncoder(w).Encode(CountdownStatus{
			Status:  "active",
			Message: "Competition is active",
			Details: "Welcome to Intra Sudo v6.0!",
		})
		return
	}

	location, _ := time.LoadLocation("Asia/Kolkata")
	now := time.Now().In(location)
	startTime := config.GetCompetitionStartTime()
	endTime := config.GetCompetitionEndTime()

	var response CountdownStatus

	if now.Before(startTime) {
		response.Status = "not_started"
		response.Message = "Intra Sudo v6.0 has not begun yet"
		startDateFormatted := startTime.Format("January 2nd, 2006")
		startTimeFormatted := startTime.Format("3:04 PM")
		response.Details = "The competition will start on " + startDateFormatted + " at " + startTimeFormatted + " IST. Please check back then!"
	} else if now.After(endTime) {
		response.Status = "ended"
		response.Message = "Intra Sudo v6.0 is now over"
		response.Details = "Thank you for participating! Results will be announced shortly."
	} else {
		response.Status = "active"
		response.Message = "Competition is active"
		response.Details = "Welcome to Intra Sudo v6.0!"
	}

	json.NewEncoder(w).Encode(response)
}
