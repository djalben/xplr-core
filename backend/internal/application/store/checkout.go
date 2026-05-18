package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/google/uuid"
	"gitlab.com/libs-artifex/wrapper/v2"
)

const transactionTypeStorePurchase = "STORE_PURCHASE"

type orderFulfillment struct {
	ActivationKey string
	QRData        string
	ProviderRef   string
	Meta          map[string]any
	Status        domain.StoreOrderStatus
}

func (uc *UseCase) checkoutFromWallet(
	ctx context.Context,
	userID domain.UUID,
	price domain.Numeric,
	productName string,
	productID *domain.UUID,
	fulfill func(ctx context.Context) (*orderFulfillment, error),
) (*PurchaseResult, error) {
	w, err := uc.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	if w.Balance.LessThan(price) {
		return nil, domain.NewInsufficientFunds()
	}

	ff, err := fulfill(ctx)
	if err != nil {
		return nil, err
	}

	status := ff.Status
	if status == "" {
		status = domain.StoreOrderStatusCompleted
	}

	metaBytes, err := json.Marshal(ff.Meta)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	oid := uuid.New()
	o := &domain.StoreOrder{
		ID:            oid,
		UserID:        userID,
		ProductID:     productID,
		ProductName:   productName,
		PriceUSD:      price,
		Status:        status,
		ActivationKey: ff.ActivationKey,
		QRData:        ff.QRData,
		ProviderRef:   ff.ProviderRef,
		Meta:          string(metaBytes),
		CreatedAt:     time.Now().UTC(),
	}

	err = w.Withdraw(price)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	err = uc.walletRepo.Update(ctx, w)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	err = uc.storeRepo.CreateOrder(ctx, o)
	if err != nil {
		createErr := wrapper.Wrap(err)

		refundErr := w.TopUp(price)
		if refundErr == nil {
			refundErr = uc.walletRepo.Update(ctx, w)
		}
		if refundErr != nil {
			return nil, wrapper.Wrap(errors.Join(createErr, refundErr))
		}

		return nil, createErr
	}

	err = uc.recordStorePurchaseTx(ctx, userID, price, productName)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	statusStr := "completed"
	if status == domain.StoreOrderStatusPending {
		statusStr = "pending"
	}

	return &PurchaseResult{
		OrderID:       oid,
		ProductName:   productName,
		PriceUSD:      price.StringFixed(2),
		ActivationKey: ff.ActivationKey,
		QRData:        ff.QRData,
		Status:        statusStr,
		ProviderRef:   ff.ProviderRef,
	}, nil
}

func (uc *UseCase) recordStorePurchaseTx(ctx context.Context, userID domain.UUID, amount domain.Numeric, details string) error {
	if uc.txRepo == nil {
		return nil
	}

	tx := domain.NewTransaction(
		userID,
		nil,
		amount,
		domain.NewNumeric(0),
		transactionTypeStorePurchase,
		"COMPLETED",
		details,
	)

	return uc.txRepo.Save(ctx, tx)
}
