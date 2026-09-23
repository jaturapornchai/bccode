package shop

import (
	"context"
	"database/sql"

	"smlcloudplatform/internal/authentication/models"
)

type ShopUserAccessLogPostgresRepository struct {
	db *sql.DB
}

func NewShopUserAccessLogPostgresRepository(db *sql.DB) IShopUserAccessLogRepository {
	return &ShopUserAccessLogPostgresRepository{db: db}
}

func (r *ShopUserAccessLogPostgresRepository) Create(ctx context.Context, shopUserAccessLog models.ShopUserAccessLog) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO shop_user_access_logs (holding_code, business_code, username, ip, last_accessed_at)
		VALUES ($1, $2, $3, $4, $5)`,
		shopUserAccessLog.HoldingCode, shopUserAccessLog.BusinessCode, shopUserAccessLog.Username, shopUserAccessLog.Ip, shopUserAccessLog.LastAccessedAt.UTC())
	return err
}
