package auth

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/djalben/xplr-core/backend/internal/pkg/utils"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type UseCase struct {
	userRepo   ports.UserRepository
	walletRepo ports.WalletRepository
	gradeRepo  ports.GradeRepository
	jwtSecret  []byte
}

func NewUseCase(userRepo ports.UserRepository, walletRepo ports.WalletRepository, gradeRepo ports.GradeRepository, jwtSecret []byte) *UseCase {
	return &UseCase{
		userRepo:   userRepo,
		walletRepo: walletRepo,
		gradeRepo:  gradeRepo,
		jwtSecret:  jwtSecret,
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

	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	user, err := domain.NewUser(email, hash)
	if err != nil {
		return nil, err
	}

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

	return user, nil
}

func (uc *UseCase) Login(ctx context.Context, email, password string) (*domain.User, error) {
	if email == "" || password == "" {
		return nil, domain.NewInvalidInput("email and password are required")
	}

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, domain.NewInvalidInput("invalid email or password")
	}

	if user.Status != domain.UserStatusActive {
		return nil, domain.NewInvalidInput("account is blocked")
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return nil, domain.NewInvalidInput("invalid email or password")
	}

	return user, nil
}
