package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/pkg/utils"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/pquerna/otp/totp"
	"gitlab.com/libs-artifex/wrapper/v2"
)

const (
	emailVerifyTTL   = 48 * time.Hour
	resetPasswordTTL = time.Hour
	totpIssuer       = "XPLR"
)

// LoginResult — итог Login: либо полный вход, либо нужен второй шаг TOTP.
type LoginResult struct {
	User     *domain.User
	MFAToken string
}

// UseCase — регистрация, вход, email, сброс пароля, TOTP.
type UseCase struct {
	userRepo      ports.UserRepository
	walletRepo    ports.WalletRepository
	gradeRepo     ports.GradeRepository
	jwtSecret     []byte
	mailer        ports.Mailer
	publicBaseURL string
}

// NewUseCase — publicBaseURL без завершающего «/» (для ссылок в письмах).
func NewUseCase(
	userRepo ports.UserRepository,
	walletRepo ports.WalletRepository,
	gradeRepo ports.GradeRepository,
	jwtSecret []byte,
	mailer ports.Mailer,
	publicBaseURL string,
) *UseCase {
	return &UseCase{
		userRepo:      userRepo,
		walletRepo:    walletRepo,
		gradeRepo:     gradeRepo,
		jwtSecret:     jwtSecret,
		mailer:        mailer,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
	}
}

func (uc *UseCase) Register(ctx context.Context, email, password string) (*domain.User, error) {
	if email == "" || password == "" {
		return nil, domain.NewInvalidInput("email and password are required")
	}

	_, err := uc.userRepo.GetByEmail(ctx, email)
	if err == nil {
		return nil, domain.NewInvalidInput("email already registered")
	}

	if !isNoRows(err) {
		return nil, wrapper.Wrap(err)
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	user, err := domain.NewUser(email, hash)
	if err != nil {
		return nil, err
	}

	// ReferralCode нужен сразу при INSERT (поле UNIQUE); формируем детерминированно от UUID.
	// Если позже потребуется более «красивый» код — можно менять формат, но важно сохранять уникальность.
	idStr := strings.ReplaceAll(user.ID.String(), "-", "")
	if len(idStr) > 8 {
		idStr = idStr[:8]
	}
	user.ReferralCode = "XPLR" + strings.ToUpper(idStr)

	plain, hashTok, err := utils.RandomTokenHex(32)
	if err != nil {
		return nil, err
	}

	exp := time.Now().UTC().Add(emailVerifyTTL)
	user.EmailVerifyTokenHash = &hashTok
	user.EmailVerifyExpiresAt = &exp

	err = uc.userRepo.Save(ctx, user)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	err = uc.walletRepo.EnsureWallet(ctx, user.ID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	err = uc.gradeRepo.EnsureGrade(ctx, user.ID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	verifyURL := uc.publicBaseURL + "/api/v1/auth/verify-email?token=" + plain
	body := "Подтвердите email, перейдя по ссылке:\n" + verifyURL

	err = uc.mailer.SendPlain(ctx, email, "Подтверждение регистрации XPLR", body)
	if err != nil {
		return nil, wrapper.Wrapf(err, "failed to send verification email")
	}

	return user, nil
}

func (uc *UseCase) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	if email == "" || password == "" {
		return nil, domain.NewInvalidInput("email and password are required")
	}

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if isNoRows(err) {
			return nil, domain.NewInvalidInput("invalid email or password")
		}

		return nil, wrapper.Wrap(err)
	}

	if user.Status != domain.UserStatusActive {
		return nil, domain.NewInvalidInput("account is blocked")
	}

	if !user.EmailVerified {
		return nil, domain.NewInvalidInput("email not verified")
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return nil, domain.NewInvalidInput("invalid email or password")
	}

	if user.TOTPEnabled && user.TOTPSecret != nil && *user.TOTPSecret != "" {
		mfaTok, errJWT := utils.GenerateMFAPendingJWT(uc.jwtSecret, user.ID, user.Email)
		if errJWT != nil {
			return nil, wrapper.Wrap(errJWT)
		}

		return &LoginResult{User: user, MFAToken: mfaTok}, nil
	}

	return &LoginResult{User: user}, nil
}

// CompleteMFALogin — второй шаг входа с TOTP.
func (uc *UseCase) CompleteMFALogin(ctx context.Context, mfaToken, totpCode string) (*domain.User, error) {
	if mfaToken == "" || totpCode == "" {
		return nil, domain.NewInvalidInput("mfa token and totp code are required")
	}

	userID, _, err := utils.ValidateMFAPendingJWT(uc.jwtSecret, mfaToken)
	if err != nil {
		return nil, domain.NewInvalidInput("invalid or expired mfa token")
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	if user.TOTPSecret == nil || *user.TOTPSecret == "" || !user.TOTPEnabled {
		return nil, domain.NewInvalidInput("totp is not enabled for this account")
	}

	if !totp.Validate(totpCode, *user.TOTPSecret) {
		return nil, domain.NewInvalidInput("invalid totp code")
	}

	return user, nil
}

// VerifyEmail — подтверждение email по одноразовому токену из письма.
func (uc *UseCase) VerifyEmail(ctx context.Context, plainToken string) error {
	if plainToken == "" {
		return domain.NewInvalidInput("token is required")
	}

	hashTok := utils.HashTokenHex(plainToken)

	user, err := uc.userRepo.GetByEmailVerifyTokenHash(ctx, hashTok)
	if err != nil {
		if isNoRows(err) {
			return domain.NewInvalidInput("invalid or expired verification token")
		}

		return wrapper.Wrap(err)
	}

	if user.EmailVerifyExpiresAt == nil || time.Now().UTC().After(*user.EmailVerifyExpiresAt) {
		return domain.NewInvalidInput("verification token expired")
	}

	user.EmailVerified = true
	user.EmailVerifyTokenHash = nil
	user.EmailVerifyExpiresAt = nil

	return uc.userRepo.Update(ctx, user)
}

// RequestPasswordReset — отправляет письмо со ссылкой сброса (если email известен).
func (uc *UseCase) RequestPasswordReset(ctx context.Context, email string) error {
	if email == "" {
		return nil
	}

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if isNoRows(err) {
			return nil
		}

		return wrapper.Wrap(err)
	}

	plain, hashTok, err := utils.RandomTokenHex(32)
	if err != nil {
		return err
	}

	exp := time.Now().UTC().Add(resetPasswordTTL)
	user.PasswordResetTokenHash = &hashTok
	user.PasswordResetExpiresAt = &exp

	err = uc.userRepo.Update(ctx, user)
	if err != nil {
		return wrapper.Wrap(err)
	}

	resetURL := uc.publicBaseURL + "/api/v1/auth/reset-password?token=" + plain
	body := "Сброс пароля XPLR:\n" + resetURL + "\nСсылка действует 1 час."

	_ = uc.mailer.SendPlain(ctx, email, "Сброс пароля XPLR", body) // не раскрываем пользователю наличие email в базе

	return nil
}

// ResetPassword — установка нового пароля по токену из письма.
func (uc *UseCase) ResetPassword(ctx context.Context, plainToken, newPassword string) error {
	if plainToken == "" || newPassword == "" {
		return domain.NewInvalidInput("token and new password are required")
	}

	hashTok := utils.HashTokenHex(plainToken)

	user, err := uc.userRepo.GetByPasswordResetTokenHash(ctx, hashTok)
	if err != nil {
		if isNoRows(err) {
			return domain.NewInvalidInput("invalid or expired reset token")
		}

		return wrapper.Wrap(err)
	}

	if user.PasswordResetExpiresAt == nil || time.Now().UTC().After(*user.PasswordResetExpiresAt) {
		return domain.NewInvalidInput("reset token expired")
	}

	newHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return wrapper.Wrap(err)
	}

	user.PasswordHash = newHash
	user.PasswordResetTokenHash = nil
	user.PasswordResetExpiresAt = nil

	return uc.userRepo.Update(ctx, user)
}

