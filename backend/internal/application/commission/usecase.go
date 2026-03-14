package commission

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type UseCase struct {
	configRepo ports.CommissionConfigRepository
}

func NewUseCase(cr ports.CommissionConfigRepository) *UseCase {
	return &UseCase{configRepo: cr}
}

func (uc *UseCase) GetByKey(ctx context.Context, key string) (*domain.CommissionConfig, error) {
	cfg, err := uc.configRepo.GetByKey(ctx, key)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return cfg, nil
}

func (uc *UseCase) Update(ctx context.Context, cfg *domain.CommissionConfig) error {
	err := uc.configRepo.Update(ctx, cfg)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

func (uc *UseCase) ListAll(ctx context.Context) ([]*domain.CommissionConfig, error) {
	temp, err := uc.configRepo.ListAll(ctx)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	var list []*domain.CommissionConfig
	list = temp // явное присваивание

	return list, nil
}
