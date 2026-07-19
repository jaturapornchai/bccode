package services

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/product/productbarcode/models"
)

func (svc ProductBarcodeHttpService) InfoBomView(holdingCode string, itemCode string, barcode string) (models.ProductBarcodeBOMView, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	doc, err := svc.findBarcodeByBusinessKey(ctx, holdingCode, itemCode, barcode)

	if err != nil {
		return models.ProductBarcodeBOMView{}, err
	}

	if len(doc.ProductBarcode.Barcode) == 0 {
		return models.ProductBarcodeBOMView{}, fmt.Errorf("barcode not found")
	}

	bomViewDict := map[string]*models.ProductBarcodeBOMView{}
	bomView := models.ProductBarcodeBOMView{}

	// BuildBOMViewCache(ctx, svc.repo.FindByBarcode,
	// 	0, &map[string]models.ProductBarcodeDoc{},
	// 	&bomViewDict,
	// 	holdingCode, doc.Barcode, []models.BOMProductBarcode{}, &bomView)

	bomView.FromProductBarcode(doc.ProductBarcodeData)

	rootKey := productBarcodeBusinessKey(doc.ItemCode, doc.Barcode)
	if _, ok := bomViewDict[rootKey]; !ok {
		bomViewDict[rootKey] = &bomView
	}

	bomView.Level = 1

	if doc.BOM != nil && len(*doc.BOM) > 0 {
		productBarcodeDict := map[string]models.ProductBarcodeDoc{}
		err = BuildBOMView(ctx, svc.findBarcodeByBusinessKey, bomView.Level, &productBarcodeDict, &bomViewDict, holdingCode, doc.BOM, &bomView.BOM)
		if err != nil {
			return models.ProductBarcodeBOMView{}, err
		}
	}

	return bomView, nil
}

func (svc ProductBarcodeHttpService) ListBomView(holdingCode string, barcodes []string) ([]models.ProductBarcodeBOMView, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	var bomViews []models.ProductBarcodeBOMView
	var bomViewDict = map[string]*models.ProductBarcodeBOMView{}
	for _, barcode := range barcodes {
		doc, err := svc.repo.FindByBarcode(ctx, holdingCode, barcode)

		if err != nil {
			return []models.ProductBarcodeBOMView{}, err
		}

		if len(doc.ProductBarcode.Barcode) == 0 {
			return []models.ProductBarcodeBOMView{}, fmt.Errorf("barcode not found")
		}

		bomView := models.ProductBarcodeBOMView{}
		bomView.FromProductBarcode(doc.ProductBarcodeData)

		docKey := productBarcodeBusinessKey(doc.ItemCode, doc.Barcode)
		if _, ok := bomViewDict[docKey]; !ok {
			bomViewDict[docKey] = &bomView
		}

		if doc.BOM != nil && len(*doc.BOM) > 0 {
			productBarcodeDict := map[string]models.ProductBarcodeDoc{}
			err = BuildBOMView(ctx, svc.findBarcodeByBusinessKey, 1, &productBarcodeDict, &bomViewDict, holdingCode, doc.BOM, &bomView.BOM)
			if err != nil {
				return []models.ProductBarcodeBOMView{}, err
			}
		}

		bomViews = append(bomViews, bomView)
	}

	return bomViews, nil
}

