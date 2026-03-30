package user

import (
	"context"
	"strings"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type UseCase struct {
	userRepo     ports.UserRepository
	walletRepo   ports.WalletRepository
	gradeRepo    ports.GradeRepository
	referralRepo ports.ReferralRepository
}

func NewUseCase(userRepo ports.UserRepository, walletRepo ports.WalletRepository, gradeRepo ports.GradeRepository, referralRepo ports.ReferralRepository) *UseCase {
	return &UseCase{
		userRepo:     userRepo,
		walletRepo:   walletRepo,
		gradeRepo:    gradeRepo,
		referralRepo: referralRepo,
	}
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

	return map[string]any{
		"id":           user.ID.String(),
		"email":        user.Email,
		"display_name": displayName,
		"balance":      balanceStr,
		"status":       string(user.Status),
		"is_admin":     user.IsAdmin,
		"role":         role,
	}, nil
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
