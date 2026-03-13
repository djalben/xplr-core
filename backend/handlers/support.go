package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/service"
)

// SubmitSupportTicketHandler — POST /api/v1/user/support
// Creates a support ticket and sends email notification to support@xplr.pro.
func SubmitSupportTicketHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	message := strings.TrimSpace(req.Message)
	if message == "" {
		http.Error(w, "Message cannot be empty", http.StatusBadRequest)
		return
	}

	// Get user email for the ticket
	user, err := repository.GetUserByID(userID)
	if err != nil {
		// Fallback to basic
		user, err = repository.GetUserByIDBasic(userID)
		if err != nil {
			log.Printf("[SUPPORT] Could not fetch user %d: %v", userID, err)
			http.Error(w, "Failed to identify user", http.StatusInternalServerError)
			return
		}
	}

	// Truncate to make a subject line (first 80 chars)
	subject := message
	if len(subject) > 80 {
		subject = subject[:80] + "..."
	}

	ticketID, err := repository.CreateSupportTicket(userID, user.Email, subject, message)
	if err != nil {
		log.Printf("[SUPPORT] Failed to create ticket for user %d: %v", userID, err)
		http.Error(w, "Failed to create support ticket", http.StatusInternalServerError)
		return
	}

	// Send email notification to support (async, don't block response)
	go func() {
		if err := service.SendSupportTicketNotification(ticketID, user.Email, subject, message); err != nil {
			log.Printf("[SUPPORT] Email notification failed for ticket #%d: %v", ticketID, err)
		}
	}()

	log.Printf("[SUPPORT] ✅ Ticket #%d created by %s (user_id=%d)", ticketID, user.Email, userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      ticketID,
		"status":  "open",
		"message": "Ticket created successfully",
	})
}
