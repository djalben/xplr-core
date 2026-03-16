package postgres

import (
	"context"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type gradeRepo struct {
	store *sqlx.DB
}

func NewGradeRepository(store *sqlx.DB) ports.GradeRepository {
	return &gradeRepo{store: store}
}

// GetByUserID — получение грейда.
func (r *gradeRepo) GetByUserID(ctx context.Context, userID domain.UUID) (*domain.UserGrade, error) {
	const query = `
		SELECT id, user_id, grade, total_spent, fee_percent, updated_at
		FROM user_grades 
		WHERE user_id = $1`

	var g domain.UserGrade

	err := r.store.GetContext(ctx, &g, query, userID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	return &g, nil
}

// Update — обновление грейда.
func (r *gradeRepo) Update(ctx context.Context, grade *domain.UserGrade) error {
	const query = `
		UPDATE user_grades 
		SET grade = $1, total_spent = $2, fee_percent = $3, updated_at = NOW()
		WHERE user_id = $4`

	_, err := r.store.ExecContext(ctx, query, grade.Grade, grade.TotalSpent, grade.FeePercent, grade.UserID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}
