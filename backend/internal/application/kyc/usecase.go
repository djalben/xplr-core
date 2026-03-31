package kyc

import (
	"context"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// UseCase — заявки KYC (пользователь и админ).
type UseCase struct {
	kycRepo  ports.KYCApplicationRepository
	userRepo ports.UserRepository
}

func NewUseCase(kycRepo ports.KYCApplicationRepository, userRepo ports.UserRepository) *UseCase {
	return &UseCase{kycRepo: kycRepo, userRepo: userRepo}
}

// SubmitApplication — пользователь подаёт заявку (JSON payload).
func (uc *UseCase) SubmitApplication(ctx context.Context, userID domain.UUID, payloadJSON string) error {
	if payloadJSON == "" {
		payloadJSON = "{}"
	}

	pending, err := uc.kycRepo.HasPendingForUser(ctx, userID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	if pending {
		return domain.NewInvalidInput("kyc application already pending")
	}

	app := domain.NewKYCApplication(userID, payloadJSON)

	return uc.kycRepo.Save(ctx, app)
}

// ListPending — заявки в статусе PENDING (админка).
func (uc *UseCase) ListPending(ctx context.Context, limit int) ([]*domain.KYCApplication, error) {
	return uc.kycRepo.ListByStatus(ctx, domain.KYCApplicationPending, limit)
}

// DecideApplication — одобрение или отклонение заявки.
func (uc *UseCase) DecideApplication(ctx context.Context, applicationID, adminID domain.UUID, approve bool, comment string) error {
	app, err := uc.kycRepo.GetByID(ctx, applicationID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	if app.Status != domain.KYCApplicationPending {
		return domain.NewInvalidInput("application is not pending")
	}

	user, err := uc.userRepo.GetByID(ctx, app.UserID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	now := time.Now().UTC()
	app.ReviewedBy = &adminID
	app.ReviewedAt = &now

	if comment != "" {
		app.AdminComment = &comment
	}

	if approve {
		app.Status = domain.KYCApplicationApproved
		user.KYCStatus = domain.KYCApproved
	} else {
		app.Status = domain.KYCApplicationRejected
		user.KYCStatus = domain.KYCRejected
	}

	err = uc.userRepo.Update(ctx, user)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return uc.kycRepo.Update(ctx, app)
}
