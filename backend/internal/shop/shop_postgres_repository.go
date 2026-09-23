package shop

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	orgaccess "smlcloudplatform/internal/organization"
	"smlcloudplatform/internal/shop/models"
)

var errHoldingNotFound = errors.New("holding not found")

type ShopPostgresRepository struct {
	db *sql.DB
}

func NewShopPostgresRepository(db *sql.DB) IShopRepository {
	return &ShopPostgresRepository{db: db}
}

func (r *ShopPostgresRepository) Transaction(ctx context.Context, fn func(context.Context) error) error {
	return runInTx(ctx, r.db, fn)
}

// holdingName is the relational display name: the Thai/first localized name, then name1.
func holdingName(shop models.Shop) string {
	if name := orgaccess.PrimaryName(shop.Names); name != "" {
		return name
	}
	return strings.TrimSpace(shop.Name1)
}

func holdingProfile(shop models.Shop) (string, error) {
	raw, err := json.Marshal(shop)
	return string(raw), err
}

func (r *ShopPostgresRepository) Create(ctx context.Context, shop models.ShopDoc) error {
	profile, err := holdingProfile(shop.Shop)
	if err != nil {
		return err
	}
	_, err = conn(ctx, r.db).ExecContext(ctx, `
		INSERT INTO holdings (code, name, tax_id, profile, is_active, created_by, created_at, updated_by, updated_at)
		VALUES ($1, $2, $3, $4, true, $5, $6, '', $6)`,
		shop.HoldingCode, holdingName(shop.Shop), strings.TrimSpace(shop.Settings.TaxID), profile, shop.CreatedBy, shop.CreatedAt.UTC())
	return err
}

func (r *ShopPostgresRepository) Update(ctx context.Context, holdingCode string, expectedActive bool, shop models.ShopDoc) error {
	profile, err := holdingProfile(shop.Shop)
	if err != nil {
		return err
	}
	result, err := conn(ctx, r.db).ExecContext(ctx, `
		UPDATE holdings SET name = $2, tax_id = $3, profile = $4, updated_by = $5, updated_at = $6
		WHERE code = $1 AND is_active = $7`,
		holdingCode, holdingName(shop.Shop), strings.TrimSpace(shop.Settings.TaxID), profile, shop.UpdatedBy, shop.UpdatedAt.UTC(), expectedActive)
	return requireOneRow(result, err)
}

func (r *ShopPostgresRepository) UpdateStatus(ctx context.Context, holdingCode string, expectedActive bool, active bool, username string, now time.Time) error {
	result, err := conn(ctx, r.db).ExecContext(ctx, `
		UPDATE holdings SET is_active = $2, updated_by = $3, updated_at = $4
		WHERE code = $1 AND is_active = $5`,
		holdingCode, active, username, now.UTC(), expectedActive)
	return requireOneRow(result, err)
}

func requireOneRow(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil || affected != 1 {
		return orgaccess.ErrStatusChangeConflict
	}
	return nil
}

func (r *ShopPostgresRepository) FindByHoldingCode(ctx context.Context, holdingCode string) (models.ShopDoc, error) {
	holdingCode = strings.TrimSpace(holdingCode)
	var (
		doc     models.ShopDoc
		name    string
		taxID   string
		profile []byte
	)
	err := conn(ctx, r.db).QueryRowContext(ctx, `
		SELECT code, name, COALESCE(tax_id, ''), profile, is_active, created_by, created_at, updated_by, updated_at
		FROM holdings WHERE code = $1`, holdingCode).
		Scan(&doc.HoldingCode, &name, &taxID, &profile, &doc.IsActive, &doc.CreatedBy, &doc.CreatedAt, &doc.UpdatedBy, &doc.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.ShopDoc{}, errHoldingNotFound
	}
	if err != nil {
		return models.ShopDoc{}, err
	}
	code, active := doc.HoldingCode, doc.IsActive
	if len(profile) > 0 {
		if err := json.Unmarshal(profile, &doc.Shop); err != nil {
			return models.ShopDoc{}, err
		}
	}
	// Relational columns are the source of truth for identity, status, name and tax id.
	doc.HoldingCode, doc.IsActive = code, active
	if strings.TrimSpace(doc.Name1) == "" {
		doc.Name1 = name
	}
	if len(doc.Names) == 0 && name != "" {
		doc.Names = orgaccess.DecodeNames(nil, name)
	}
	if strings.TrimSpace(doc.Settings.TaxID) == "" {
		doc.Settings.TaxID = taxID
	}
	doc.GuidFixed = code
	doc.HoldingUID = code
	return doc, nil
}

func (r *ShopPostgresRepository) RecordAudit(ctx context.Context, audit orgaccess.Audit) error {
	return orgaccess.RecordAudit(ctx, conn(ctx, r.db), audit)
}
