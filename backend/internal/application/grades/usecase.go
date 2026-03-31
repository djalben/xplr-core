package grades

import (
	"context"
	"strings"

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

// GetByUserID — возвращает грейд пользователя.
func (uc *UseCase) GetByUserID(ctx context.Context, userID domain.UUID) (*domain.UserGrade, error) {
	g, err := uc.gradeRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return g, nil
}

// ChangeGrade — меняет грейд пользователя.
func (uc *UseCase) ChangeGrade(ctx context.Context, userID domain.UUID, grade string) error {
	g, err := uc.gradeRepo.GetByUserID(ctx, userID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	switch strings.ToUpper(strings.TrimSpace(grade)) {
	case domain.UserGradeStandard, domain.UserGradeGold:
		grade = strings.ToUpper(strings.TrimSpace(grade))
	default:
		return domain.NewInvalidInput("invalid grade: only STANDARD and GOLD are allowed")
	}

	g.Grade = grade
	// Можно обновить fee_percent по грейду из commission_config, если нужно

	err = uc.gradeRepo.Update(ctx, g)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}
