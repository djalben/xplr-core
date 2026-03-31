package user

import (
	"context"
	"strings"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/pkg/utils"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type UseCase struct {
	userRepo       ports.UserRepository
	walletRepo     ports.WalletRepository
	gradeRepo      ports.GradeRepository
	referralRepo   ports.ReferralRepository
	commissionRepo ports.CommissionConfigRepository
}

func NewUseCase(
	userRepo ports.UserRepository,
	walletRepo ports.WalletRepository,
	gradeRepo ports.GradeRepository,
	referralRepo ports.ReferralRepository,
	commissionRepo ports.CommissionConfigRepository,
) *UseCase {
	return &UseCase{
		userRepo:       userRepo,
		walletRepo:     walletRepo,
		gradeRepo:      gradeRepo,
		referralRepo:   referralRepo,
		commissionRepo: commissionRepo,
	}
}

// GetByID — пользователь по id (BFF / совместимость).
func (uc *UseCase) GetByID(ctx context.Context, userID domain.UUID) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, userID)
}

// GetMe возвращает данные текущего пользователя для фронта (BFF-совместимость).
func (uc *UseCase) GetMe(ctx context.Context, userID domain.UUID) (map[string]any, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	balance, _ := uc.walletRepo.GetByUserID(ctx, userID)
	balanceStr := "0"
	if balance != nil {
		balanceStr = balance.Balance.String()
	}

	displayName := user.Email
	if at := strings.Index(user.Email, "@"); at > 0 {
		displayName = user.Email[:at]
	}

	role := "user"
	if user.IsAdmin {
		role = "admin"
	}

	sbpEnabled := true
	sbpMessage := ""

	if uc.commissionRepo != nil {
		cfg, err := uc.commissionRepo.GetByKey(ctx, "sbp_topup_enabled")
		if err == nil && cfg.Value.LessThan(domain.NewNumeric(0.5)) {
			sbpEnabled = false
			sbpMessage = "Пополнение через СБП временно недоступно. Выберите другой способ или попробуйте позже."
		}
	}

	return map[string]any{
		"id":                     user.ID.String(),
		"email":                  user.Email,
		"display_name":           displayName,
		"balance":                balanceStr,
		"status":                 string(user.Status),
		"is_admin":               user.IsAdmin,
		"role":                   role,
		"email_verified":         user.EmailVerified,
		"kyc_status":             string(user.KYCStatus),
		"totp_enabled":           user.TOTPEnabled,
		"notify_email":           user.NotifyEmail,
		"notify_telegram":        user.NotifyTelegram,
		"notify_transactions":    user.NotifyTransactions,
		"notify_balance":         user.NotifyBalance,
		"notify_security":        user.NotifySecurity,
		"notify_card_operations": user.NotifyCardOperations,
		"sbp_topup_enabled":      sbpEnabled,
		"sbp_topup_message":      sbpMessage,
	}, nil
}

// SetNotificationPreferences — минимум один канал (email и/или telegram) + привязки аккаунтов.
func (uc *UseCase) SetNotificationPreferences(ctx context.Context, userID domain.UUID, notifyEmail, notifyTelegram bool) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	user.NotifyEmail = notifyEmail
	user.NotifyTelegram = notifyTelegram

	return uc.validateAndSaveNotificationChannels(ctx, user)
}

// PartnerNotifPatch — тело PATCH /user/settings/notifications (фронт партнёра).
type PartnerNotifPatch struct {
	NotificationPref     *string
	NotifyTransactions   *bool
	NotifyBalance        *bool
	NotifySecurity       *bool
	NotifyCardOperations *bool
}

// PatchPartnerNotificationSettings — канал (both|email|telegram) и типы уведомлений.
func (uc *UseCase) PatchPartnerNotificationSettings(ctx context.Context, userID domain.UUID, p PartnerNotifPatch) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	if p.NotificationPref != nil {
		switch *p.NotificationPref {
		case "both":
			user.NotifyEmail = true
			user.NotifyTelegram = true
		case "email":
			user.NotifyEmail = true
			user.NotifyTelegram = false
		case "telegram":
			user.NotifyEmail = false
			user.NotifyTelegram = true
		default:
			return domain.NewInvalidInput("notification_pref must be both, email or telegram")
		}
	}

	if p.NotifyTransactions != nil {
		user.NotifyTransactions = *p.NotifyTransactions
	}

	if p.NotifyBalance != nil {
		user.NotifyBalance = *p.NotifyBalance
	}

	if p.NotifySecurity != nil {
		user.NotifySecurity = *p.NotifySecurity
	}

	if p.NotifyCardOperations != nil {
		user.NotifyCardOperations = *p.NotifyCardOperations
	}

	return uc.validateAndSaveNotificationChannels(ctx, user)
}

