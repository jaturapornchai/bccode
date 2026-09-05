package services

import (
	"context"
	common "smlcloudplatform/internal/models"
	productmodels "smlcloudplatform/internal/product/product/models"
	"smlcloudplatform/internal/product/product/outbox"
	"smlcloudplatform/internal/product/productbarcode/models"
	unitconfig "smlcloudplatform/internal/product/unit/config"
	unitmodels "smlcloudplatform/internal/product/unit/models"
	unitservices "smlcloudplatform/internal/product/unit/services"
)

func barcodeProjectionMessages(topic string, doc models.ProductBarcodeDoc, createdProduct *productmodels.ProductDoc, createdUnit *unitmodels.UnitDoc) ([]outbox.Message, error) {
	key := outbox.BarcodeAggregateKey(doc.HoldingCode, doc.BusinessCode, doc.GuidFixed)
	messages := []outbox.Message{}
	if createdUnit != nil {
		message, err := outbox.NewMessage(unitconfig.MQ_TOPIC_CREATED, key, *createdUnit)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	if createdProduct != nil {
		message, err := outbox.NewMessage("when-product-created", key, *createdProduct)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	message, err := outbox.NewMessage(topic, key, doc)
	if err != nil {
		return nil, err
	}
	return append(messages, message), nil
}

// Company Barcode creation uses the same Unit rules, but all source documents
// and their delivery intent commit or roll back together.
func (svc ProductBarcodeHttpService) ensureBarcodeUnit(ctx context.Context, holdingCode, username string, doc *models.ProductBarcodeDoc) (*unitmodels.UnitDoc, error) {
	if doc.ItemUnitCode == "" {
		return nil, nil
	}
	unit, err := svc.repoUnit.FindByDocIndentityGuid(ctx, holdingCode, "unitcode", doc.ItemUnitCode)
	if err != nil {
		return nil, err
	}
	var created *unitmodels.UnitDoc
	if unit.UnitCode == "" {
		unit, err = unitservices.NewUnitDoc(holdingCode, username, unitmodels.Unit{
			UnitCode: doc.ItemUnitCode, Names: doc.ItemUnitNames,
			UnitName: common.UnitName{UnitName1: doc.ItemUnitCode},
		})
		if err != nil {
			return nil, err
		}
		if _, err := svc.repoUnit.Create(ctx, unit); err != nil {
			return nil, err
		}
		created = &unit
	}
	doc.ItemUnitGuid = unit.GuidFixed
	doc.ItemUnitNames = unit.Names
	return created, nil
}