func (svc ProductBarcodeHttpService) BuildBOM() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func BuildBOMView(
	ctx context.Context,
	findByBarcode func(ctx context.Context, holdingCode string, itemCode string, barcode string) (models.ProductBarcodeDoc, error),
	currentLevel int,
	productBarcodeDict *map[string]models.ProductBarcodeDoc,
	bomViewDict *map[string]*models.ProductBarcodeBOMView,
	holdingCode string,
	BOMs *[]models.BOMProductBarcode,
	bomView *[]models.ProductBarcodeBOMView) error {

	if currentLevel > 10 {
		return fmt.Errorf("BOM level is too deep")
	}

	currentLevel += 1

	for _, bom := range *BOMs {
		bomKey := productBarcodeBusinessKey(bom.ItemCode, bom.Barcode)

		tempBOMView := models.ProductBarcodeBOMView{}
		tempBOMView.Level = currentLevel

		if _, bomOk := (*bomViewDict)[bomKey]; bomOk {
			tempBOMView = *(*bomViewDict)[bomKey]
		} else {
			var tempDoc = models.ProductBarcodeDoc{}
			if _, ok := (*productBarcodeDict)[bomKey]; !ok {
				findDoc, err := findByBarcode(ctx, holdingCode, bom.ItemCode, bom.Barcode)

				if err != nil {
					return err
				}

				tempDoc = findDoc
			} else {
				tempDoc = (*productBarcodeDict)[bomKey]
			}

			if _, ok := (*productBarcodeDict)[bomKey]; !ok {
				(*productBarcodeDict)[bomKey] = tempDoc
			}

			tempBOMView.FromProductBOM(tempDoc.ProductBarcodeData, bom)

			// if _, ok := (*bomViewDict)[tempDoc.Barcode]; !ok {
			// 	(*bomViewDict)[tempDoc.Barcode] = &tempBOMView
			// }

			if tempDoc.BOM != nil && len(*tempDoc.BOM) > 0 {
				err := BuildBOMView(ctx, findByBarcode, currentLevel, productBarcodeDict, bomViewDict, holdingCode, tempDoc.BOM, &tempBOMView.BOM)

				if err != nil {
					return err
				}
			}
		}

		if tempBOMView.BOM == nil {
			tempBOMView.BOM = []models.ProductBarcodeBOMView{}
		}

		*bomView = append(*bomView, tempBOMView)

	}

	return nil
}

func BuildBOMViewCache(
	ctx context.Context,
	findByBarcode func(ctx context.Context, holdingCode string, itemCode string, barcode string) (models.ProductBarcodeDoc, error),
	currentLevel int,
	productBarcodeDict *map[string]models.ProductBarcodeDoc,
	bomViewDict *map[string]*models.ProductBarcodeBOMView,
	holdingCode string,
	itemCode string,
	barcode string,
	childBOMs []models.BOMProductBarcode,
	bomView *models.ProductBarcodeBOMView) error {

	if currentLevel > 10 {
		return fmt.Errorf("BOM level is too deep")
	}

	tempBOMView := models.ProductBarcodeBOMView{}
	tempBOMView.Level = currentLevel

	key := productBarcodeBusinessKey(itemCode, barcode)
	if _, bomOk := (*bomViewDict)[key]; bomOk {
		tempBOMView = *(*bomViewDict)[key]
	} else {
		var tempDoc = models.ProductBarcodeDoc{}
		if _, ok := (*productBarcodeDict)[key]; !ok {
			findDoc, err := findByBarcode(ctx, holdingCode, itemCode, barcode)

			if err != nil {
				return err
			}

			tempDoc = findDoc
		} else {
			tempDoc = (*productBarcodeDict)[key]
		}

		if _, ok := (*productBarcodeDict)[key]; !ok {
			(*productBarcodeDict)[key] = tempDoc
		}

		var tempBOMs []models.BOMProductBarcode
		if len(childBOMs) == 0 {
			tempBOMView.FromProductBarcode(tempDoc.ProductBarcodeData)

		} else if tempDoc.BOM != nil && len(*tempDoc.BOM) > 0 {
			for _, bom := range childBOMs {
				tempBOMView.FromProductBOM(tempDoc.ProductBarcodeData, bom)
			}
		} else {
			tempBOMView.FromProductBarcode(tempDoc.ProductBarcodeData)
		}

		if tempDoc.BOM != nil {
			tempBOMs = *tempDoc.BOM
		}

		if len(tempBOMs) != 0 {
			for _, bom := range tempBOMs {
				if tempBOMs != nil {
					err := BuildBOMViewCache(ctx, findByBarcode, currentLevel+1, productBarcodeDict, bomViewDict, holdingCode, bom.ItemCode, bom.Barcode, *tempDoc.BOM, &tempBOMView)

					if err != nil {
						return err
					}
				}
			}
		}
	}

	if tempBOMView.BOM == nil {
		tempBOMView.BOM = []models.ProductBarcodeBOMView{}
	}

	if currentLevel == 0 {
		*bomView = tempBOMView
	} else {
		bomView.BOM = append(bomView.BOM, tempBOMView)
	}

	return nil
}

func productBarcodeBusinessKey(itemCode string, barcode string) string {
	return itemCode + "\x00" + barcode
}
