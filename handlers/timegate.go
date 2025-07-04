package handlers

import (
	"crypto/md5"
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

	fmt.Printf("DEBUG CountdownStatus: Now=%s Start=%s End=%s\n", now.Format("2006-01-02 15:04:05"), startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("DEBUG CountdownStatus: Before=%t After=%t\n", now.Before(startTime), now.After(endTime))

	var response CountdownStatus

	if now.Before(startTime) {
		response.Status = "not_started"
		response.Message = "Intra Sudo v6.0 has not begun yet"
		day := startTime.Day()
		var suffix string
		switch {
		case day >= 11 && day <= 13:
			suffix = "th"
		case day%10 == 1:
			suffix = "st"
		case day%10 == 2:
			suffix = "nd"
		case day%10 == 3:
			suffix = "rd"
		default:
			suffix = "th"
		}
		startDateFormatted := startTime.Format("January 2") + suffix + startTime.Format(", 2006")
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

	fmt.Printf("DEBUG CountdownStatus: Final status=%s\n", response.Status)
	json.NewEncoder(w).Encode(response)
}

type CountdownChecksum struct {
	Checksum string `json:"checksum"`
}

func CountdownChecksumHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if !config.IsCountdownEnabled() {
		checksum := fmt.Sprintf("%x", md5.Sum([]byte("active")))
		json.NewEncoder(w).Encode(CountdownChecksum{Checksum: checksum})
		return
	}

	location, _ := time.LoadLocation("Asia/Kolkata")
	now := time.Now().In(location)
	startTime := config.GetCompetitionStartTime()
	endTime := config.GetCompetitionEndTime()

	var status string
	if now.Before(startTime) {
		status = "not_started"
	} else if now.After(endTime) {
		status = "ended"
	} else {
		status = "active"
	}

	// Create checksum based on status and timestamps to detect changes
	checksumData := fmt.Sprintf("%s-%s-%s", status, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))
	checksum := fmt.Sprintf("%x", md5.Sum([]byte(checksumData)))

	json.NewEncoder(w).Encode(CountdownChecksum{Checksum: checksum})
}
