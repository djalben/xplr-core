package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/service"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

// AdminDashboardStatsHandler - GET /api/v1/admin/dashboard
func AdminDashboardStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := repository.GetAdminDashboardStats()
	if err != nil {
		log.Printf("AdminDashboardStats: error: %v", err)
		http.Error(w, "Failed to fetch stats", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// AdminSearchUsersHandler - GET /api/v1/admin/users/search?q=email&limit=50
func AdminSearchUsersHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if strings.TrimSpace(q) == "" {
		http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	users, err := repository.SearchUsersByEmail(q, limit)
	if err != nil {
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// AdminUpdateUserGradeHandler - PATCH /api/v1/admin/users/{id}/grade
func AdminUpdateUserGradeHandler(w http.ResponseWriter, r *http.Request) {
	adminID, _ := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	targetID, err := strconv.Atoi(vars["id"])
	if err != nil || targetID <= 0 {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	var req struct {
		Grade string `json:"grade"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	grade := strings.TrimSpace(strings.ToUpper(req.Grade))
	if err := repository.AdminUpdateUserGrade(targetID, grade); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	repository.WriteAdminLog(adminID, fmt.Sprintf("Изменен грейд юзера %d на %s", targetID, grade))

	// Отправка уведомления о смене грейда (async)
	go func(uid int, g string) {
		targetUser, err := repository.GetUserByID(uid)
		if err != nil {
			log.Printf("[EMAIL] Cannot fetch user %d for grade email: %v", uid, err)
			return
		}
		if err := service.SendGradeChangeEmail(targetUser.Email, g); err != nil {
			log.Printf("[EMAIL] Failed to send grade change email to user %d: %v", uid, err)
		}
	}(targetID, grade)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"user_id": targetID, "grade": grade})
}

// AdminGetCommissionConfigHandler - GET /api/v1/admin/commissions
func AdminGetCommissionConfigHandler(w http.ResponseWriter, r *http.Request) {
	configs, err := repository.GetAllCommissionConfigs()
	if err != nil {
		http.Error(w, "Failed to fetch commission config", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configs)
}

// AdminToggleBlockHandler - POST /api/v1/admin/users/{id}/toggle-block
func AdminToggleBlockHandler(w http.ResponseWriter, r *http.Request) {
	adminID, _ := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	targetID, err := strconv.Atoi(vars["id"])
	if err != nil || targetID <= 0 {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	newBlocked, email, err := repository.ToggleUserBlock(targetID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	action := "разблокирован"
	if newBlocked {
		action = "заблокирован"
	}
	repository.WriteAdminLog(adminID, fmt.Sprintf("Пользователь %s (ID %d) %s", email, targetID, action))
	log.Printf("[ADMIN] User %d (%s) %s by admin %d", targetID, email, action, adminID)

	// Notify user about block/unblock (email-only for block, since TG may be unavailable)
	if newBlocked && email != "" {
		log.Printf("[EVENT] Admin %d performed user_block (target=%d, email=%s). Triggering notifications...", adminID, targetID, email)
		go func() {
			if err := service.SendGenericEmail(email, "Аккаунт заблокирован",
				"🔒 <b>Ваш аккаунт временно заблокирован администратором.</b><br><br>"+
					"Если вы считаете, что это ошибка, обратитесь в поддержку: "+
					"<a href=\"mailto:admin@xplr.pro\">admin@xplr.pro</a>"); err != nil {
				log.Printf("[NOTIFY] ❌ Failed to send block email to user %d (%s): %v", targetID, email, err)
			} else {
				log.Printf("[NOTIFY] ✅ Block notification email sent to user %d (%s)", targetID, email)
			}
		}()
	} else if !newBlocked {
		log.Printf("[EVENT] Admin %d performed user_unblock (target=%d, email=%s). Triggering notifications...", adminID, targetID, email)
		go service.NotifyUser(targetID, "Аккаунт разблокирован",
			"✅ <b>Ваш аккаунт был разблокирован.</b>\n\nВы снова можете пользоваться всеми сервисами XPLR.\n\n"+
				"<a href=\"https://xplr.pro\">Открыть XPLR</a>")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"is_blocked": newBlocked,
		"email":      email,
	})
}

// AdminUpdateCommissionConfigHandler - PATCH /api/v1/admin/commissions/{id}
func AdminUpdateCommissionConfigHandler(w http.ResponseWriter, r *http.Request) {
	adminID, _ := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	configID, err := strconv.Atoi(vars["id"])
	if err != nil || configID <= 0 {
		http.Error(w, "invalid config id", http.StatusBadRequest)
		return
	}
	var req struct {
		Value decimal.Decimal `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := repository.UpdateCommissionConfig(configID, req.Value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	repository.WriteAdminLog(adminID, fmt.Sprintf("Изменена комиссия config_id=%d, value=%s", configID, req.Value.String()))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": configID, "value": req.Value.String()})
}

// AdminGetSupportTicketsHandler - GET /api/v1/admin/tickets?status=open
func AdminGetSupportTicketsHandler(w http.ResponseWriter, r *http.Request) {
	statusFilter := r.URL.Query().Get("status")
	tickets, err := repository.GetAllSupportTickets(statusFilter)
	if err != nil {
		http.Error(w, "Failed to fetch tickets", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tickets)
}

// AdminUpdateTicketStatusHandler - PATCH /api/v1/admin/tickets/{id}
// SECURITY: Only the claiming admin or super-admin can close/resolve a ticket.
func AdminUpdateTicketStatusHandler(w http.ResponseWriter, r *http.Request) {
	adminID, _ := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	ticketID, err := strconv.Atoi(vars["id"])
	if err != nil || ticketID <= 0 {
		http.Error(w, "invalid ticket id", http.StatusBadRequest)
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	// Ownership check for close/resolve: only claimer or super-admin
	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status == "resolved" || status == "closed" {
		claimedBy := repository.GetSupportTicketClaimedBy(ticketID)
		if claimedBy != 0 && claimedBy != adminID {
			// Check if super-admin
			caller, err := repository.GetUserByID(adminID)
			if err != nil || caller.Email != superAdminEmail {
				log.Printf("[SECURITY] ⛔ Admin %d tried to %s ticket %d owned by admin %d", adminID, status, ticketID, claimedBy)
				http.Error(w, "Этот тикет находится в работе у другого администратора", http.StatusForbidden)
				return
			}
		}
	}

	if err := repository.UpdateSupportTicketStatus(ticketID, req.Status, adminID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	repository.WriteAdminLog(adminID, fmt.Sprintf("Обновлен статус тикета %d на %s", ticketID, req.Status))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": ticketID, "status": req.Status})
}

// AdminEmergencyFreezeHandler - POST /api/v1/admin/users/{id}/emergency-freeze
func AdminEmergencyFreezeHandler(w http.ResponseWriter, r *http.Request) {
	adminID, _ := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	targetID, err := strconv.Atoi(vars["id"])
	if err != nil || targetID <= 0 {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	frozenCards, err := repository.EmergencyFreezeUser(targetID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	repository.WriteAdminLog(adminID, fmt.Sprintf("🚨 EMERGENCY FREEZE юзера %d — %d карт заморожено, статус BANNED, баланс обнулён", targetID, frozenCards))

	// Уведомление пользователю о блокировке (async)
	go func(uid, cards int) {
		targetUser, err := repository.GetUserByID(uid)
		if err != nil {
			log.Printf("[EMAIL] Cannot fetch user %d for freeze notification: %v", uid, err)
			return
		}
		if err := service.SendEmergencyFreezeNotification(targetUser.Email, cards); err != nil {
			log.Printf("[EMAIL] Failed to send freeze notification to user %d: %v", uid, err)
		}
	}(targetID, frozenCards)

	// Notify user via NotifyUser (respects user's channel pref)
	go service.NotifyUser(targetID, "Аккаунт заблокирован",
		fmt.Sprintf("🚨 <b>Ваш аккаунт заблокирован</b>\n\n"+
			"Заморожено карт: <b>%d</b>\n"+
			"Статус: <b>BANNED</b>\n"+
			"Баланс: <b>обнулён</b>\n\n"+
			"Если вы считаете, что это ошибка — свяжитесь с поддержкой.", frozenCards))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":      targetID,
		"frozen_cards": frozenCards,
		"status":       "BANNED",
		"balance":      "0",
	})
}

// AdminGetChatsHandler - GET /api/v1/admin/chats?status=open
func AdminGetChatsHandler(w http.ResponseWriter, r *http.Request) {
	statusFilter := r.URL.Query().Get("status")
	chats, err := repository.GetAllChatConversationsForAdmin(statusFilter)
	if err != nil {
		log.Printf("[ADMIN] Failed to fetch chats: %v", err)
		http.Error(w, "Failed to fetch chats", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chats)
}

// AdminGetChatMessagesHandler - GET /api/v1/admin/chats/{id}/messages
func AdminGetChatMessagesHandler(w http.ResponseWriter, r *http.Request) {
	convIDStr := mux.Vars(r)["id"]
	convID, err := strconv.Atoi(convIDStr)
	if err != nil || convID <= 0 {
		http.Error(w, "invalid chat id", http.StatusBadRequest)
		return
	}
	msgs, err := repository.GetChatMessages(convID)
	if err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}
	if msgs == nil {
		msgs = []repository.ChatMessage{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msgs)
}

// AdminGetTranslationsHandler - GET /api/v1/admin/translations?lang=ru
func AdminGetTranslationsHandler(w http.ResponseWriter, r *http.Request) {
	langFilter := r.URL.Query().Get("lang")
	translations, err := repository.GetAllTranslations(langFilter)
	if err != nil {
		log.Printf("[ADMIN] Failed to fetch translations: %v", err)
		http.Error(w, "Failed to fetch translations", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(translations)
}

// AdminUpsertTranslationHandler - PUT /api/v1/admin/translations
func AdminUpsertTranslationHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MsgKey string `json:"msg_key"`
		Lang   string `json:"lang"`
		Value  string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.MsgKey == "" || req.Lang == "" {
		http.Error(w, "msg_key and lang are required", http.StatusBadRequest)
		return
	}
	if err := repository.UpsertTranslation(req.MsgKey, req.Lang, req.Value); err != nil {
		log.Printf("[ADMIN] Failed to upsert translation: %v", err)
		http.Error(w, "Failed to save translation", http.StatusInternalServerError)
		return
	}
	adminID, _ := r.Context().Value(middleware.UserIDKey).(int)
	repository.WriteAdminLog(adminID, fmt.Sprintf("Перевод обновлён: %s [%s]", req.MsgKey, req.Lang))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// AdminDeleteTranslationHandler - DELETE /api/v1/admin/translations/{id}
func AdminDeleteTranslationHandler(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := repository.DeleteTranslation(id); err != nil {
		http.Error(w, "Failed to delete translation", http.StatusInternalServerError)
		return
	}
	adminID, _ := r.Context().Value(middleware.UserIDKey).(int)
	repository.WriteAdminLog(adminID, fmt.Sprintf("Перевод удалён: id=%d", id))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// VerifyStaffPINHandler - POST /api/v1/verify-staff-pin
// Проверяет ПИН-код для доступа к админке.
// ПЕРВЫМ делом проверяет is_admin — если пользователь не админ, возвращает 403.
func VerifyStaffPINHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[AUTH] VerifyStaffPIN handler called: method=%s path=%s", r.Method, r.URL.Path)

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		log.Printf("[AUTH] PIN attempt: no userID in context — 401")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// ═══ ЖЕЛЕЗНЫЙ БАРЬЕР: проверяем is_admin ПЕРВЫМ делом ═══
	user, err := repository.GetUserByID(userID)
	if err != nil {
		log.Printf("[STAFF-PIN] ❌ User %d: DB error: %v", userID, err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if !user.IsAdmin && user.Role != "admin" {
		log.Printf("[STAFF-PIN] ⛔ DENIED: user %d (%s) is NOT admin — PIN check blocked", userID, user.Email)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Только для админов: проверяем ПИН
	var req struct {
		PIN string `json:"pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	expectedPIN := getEnvOrDefault("STAFF_PIN", "1337")
	log.Printf("[AUTH] PIN check for user %d: env STAFF_PIN set=%v, using default=%v", userID, os.Getenv("STAFF_PIN") != "", os.Getenv("STAFF_PIN") == "")
	if req.PIN != expectedPIN {
		log.Printf("[AUTH] PIN attempt for user %d: %v", userID, "DENIED — wrong PIN")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid_pin"})
		return
	}

	log.Printf("[AUTH] PIN attempt for user %d: %v", userID, "GRANTED — staff access")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "access": "granted"})
}

func getEnvOrDefault(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return fallback
}

// AdminGetLogsHandler - GET /api/v1/admin/logs
func AdminGetLogsHandler(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	logs, err := repository.GetAdminLogs(limit)
	if err != nil {
		http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
