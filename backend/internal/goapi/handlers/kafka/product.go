package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mypg"
	build "smlcloudplatform/internal/goapi/process/build"
	"smlcloudplatform/internal/utils"
)

// MongoProductModel — minimal projection of the mainapi ProductDoc payload
type MongoProductModel struct {
	HoldingCode  string `json:"holdingcode"`
	BusinessCode string `json:"businesscode"`
	Code         string `json:"code"`
	Names        []struct {
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
	p.HoldingCode = strings.TrimSpace(p.HoldingCode)
	p.BusinessCode = utils.NormalizeBusinessCode(p.BusinessCode)
	p.Code = utils.NormalizeBusinessCode(p.Code)
	if p.HoldingCode == "" || p.BusinessCode == "" || p.Code == "" {
		logger.Warn("Product message missing HoldingCode, BusinessCode or Code")
		return fmt.Errorf("missing HoldingCode, BusinessCode or Code for product upsert")
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
		INSERT INTO product (holding_code, businesscode, itemcode, name0, unitcode, unitname)
		VALUES ($1, $2, $3, $4, '', '')
		ON CONFLICT ON CONSTRAINT product_company_itemcode_unique
		DO UPDATE SET name0 = EXCLUDED.name0`,
		p.HoldingCode, p.BusinessCode, p.Code, name0)
	if err != nil {
		return fmt.Errorf("error upserting product %s: %v", p.Code, err)
	}
	logger.Info("Upserted product (PG): %s", p.Code)

	// Stock projection is intentionally not touched until stock rows also carry
	// BusinessCode. Recalculating by Holding+itemcode can mix two companies.
	return nil
}

// OnConsumeMessageProductDelete — when-product-deleted
func OnConsumeMessageProductDelete(msg string) error {
	var p MongoProductModel
	if err := json.Unmarshal([]byte(msg), &p); err != nil {
		logger.Error("unmarshaling product for deletion: %v", err)
		return err
	}
	p.HoldingCode = strings.TrimSpace(p.HoldingCode)
	p.BusinessCode = utils.NormalizeBusinessCode(p.BusinessCode)
	p.Code = utils.NormalizeBusinessCode(p.Code)
	if p.HoldingCode == "" || p.BusinessCode == "" || p.Code == "" {
		return fmt.Errorf("missing HoldingCode, BusinessCode or Code for product deletion")
	}

	build.DatabaseChecker(p.HoldingCode, false)

	db, err := mypg.PgSqlFastConnect(p.HoldingCode)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	result, err := db.ExecContext(
		context.Background(),
		"DELETE FROM product WHERE holding_code = $1 AND businesscode = $2 AND itemcode = $3",
		p.HoldingCode,
		p.BusinessCode,
		p.Code,
	)
	if err != nil {
		return fmt.Errorf("error deleting product %s: %v", p.Code, err)
	}
	rows, _ := result.RowsAffected()
	logger.Info("Deleted %d product row(s) for itemcode %s (PG)", rows, p.Code)
	return nil
}
