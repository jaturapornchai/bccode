package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mypg"
	build "smlcloudplatform/internal/goapi/process/build"
	processstock "smlcloudplatform/internal/goapi/process/process-stock"
	"smlcloudplatform/internal/utils"
)

// MongoProductModel — minimal projection of the mainapi ProductDoc payload
type MongoProductModel struct {
	HoldingCode string `json:"holdingcode"`
	Code        string `json:"code"`
	Names       []struct {
		Name *string `json:"name"`
	} `json:"names"`
}

// OnConsumeMessageProductCreateOrUpdate — when-product-created / when-product-updated
func OnConsumeMessageProductCreateOrUpdate(msg string) error {
	var p MongoProductModel
	if err := json.Unmarshal([]byte(msg), &p); err != nil {
		logger.Error("unmarshaling product: %v", err)
		return err
	}
	p.Code = utils.NormalizeBusinessCode(p.Code)
	if p.HoldingCode == "" || p.Code == "" {
		logger.Warn("Product message missing HoldingCode or Code")
		return fmt.Errorf("missing HoldingCode or Code for product upsert")
	}

	build.DatabaseChecker(p.HoldingCode, false)

	db, err := mypg.PgSqlFastConnect(p.HoldingCode)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	name0 := p.Code // name0 is NOT NULL — fall back to itemcode
	if len(p.Names) > 0 && p.Names[0].Name != nil && *p.Names[0].Name != "" {
		name0 = *p.Names[0].Name
	}

	// Metadata-only upsert. Stock columns (balanceqty*, pending*) belong to
	// processstock/batchUpdateProduct — never written here.
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO product (itemcode, name0, unitcode, unitname)
		VALUES ($1, $2, '', '')
		ON CONFLICT ON CONSTRAINT product_itemcode_unique
		DO UPDATE SET name0 = EXCLUDED.name0`,
		p.Code, name0)
	if err != nil {
		return fmt.Errorf("error upserting product %s: %v", p.Code, err)
	}
	logger.Info("Upserted product (PG): %s", p.Code)

	// Recalc balance columns (same call ProductBarcodeBuild makes) — makes
	// resync self-healing for tenants that already have stock transactions.
	processstock.ProcessProductBalanceUpdateByItemsAsync(db, []string{p.Code})
	return nil
}

// OnConsumeMessageProductDelete — when-product-deleted
func OnConsumeMessageProductDelete(msg string) error {
	var p MongoProductModel
	if err := json.Unmarshal([]byte(msg), &p); err != nil {
		logger.Error("unmarshaling product for deletion: %v", err)
		return err
	}
	p.Code = utils.NormalizeBusinessCode(p.Code)
	if p.HoldingCode == "" || p.Code == "" {
		return fmt.Errorf("missing HoldingCode or Code for product deletion")
	}

	build.DatabaseChecker(p.HoldingCode, false)

	db, err := mypg.PgSqlFastConnect(p.HoldingCode)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	result, err := db.ExecContext(context.Background(), "DELETE FROM product WHERE itemcode = $1", p.Code)
	if err != nil {
		return fmt.Errorf("error deleting product %s: %v", p.Code, err)
	}
	rows, _ := result.RowsAffected()
	logger.Info("Deleted %d product row(s) for itemcode %s (PG)", rows, p.Code)
	return nil
}
