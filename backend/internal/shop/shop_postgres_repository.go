package shop

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	organization "smlcloudplatform/internal/organization"
	"smlcloudplatform/internal/shop/models"
	micromodels "smlcloudplatform/pkg/microservice/models"
)

type ShopPostgresRepository struct {
	db *sql.DB
}

func NewShopPostgresRepository(db *sql.DB) IShopRepository {
	return &ShopPostgresRepository{db: db}
}

func (r *ShopPostgresRepository) Transaction(ctx context.Context, queryFunc func(context.Context) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := queryFunc(ctx); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ShopPostgresRepository) EnsureBootstrapIndexes(ctx context.Context) error {
	return nil
}

func (r *ShopPostgresRepository) Create(ctx context.Context, shop models.ShopDoc) (string, error) {
	query := `INSERT INTO holdings (code, name, tax_id, is_active, created_at)
	          VALUES ($1, $2, $3, $4, now())
	          ON CONFLICT (code) DO UPDATE
	          SET name = EXCLUDED.name,
	              tax_id = EXCLUDED.tax_id,
	              is_active = EXCLUDED.is_active`
	_, err := r.db.ExecContext(ctx, query,
		strings.TrimSpace(shop.HoldingCode),
		shop.Name1,
		shop.Settings.TaxID,
		shop.IsActive,
	)
	if err != nil {
		return "", err
	}
	return shop.HoldingCode, nil
}

func (r *ShopPostgresRepository) CreateCodeClaim(ctx context.Context, claim organization.OrganizationCodeClaim) error {
	return nil
}

func (r *ShopPostgresRepository) CreateAudit(ctx context.Context, audit organization.OrganizationAudit) error {
	return nil
}

func (r *ShopPostgresRepository) CreateOutbox(ctx context.Context, event organization.OrganizationOutboxEvent) error {
	return nil
}

func (r *ShopPostgresRepository) Update(ctx context.Context, guid string, expectedVersion int64, expectedActive bool, shop models.ShopDoc) error {
	query := `UPDATE holdings SET name = $2, tax_id = $3, is_active = $4 WHERE code = $1`
	_, err := r.db.ExecContext(ctx, query,
		shop.HoldingCode,
		shop.Name1,
		shop.Settings.TaxID,
		shop.IsActive,
	)
	return err
}

func (r *ShopPostgresRepository) FindByGuid(ctx context.Context, guid string) (models.ShopDoc, error) {
	return r.FindByHoldingCode(ctx, guid)
}

func (r *ShopPostgresRepository) FindByHoldingCode(ctx context.Context, holdingCode string) (models.ShopDoc, error) {
	holdingCode = strings.TrimSpace(holdingCode)
	var (
		code     string
		name     string
		taxID    sql.NullString
		isActive bool
	)

	query := `SELECT code, name, tax_id, is_active FROM holdings WHERE LOWER(code) = LOWER($1) LIMIT 1`
	err := r.db.QueryRowContext(ctx, query, holdingCode).Scan(&code, &name, &taxID, &isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.ShopDoc{}, mongo.ErrNoDocuments
		}
		return models.ShopDoc{}, err
	}

	doc := models.ShopDoc{
		HoldingUID: code,
		IsDeleted:  !isActive,
		ShopInfo: models.ShopInfo{
			Shop: models.Shop{
				HoldingCode: code,
				Name1:       name,
				IsActive:    isActive,
				Settings: models.ShopSettings{
					TaxID: taxID.String,
				},
			},
		},
	}
	objID, _ := primitive.ObjectIDFromHex(fmt.Sprintf("%024x", uuid.New().ID()))
	doc.ID = objID

	return doc, nil
}

func (r *ShopPostgresRepository) FindPage(ctx context.Context, pageable micromodels.Pageable) ([]models.ShopInfo, mongopagination.PaginationData, error) {
	query := `SELECT code, name, tax_id, is_active FROM holdings WHERE is_active = true ORDER BY code`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}
	defer rows.Close()

	var result []models.ShopInfo
	for rows.Next() {
		var (
			code     string
			name     string
			taxID    sql.NullString
			isActive bool
		)
		if err := rows.Scan(&code, &name, &taxID, &isActive); err != nil {
			continue
		}
		info := models.ShopInfo{
			Shop: models.Shop{
				HoldingCode: code,
				Name1:       name,
				IsActive:    isActive,
				Settings: models.ShopSettings{
					TaxID: taxID.String,
				},
			},
		}
		result = append(result, info)
	}

	pagination := mongopagination.PaginationData{
		Total:     int64(len(result)),
		Page:      1,
		PerPage:   int64(max(len(result), 1)),
		TotalPage: 1,
	}

	return result, pagination, nil
}

func (r *ShopPostgresRepository) Delete(ctx context.Context, guid string, username string) error {
	query := `UPDATE holdings SET is_active = false WHERE LOWER(code) = LOWER($1)`
	_, err := r.db.ExecContext(ctx, query, strings.TrimSpace(guid))
	return err
}
