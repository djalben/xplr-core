package ports

import (
	"context"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
)

// TrustedDeviceRepository — хранение trusted-device токенов (в БД только SHA256 hex).
type TrustedDeviceRepository interface {
	Add(ctx context.Context, td *domain.TrustedDevice) error
	IsTrusted(ctx context.Context, userID domain.UUID, tokenHash string, now time.Time) (bool, error)
	TouchLastUsed(ctx context.Context, tokenHash string, now time.Time) error
	RevokeAll(ctx context.Context, userID domain.UUID) error
}
