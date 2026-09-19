package shop

import (
	"context"
	"database/sql"
	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/pkg/microservice"
)

type IShopUserAccessLogRepository interface {
	Create(ctx context.Context, shopUserAccessLog models.ShopUserAccessLog) error
}

type ShopUserAccessLogRepository struct {
	pst microservice.IPersisterMongo
}

func NewShopUserAccessLogRepository(pst microservice.IPersisterMongo) IShopUserAccessLogRepository {
	return ShopUserAccessLogRepository{
		pst: pst,
	}
}

func (svc ShopUserAccessLogRepository) Create(ctx context.Context, shopUserAccessLog models.ShopUserAccessLog) error {
	if svc.pst == nil {
		return nil
	}
	_, err := svc.pst.Create(ctx, &models.ShopUserAccessLog{}, shopUserAccessLog)
	if err != nil {
		return err
	}
	return nil
}

type ShopUserAccessLogPostgresRepository struct {
	db *sql.DB
}

func NewShopUserAccessLogPostgresRepository(db *sql.DB) IShopUserAccessLogRepository {
	return &ShopUserAccessLogPostgresRepository{db: db}
}

func (r *ShopUserAccessLogPostgresRepository) Create(ctx context.Context, shopUserAccessLog models.ShopUserAccessLog) error {
	if r.db == nil {
		return nil
	}
	query := `INSERT INTO shop_user_access_logs (holding_code, username, ip, last_accessed_at)
	          VALUES ($1, $2, $3, $4)`
	_, _ = r.db.ExecContext(ctx, query, shopUserAccessLog.HoldingCode, shopUserAccessLog.Username, shopUserAccessLog.Ip, shopUserAccessLog.LastAccessedAt)
	return nil
}

