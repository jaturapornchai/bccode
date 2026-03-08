package services

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/product/productbarcode/models"
)

func (svc ProductBarcodeHttpService) InfoBomView(shopID string, barcode string) (models.ProductBarcodeBOMView, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	doc, err := svc.repo.FindByBarcode(ctx, shopID, barcode)

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
	// 	shopID, doc.Barcode, []models.BOMProductBarcode{}, &bomView)

	bomView.FromProductBarcode(doc.ProductBarcodeData)

	if _, ok := bomViewDict[doc.Barcode]; !ok {
		bomViewDict[doc.Barcode] = &bomView
	}

	bomView.Level = 1

	if doc.BOM != nil && len(*doc.BOM) > 0 {
		productBarcodeDict := map[string]models.ProductBarcodeDoc{}
		err = BuildBOMView(ctx, svc.repo.FindByBarcode, bomView.Level, &productBarcodeDict, &bomViewDict, shopID, doc.BOM, &bomView.BOM)
		if err != nil {
			return models.ProductBarcodeBOMView{}, err
		}
	}

	return bomView, nil
}

func (svc ProductBarcodeHttpService) ListBomView(shopID string, barcodes []string) ([]models.ProductBarcodeBOMView, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	var bomViews []models.ProductBarcodeBOMView
	var bomViewDict = map[string]*models.ProductBarcodeBOMView{}
	for _, barcode := range barcodes {
		doc, err := svc.repo.FindByBarcode(ctx, shopID, barcode)

		if err != nil {
			return []models.ProductBarcodeBOMView{}, err
		}

		if len(doc.ProductBarcode.Barcode) == 0 {
			return []models.ProductBarcodeBOMView{}, fmt.Errorf("barcode not found")
		}

		bomView := models.ProductBarcodeBOMView{}
		bomView.FromProductBarcode(doc.ProductBarcodeData)

		if _, ok := bomViewDict[doc.Barcode]; !ok {
			bomViewDict[doc.Barcode] = &bomView
		}

		if doc.BOM != nil && len(*doc.BOM) > 0 {
			productBarcodeDict := map[string]models.ProductBarcodeDoc{}
			err = BuildBOMView(ctx, svc.repo.FindByBarcode, 1, &productBarcodeDict, &bomViewDict, shopID, doc.BOM, &bomView.BOM)
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
	findByBarcode func(ctx context.Context, shopID string, barcode string) (models.ProductBarcodeDoc, error),
	currentLevel int,
	productBarcodeDict *map[string]models.ProductBarcodeDoc,
	bomViewDict *map[string]*models.ProductBarcodeBOMView,
	shopID string,
	BOMs *[]models.BOMProductBarcode,
	bomView *[]models.ProductBarcodeBOMView) error {

	if currentLevel > 10 {
		return fmt.Errorf("BOM level is too deep")
	}

	currentLevel += 1

	for _, bom := range *BOMs {

		tempBOMView := models.ProductBarcodeBOMView{}
		tempBOMView.Level = currentLevel

		if _, bomOk := (*bomViewDict)[bom.Barcode]; bomOk {
			tempBOMView = *(*bomViewDict)[bom.Barcode]
		} else {
			var tempDoc = models.ProductBarcodeDoc{}
			if _, ok := (*productBarcodeDict)[bom.Barcode]; !ok {
				findDoc, err := findByBarcode(ctx, shopID, bom.Barcode)

				if err != nil {
					return err
				}

				tempDoc = findDoc
			} else {
				tempDoc = (*productBarcodeDict)[bom.Barcode]
			}

			if _, ok := (*productBarcodeDict)[tempDoc.ProductBarcode.Barcode]; !ok {
				(*productBarcodeDict)[bom.Barcode] = tempDoc
			}

			tempBOMView.FromProductBOM(tempDoc.ProductBarcodeData, bom)

			// if _, ok := (*bomViewDict)[tempDoc.Barcode]; !ok {
			// 	(*bomViewDict)[tempDoc.Barcode] = &tempBOMView
			// }

			if tempDoc.BOM != nil && len(*tempDoc.BOM) > 0 {
				err := BuildBOMView(ctx, findByBarcode, currentLevel, productBarcodeDict, bomViewDict, shopID, tempDoc.BOM, &tempBOMView.BOM)

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
	findByBarcode func(ctx context.Context, shopID string, barcode string) (models.ProductBarcodeDoc, error),
	currentLevel int,
	productBarcodeDict *map[string]models.ProductBarcodeDoc,
	bomViewDict *map[string]*models.ProductBarcodeBOMView,
	shopID string,
	barcode string,
	childBOMs []models.BOMProductBarcode,
	bomView *models.ProductBarcodeBOMView) error {

	if currentLevel > 10 {
		return fmt.Errorf("BOM level is too deep")
	}

	tempBOMView := models.ProductBarcodeBOMView{}
	tempBOMView.Level = currentLevel

	if _, bomOk := (*bomViewDict)[barcode]; bomOk {
		tempBOMView = *(*bomViewDict)[barcode]
	} else {
		var tempDoc = models.ProductBarcodeDoc{}
		if _, ok := (*productBarcodeDict)[barcode]; !ok {
			findDoc, err := findByBarcode(ctx, shopID, barcode)

			if err != nil {
				return err
			}

			tempDoc = findDoc
		} else {
			tempDoc = (*productBarcodeDict)[barcode]
		}

		if _, ok := (*productBarcodeDict)[tempDoc.ProductBarcode.Barcode]; !ok {
			(*productBarcodeDict)[barcode] = tempDoc
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

		if tempDoc.GuidFixed == "2dJ5kfBc9tTIcHeB14I4P4PoTqP" {
			fmt.Println("tempDoc", tempDoc)
			fmt.Println("tempBOMs", tempBOMs)
		}

		if len(tempBOMs) != 0 {
			for _, bom := range tempBOMs {
				if tempBOMs != nil {
					err := BuildBOMViewCache(ctx, findByBarcode, currentLevel+1, productBarcodeDict, bomViewDict, shopID, bom.Barcode, *tempDoc.BOM, &tempBOMView)

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