// SetupTOTP — генерирует секрет и сохраняет в профиле (ещё без включения 2FA).
func (uc *UseCase) SetupTOTP(ctx context.Context, userID domain.UUID) (otpauthURL, plainSecret string, err error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", "", wrapper.Wrap(err)
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      totpIssuer,
		AccountName: user.Email,
	})
	if err != nil {
		return "", "", wrapper.Wrap(err)
	}

	sec := key.Secret()
	user.TOTPSecret = &sec
	user.TOTPEnabled = false

	err = uc.userRepo.Update(ctx, user)
	if err != nil {
		return "", "", wrapper.Wrap(err)
	}

	return key.URL(), sec, nil
}

// ConfirmTOTP — подтверждает код и включает 2FA.
func (uc *UseCase) ConfirmTOTP(ctx context.Context, userID domain.UUID, code string) error {
	if code == "" {
		return domain.NewInvalidInput("code is required")
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	if user.TOTPSecret == nil || *user.TOTPSecret == "" {
		return domain.NewInvalidInput("run totp setup first")
	}

	if !totp.Validate(code, *user.TOTPSecret) {
		return domain.NewInvalidInput("invalid totp code")
	}

	user.TOTPEnabled = true

	return uc.userRepo.Update(ctx, user)
}

// DisableTOTP — выключает 2FA после проверки пароля и текущего TOTP.
func (uc *UseCase) DisableTOTP(ctx context.Context, userID domain.UUID, password, code string) error {
	if password == "" || code == "" {
		return domain.NewInvalidInput("password and totp code are required")
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return domain.NewInvalidInput("invalid password")
	}

	if user.TOTPSecret == nil || *user.TOTPSecret == "" || !user.TOTPEnabled {
		return domain.NewInvalidInput("totp is not enabled")
	}

	if !totp.Validate(code, *user.TOTPSecret) {
		return domain.NewInvalidInput("invalid totp code")
	}

	user.TOTPEnabled = false
	user.TOTPSecret = nil

	return uc.userRepo.Update(ctx, user)
}

// ResendEmailVerification — повторная отправка письма с ссылкой подтверждения.
func (uc *UseCase) ResendEmailVerification(ctx context.Context, userID domain.UUID) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	if user.EmailVerified {
		return nil
	}

	plain, hashTok, err := utils.RandomTokenHex(32)
	if err != nil {
		return err
	}

	exp := time.Now().UTC().Add(emailVerifyTTL)
	user.EmailVerifyTokenHash = &hashTok
	user.EmailVerifyExpiresAt = &exp

	err = uc.userRepo.Update(ctx, user)
	if err != nil {
		return wrapper.Wrap(err)
	}

	verifyURL := uc.publicBaseURL + "/api/v1/auth/verify-email?token=" + plain
	body := "Подтвердите email:\n" + verifyURL

	_ = uc.mailer.SendPlain(ctx, user.Email, "Подтверждение email XPLR", body)

	return nil
}

func isNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
