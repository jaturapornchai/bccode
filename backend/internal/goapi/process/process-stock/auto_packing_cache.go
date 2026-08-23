package processstock

import (
	"database/sql"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/mypg"
	"strings"
)

const packingQuerySingle = `SELECT DISTINCT
    unitname,
    barcoderefunitstand,
    barcoderefunitdivide,
    barcoderefunitstand / NULLIF(barcoderefunitdivide, 1) AS unitratio
FROM productbarcode
WHERE itemcode = $1
AND barcoderefunitstand > 0
AND barcoderefunitdivide > 0
ORDER BY unitratio DESC`

// BuildAutoPackingCache fetches packing info for a batch of item codes to minimize per-item queries.
func BuildAutoPackingCache(db *sql.DB, itemCodes []string) map[string][]models.ProductBarcodePackingStruct {
	return buildAutoPackingCache(db, "", "", itemCodes)
}

func BuildAutoPackingCacheCompany(db *sql.DB, holdingCode, businessCode string, itemCodes []string) map[string][]models.ProductBarcodePackingStruct {
	return buildAutoPackingCache(db, strings.TrimSpace(holdingCode), strings.ToUpper(strings.TrimSpace(businessCode)), itemCodes)
}

func buildAutoPackingCache(db *sql.DB, holdingCode, businessCode string, itemCodes []string) map[string][]models.ProductBarcodePackingStruct {
	cache := make(map[string][]models.ProductBarcodePackingStruct)
	if len(itemCodes) == 0 {
		return cache
	}

	placeholders := make([]string, len(itemCodes))
	args := make([]any, 0, len(itemCodes)+1)
	offset := 0
	companyWhere := ""
	if businessCode != "" {
		offset = 2
		args = append(args, holdingCode, businessCode)
		companyWhere = "holding_code = $1 AND businesscode = $2 AND "
	}
	for i, code := range itemCodes {
		placeholders[i] = fmt.Sprintf("$%d", offset+i+1)
		args = append(args, code)
	}

	batchQuery := `SELECT DISTINCT
        itemcode,
        unitname,
        barcoderefunitstand,
        barcoderefunitdivide,
        barcoderefunitstand / NULLIF(barcoderefunitdivide, 1) AS unitratio
    FROM productbarcode
    WHERE ` + companyWhere + `itemcode IN (` + strings.Join(placeholders, ",") + `)
    AND barcoderefunitstand > 0
    AND barcoderefunitdivide > 0
    ORDER BY itemcode, unitratio DESC`

	rows, err := mypg.QuerySelectAll(db, batchQuery, args...)
	if err != nil {
		logger.Error("querying auto packing batch: %v", err)
		return cache
	}

	for _, row := range rows {
		itemCode := mypg.GetStringValue(row, "itemcode")
		cache[itemCode] = append(cache[itemCode], models.ProductBarcodePackingStruct{
			UnitName:             mypg.GetStringValue(row, "unitname"),
			BarcodeRefUnitStand:  mypg.GetFloat64Value(row, "barcoderefunitstand"),
			BarcodeRefUnitDivide: mypg.GetFloat64Value(row, "barcoderefunitdivide"),
		})
	}
	return cache
}

// FetchPackingForItem returns cached packing rows for an item, querying on-demand if needed.
func FetchPackingForItem(db *sql.DB, cache map[string][]models.ProductBarcodePackingStruct, itemCode string) []models.ProductBarcodePackingStruct {
	if itemCode == "" {
		return nil
	}
	if cache != nil {
		if pack, ok := cache[itemCode]; ok {
			return pack
		}
	}

	rows, err := mypg.QuerySelectAll(db, packingQuerySingle, itemCode)
	if err != nil {
		logger.Error("querying auto packing: %v", err)
		return nil
	}

	packs := make([]models.ProductBarcodePackingStruct, 0, len(rows))
	for _, row := range rows {
		packs = append(packs, models.ProductBarcodePackingStruct{
			UnitName:             mypg.GetStringValue(row, "unitname"),
			BarcodeRefUnitStand:  mypg.GetFloat64Value(row, "barcoderefunitstand"),
			BarcodeRefUnitDivide: mypg.GetFloat64Value(row, "barcoderefunitdivide"),
		})
	}

	if len(packs) > 0 {
		cache[itemCode] = packs
	}
	return packs
}
