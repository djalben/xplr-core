package grades

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type UseCase struct {
	gradeRepo ports.GradeRepository
}

func NewUseCase(gr ports.GradeRepository) *UseCase {
	return &UseCase{gradeRepo: gr}
}

// ChangeGrade — меняет грейд пользователя.
func (uc *UseCase) ChangeGrade(ctx context.Context, userID domain.UUID, grade string) error {
	g, err := uc.gradeRepo.GetByUserID(ctx, userID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	// Проверка на допустимый grade
	switch grade {
	case "STANDARD", "SILVER", "GOLD", "PLATINUM", "BLACK":
		// OK
	default:
		return domain.NewInvalidInput("invalid grade")
	}

	g.Grade = grade
	// Можно обновить fee_percent по грейду из commission_config, если нужно

	err = uc.gradeRepo.Update(ctx, g)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}