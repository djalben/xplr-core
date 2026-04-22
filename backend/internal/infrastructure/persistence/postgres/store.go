package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type storeRepo struct {
	db *sqlx.DB
}

func NewStoreRepository(db *sqlx.DB) ports.StoreRepository {
	return &storeRepo{db: db}
}

func (r *storeRepo) ListCategories(ctx context.Context) ([]*domain.StoreCategory, error) {
	const q = `SELECT id,
       slug,
       name,
       COALESCE(description, '') AS description,
       COALESCE(icon, '') AS icon,
       COALESCE(image_url, '') AS image_url,
       sort_order,
       created_at
FROM store_categories ORDER BY sort_order, slug`
	var out []*domain.StoreCategory
	err := r.db.SelectContext(ctx, &out, q)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	if out == nil {
		out = []*domain.StoreCategory{}
	}

	return out, nil
}

func (r *storeRepo) ListProducts(ctx context.Context, filter ports.StoreProductFilter) ([]*domain.StoreProduct, error) {
	q := `SELECT p.id,
       p.category_id,
       c.slug AS category_slug,
       p.provider,
       p.external_id,
       p.name,
       COALESCE(p.description, '') AS description,
       COALESCE(p.country, '') AS country,
       COALESCE(p.country_code, '') AS country_code,
       p.product_type,
       p.price_usd,
       p.cost_price,
       p.markup_percent,
       COALESCE(p.data_gb, '') AS data_gb,
       p.validity_days,
       COALESCE(p.image_url, '') AS image_url,
       p.in_stock,
       COALESCE(p.meta, '{}') AS meta,
       p.sort_order
FROM store_products p JOIN store_categories c ON c.id = p.category_id
WHERE p.in_stock = TRUE`

	args := []any{}
	arg := 1

	if filter.CategorySlug != "" {
		q += fmt.Sprintf(" AND c.slug = $%d", arg)
		args = append(args, filter.CategorySlug)
		arg++
	}
	if filter.Country != "" {
		q += fmt.Sprintf(" AND (LOWER(p.country) LIKE LOWER($%d) OR LOWER(p.country_code) = LOWER($%d))", arg, arg+1)
		args = append(args, "%"+filter.Country+"%", filter.Country)
		// arg is only used for further placeholders; keep in sync with appended args.
		arg += 2
	}
	if filter.Search != "" {
		q += fmt.Sprintf(" AND (LOWER(p.name) LIKE LOWER($%d) OR LOWER(p.country) LIKE LOWER($%d))", arg, arg+1)
		s := "%" + strings.ToLower(filter.Search) + "%"
		args = append(args, s, s)
	}
	q += " ORDER BY p.sort_order, p.name"

	var out []*domain.StoreProduct
	err := r.db.SelectContext(ctx, &out, q, args...)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	if out == nil {
		out = []*domain.StoreProduct{}
	}

	return out, nil
}

func (r *storeRepo) AdminListProducts(ctx context.Context, filter ports.StoreAdminProductFilter) ([]*domain.StoreProduct, error) {
	q := `SELECT p.id,
       p.category_id,
       c.slug AS category_slug,
       p.provider,
       p.external_id,
       p.name,
       COALESCE(p.description, '') AS description,
       COALESCE(p.country, '') AS country,
       COALESCE(p.country_code, '') AS country_code,
       p.product_type,
       p.price_usd,
       p.cost_price,
       p.markup_percent,
       COALESCE(p.data_gb, '') AS data_gb,
       p.validity_days,
       COALESCE(p.image_url, '') AS image_url,
       p.in_stock,
       COALESCE(p.meta, '{}') AS meta,
       p.sort_order
FROM store_products p JOIN store_categories c ON c.id = p.category_id
WHERE 1=1`

	args := []any{}
	arg := 1

	if filter.ProductType != nil && *filter.ProductType != "" {
		q += fmt.Sprintf(" AND p.product_type = $%d", arg)
		args = append(args, string(*filter.ProductType))
		arg++
	}

	q += " ORDER BY p.sort_order, p.name"

	var out []*domain.StoreProduct
	err := r.db.SelectContext(ctx, &out, q, args...)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	if out == nil {
		out = []*domain.StoreProduct{}
	}

	return out, nil
}

func (r *storeRepo) AdminUpdateProduct(ctx context.Context, p *domain.StoreProduct) error {
	const q = `UPDATE store_products
SET price_usd = $1, cost_price = $2, markup_percent = $3, image_url = $4, in_stock = $5, meta = $6, sort_order = $7, updated_at = NOW()
WHERE id = $8`

	_, err := r.db.ExecContext(ctx, q, p.PriceUSD, p.CostPrice, p.MarkupPct, p.ImageURL, p.InStock, p.Meta, p.SortOrder, p.ID)
	if err != nil {
		return wrapper.Wrap(err)
	}

	return nil
}