// ChangePassword — смена пароля в ЛК.
func (uc *UseCase) ChangePassword(ctx context.Context, userID domain.UUID, oldPassword, newPassword string) error {
	if oldPassword == "" || newPassword == "" {
		return domain.NewInvalidInput("old and new password are required")
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	if !utils.CheckPasswordHash(oldPassword, user.PasswordHash) {
		return domain.NewInvalidInput("invalid current password")
	}

	newHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return wrapper.Wrap(err)
	}

	user.PasswordHash = newHash

	return uc.userRepo.Update(ctx, user)
}

// UnlinkTelegram — отвязать чат; канал telegram в уведомлениях сбрасывается при необходимости.
func (uc *UseCase) UnlinkTelegram(ctx context.Context, userID domain.UUID) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	user.TelegramChatID = nil
	user.TelegramLinkCode = nil
	user.TelegramLinkExpiresAt = nil

	if user.NotifyTelegram {
		user.NotifyTelegram = false
	}

	if !user.NotifyEmail {
		user.NotifyEmail = true
	}

	return uc.validateAndSaveNotificationChannels(ctx, user)
}

// IssueTelegramLinkCode — одноразовый код для привязки Telegram (бот передаёт пользователю).
func (uc *UseCase) IssueTelegramLinkCode(ctx context.Context, userID domain.UUID) (code string, expiresAt time.Time, err error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", time.Time{}, wrapper.Wrap(err)
	}

	plain, _, err := utils.RandomTokenHex(4)
	if err != nil {
		return "", time.Time{}, err
	}

	exp := time.Now().UTC().Add(15 * time.Minute)
	user.TelegramLinkCode = &plain
	user.TelegramLinkExpiresAt = &exp

	err = uc.userRepo.Update(ctx, user)
	if err != nil {
		return "", time.Time{}, wrapper.Wrap(err)
	}

	return plain, exp, nil
}

// LinkTelegram — привязка chat_id после ввода кода (уникальность chat_id в БД).
func (uc *UseCase) LinkTelegram(ctx context.Context, userID domain.UUID, chatID int64, code string) error {
	if chatID == 0 || code == "" {
		return domain.NewInvalidInput("chat_id and code are required")
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	if user.TelegramLinkCode == nil || *user.TelegramLinkCode != code {
		return domain.NewInvalidInput("invalid link code")
	}

	if user.TelegramLinkExpiresAt == nil || time.Now().UTC().After(*user.TelegramLinkExpiresAt) {
		return domain.NewInvalidInput("link code expired")
	}

	other, err := uc.userRepo.GetByTelegramChatID(ctx, chatID)
	if err != nil && !isNoRowsUser(err) {
		return wrapper.Wrap(err)
	}

	if other != nil && other.ID != userID {
		return domain.NewAlreadyExists("этот Telegram уже привязан к другому аккаунту")
	}

	user.TelegramChatID = &chatID
	user.TelegramLinkCode = nil
	user.TelegramLinkExpiresAt = nil

	return uc.userRepo.Update(ctx, user)
}

func isNoRowsUser(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(strings.ToLower(err.Error()), "no rows")
}

// GetReferralInfo — данные реферальной программы.
func (uc *UseCase) GetReferralInfo(ctx context.Context, userID domain.UUID) (map[string]any, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	code := user.ReferralCode
	if code == "" {
		idStr := strings.ReplaceAll(user.ID.String(), "-", "")
		if len(idStr) > 8 {
			idStr = idStr[:8]
		}
		code = "XPLR" + strings.ToUpper(idStr)
		user.ReferralCode = code

		_ = uc.userRepo.Update(ctx, user)
	}

	totalReferrals, _ := uc.referralRepo.CountByReferrer(ctx, userID)
	earnings, _ := uc.referralRepo.TotalEarningsByReferrer(ctx, userID)

	link := "https://xplr.app/ref/" + code

	return map[string]any{
		"referral_code":       code,
		"referral_link":       link,
		"reward_per_referral": 10,
		"bonus_for_new":       5,
		"stats": map[string]any{
			"total_referrals": totalReferrals,
			"total_earnings":  earnings.String(),
			"pending_amount":  "0",
		},
		"recent_referrals": []any{},
	}, nil
}

func (uc *UseCase) validateAndSaveNotificationChannels(ctx context.Context, user *domain.User) error {
	if !user.NotifyEmail && !user.NotifyTelegram {
		return domain.NewInvalidInput("выберите хотя бы один канал уведомлений: email или telegram")
	}

	if user.NotifyEmail && !user.EmailVerified {
		return domain.NewInvalidInput("подтвердите email, чтобы получать уведомления на почту")
	}

	if user.NotifyTelegram && (user.TelegramChatID == nil || *user.TelegramChatID == 0) {
		return domain.NewInvalidInput("привяжите Telegram, чтобы получать уведомления в Telegram")
	}

	return uc.userRepo.Update(ctx, user)
}
