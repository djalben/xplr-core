package grades_test

import (
	"context"
	"errors"
	"testing"

	"github.com/djalben/xplr-core/backend/internal/application/card/mocks"
	"github.com/djalben/xplr-core/backend/internal/application/grades"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

func TestUseCase_ChangeGrade_OK_STANDARD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mocks.NewMockGradeRepository(ctrl)
	uc := grades.NewUseCase(repo)

	g := &domain.UserGrade{UserID: uid, Grade: domain.UserGradeGold, FeePercent: domain.NewNumeric(4.5)}
	repo.EXPECT().GetByUserID(gomock.Any(), uid).Return(g, nil)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, upd *domain.UserGrade) error {
		if upd.Grade != domain.UserGradeStandard {
			t.Fatalf("grade: %s", upd.Grade)
		}

		return nil
	})

	err := uc.ChangeGrade(ctx, uid, "standard")
	if err != nil {
		t.Fatal(err)
	}
}

func TestUseCase_ChangeGrade_OK_GOLD_normalized(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uid := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mocks.NewMockGradeRepository(ctrl)
	uc := grades.NewUseCase(repo)

	g := &domain.UserGrade{UserID: uid, Grade: domain.UserGradeStandard}
	repo.EXPECT().GetByUserID(gomock.Any(), uid).Return(g, nil)
	repo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, upd *domain.UserGrade) error {
		if upd.Grade != domain.UserGradeGold {
			t.Fatalf("grade: %s", upd.Grade)
		}

		return nil
	})

	err := uc.ChangeGrade(ctx, uid, "  gold ")
	if err != nil {
		t.Fatal(err)
	}
}

func TestUseCase_ChangeGrade_Invalid(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uid := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mocks.NewMockGradeRepository(ctrl)
	uc := grades.NewUseCase(repo)

	g := &domain.UserGrade{UserID: uid, Grade: domain.UserGradeStandard}
	repo.EXPECT().GetByUserID(gomock.Any(), uid).Return(g, nil)

	err := uc.ChangeGrade(ctx, uid, "SILVER")
	if err == nil || !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
}
