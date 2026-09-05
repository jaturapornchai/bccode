package services

import (
	"errors"
	"smlcloudplatform/internal/product/unit/models"
	"smlcloudplatform/internal/utils"
	"time"
)

// NewUnitDoc applies the existing Unit creation rules without side effects, so
// callers can persist the document and delivery intent in their own transaction.
func NewUnitDoc(holdingCode, username string, unit models.Unit) (models.UnitDoc, error) {
	unit.UnitCode = utils.NormalizeBusinessCode(unit.UnitCode)
	if unit.UnitCode == "" {
		return models.UnitDoc{}, errors.New("unit code is required")
	}
	doc := models.UnitDoc{}
	doc.HoldingCode = holdingCode
	doc.GuidFixed = utils.NewGUID()
	doc.Unit = unit
	syncUnitNames(&doc.Unit)
	doc.CreatedBy = username
	doc.CreatedAt = time.Now()
	return doc, nil
}
