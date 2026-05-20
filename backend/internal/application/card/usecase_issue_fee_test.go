package card

import (
	"context"
	"database/sql"
	"testing"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type stubCommissionRepo struct {
	getByKey func(ctx context.Context, key string) (*domain.CommissionConfig, error)
}

func (s *stubCommissionRepo) GetByKey(ctx context.Context, key string) (*domain.CommissionConfig, error) {
	return s.getByKey(ctx, key)
}

func (s *stubCommissionRepo) Update(context.Context, *domain.CommissionConfig) error {
	return nil
}

func (s *stubCommissionRepo) ListAll(context.Context) ([]*domain.CommissionConfig, error) {
	return nil, nil
}

func TestUseCase_cardIssueFee_missingKeyUsesDefault(t *testing.T) {
	t.Parallel()

	uc := NewUseCase(nil, nil, nil, nil, &stubCommissionRepo{
		getByKey: func(ctx context.Context, key string) (*domain.CommissionConfig, error) {
			return nil, wrapper.Wrap(sql.ErrNoRows)
		},
	})

	fee, err := uc.cardIssueFee(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	want := domain.NewNumeric(2)
	if !fee.Equal(want) {
		t.Fatalf("got %s want %s", fee, want)
	}
}

func TestUseCase_cardIssueFee_nilRepoUsesDefault(t *testing.T) {
	t.Parallel()

	uc := NewUseCase(nil, nil, nil, nil, nil)

	fee, err := uc.cardIssueFee(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	want := domain.NewNumeric(2)
	if !fee.Equal(want) {
		t.Fatalf("got %s want %s", fee, want)
	}
}