func (r *storeRepo) AdminBulkAddMarkup(ctx context.Context, productType domain.StoreProductType, delta domain.Numeric) (int64, error) {
	const q = `UPDATE store_products
SET markup_percent = markup_percent + $1, updated_at = NOW()
WHERE product_type = $2`

	res, err := r.db.ExecContext(ctx, q, delta, string(productType))
	if err != nil {
		return 0, wrapper.Wrap(err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return 0, wrapper.Wrap(err)
	}

	return n, nil
}

func (r *storeRepo) GetProductByID(ctx context.Context, id domain.UUID) (*domain.StoreProduct, error) {
	const q = `SELECT p.id,
       p.category_id,
       c.slug AS category_slug,
       p.provider,
       p.external_id,
       p.name,
       COALESCE(p.description, '') AS description,
       COALESCE(p.country, '') AS country,
       COALESCE(p.country_code, '') AS country_code,
       p.product_type,
       p.price_usd,
       p.cost_price,
       p.markup_percent,
       COALESCE(p.data_gb, '') AS data_gb,
       p.validity_days,
       COALESCE(p.image_url, '') AS image_url,
       p.in_stock,
       COALESCE(p.meta, '{}') AS meta,
       p.sort_order
FROM store_products p JOIN store_categories c ON c.id = p.category_id
WHERE p.id = $1`
	var p domain.StoreProduct
	err := r.db.GetContext(ctx, &p, q, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, wrapper.Wrap(domain.NewNotFound("product not found"))
		}

		return nil, wrapper.Wrap(err)
	}

	return &p, nil
}

func (r *storeRepo) CreateOrder(ctx context.Context, o *domain.StoreOrder) error {
	const q = `INSERT INTO store_orders
(id, user_id, product_id, product_name, price_usd, status, activation_key, qr_data, provider_ref, meta, created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err := r.db.ExecContext(ctx, q,
		o.ID, o.UserID, o.ProductID, o.ProductName, o.PriceUSD, string(o.Status),
		o.ActivationKey, o.QRData, o.ProviderRef, o.Meta, o.CreatedAt,
	)

	return wrapper.Wrap(err)
}

func (r *storeRepo) ListOrdersByUser(ctx context.Context, userID domain.UUID, limit int) ([]*domain.StoreOrder, error) {
	const q = `SELECT id,
       user_id,
       product_id,
       product_name,
       price_usd,
       status,
       COALESCE(activation_key, '') AS activation_key,
       COALESCE(qr_data, '') AS qr_data,
       COALESCE(provider_ref, '') AS provider_ref,
       COALESCE(meta, '{}') AS meta,
       created_at
FROM store_orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`
	var out []*domain.StoreOrder
	err := r.db.SelectContext(ctx, &out, q, userID, limit)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}
	if out == nil {
		out = []*domain.StoreOrder{}
	}

	return out, nil
}

func (r *storeRepo) GetLatestCompletedOrderByProviderRef(ctx context.Context, providerRef string) (*domain.StoreOrder, error) {
	const q = `SELECT id,
       user_id,
       product_id,
       product_name,
       price_usd,
       status,
       COALESCE(activation_key, '') AS activation_key,
       COALESCE(qr_data, '') AS qr_data,
       COALESCE(provider_ref, '') AS provider_ref,
       COALESCE(meta, '{}') AS meta,
       created_at
FROM store_orders WHERE provider_ref = $1 AND status = 'COMPLETED' ORDER BY created_at DESC LIMIT 1`
	var o domain.StoreOrder
	err := r.db.GetContext(ctx, &o, q, providerRef)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, wrapper.Wrap(domain.NewNotFound("order not found"))
		}

		return nil, wrapper.Wrap(err)
	}

	return &o, nil
}

func (r *storeRepo) GetLatestCompletedOrderMetaByProviderRef(ctx context.Context, providerRef string, userID *domain.UUID) (string, error) {
	q := `SELECT COALESCE(meta,'{}') FROM store_orders WHERE provider_ref = $1 AND status = 'COMPLETED'`
	args := []any{providerRef}
	if userID != nil {
		q += " AND user_id = $2"
		args = append(args, *userID)
	}
	q += " ORDER BY created_at DESC LIMIT 1"
	var meta string
	err := r.db.GetContext(ctx, &meta, q, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", wrapper.Wrap(domain.NewNotFound("order not found"))
		}

		return "", wrapper.Wrap(err)
	}

	return meta, nil
}

func (r *storeRepo) SoftDeleteOrdersByProviderRef(ctx context.Context, providerRef string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE store_orders SET status = 'DELETED' WHERE provider_ref = $1 AND status = 'COMPLETED'`, providerRef)

	return wrapper.Wrap(err)
}

func (r *storeRepo) UpdateOrderMetaByProviderRef(ctx context.Context, providerRef string, meta string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE store_orders SET meta = $1 WHERE provider_ref = $2 AND status = 'COMPLETED'`, meta, providerRef)

	return wrapper.Wrap(err)
}
