package services

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/product/productbarcode/usecases"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"strings"

	commonModels "smlcloudplatform/internal/models"
	msModels "smlcloudplatform/pkg/microservice/models"
)

type IProductBarcodeConsumeService interface {
	UpdateRefBarcode(holdingCode string, doc models.ProductBarcodeDoc) error
	UpdateProductType(holdingCode string, doc models.ProductType) error
	UpdateProductGroup(holdingCode string, doc models.ProductGroup) error
	UpdateProductUnit(holdingCode string, doc models.ProductUnit) error
	UpdateProductOrderType(holdingCode string, doc models.ProductOrderType) error
	UpSert(holdingCode string, businessCode string, barcode string, doc models.ProductBarcodeDoc) (*models.ProductBarcodePg, error)
	Delete(ctx context.Context, holdingCode string, businessCode string, barcode string) error
	ReSync(holdingCode string, businessCode string) error
}

type ProductBarcodeConsumeService struct {
	productPgRepo         repositories.ICompanyProductBarcodePGRepository
	productMongoRepo      repositories.IProductBarcodeRepository
	productClickhouseRepo repositories.IProductBarcodeClickhouseRepository
	phaser                usecases.IProductBarcodePhaser
}

func NewProductBarcodeConsumerService(
	pst microservice.IPersister,
	mongoDBPersister microservice.IPersisterMongo,
	clickhousePersister microservice.IPersisterClickHouse,
	phaser usecases.IProductBarcodePhaser,
) IProductBarcodeConsumeService {

	productPgRepo := repositories.NewProductBarcodePGRepository(pst)
	productMongoRepo := repositories.NewProductBarcodeRepository(mongoDBPersister, nil)
	productClickhouseRepo := repositories.NewProductBarcodeClickhouseRepository(clickhousePersister)

	return &ProductBarcodeConsumeService{
		productPgRepo:         productPgRepo,
		productMongoRepo:      productMongoRepo,
		phaser:                phaser,
		productClickhouseRepo: productClickhouseRepo,
	}
}

func (svc ProductBarcodeConsumeService) UpdateRefBarcode(holdingCode string, doc models.ProductBarcodeDoc) error {

	refProductBarcode := doc.ToRefBarcode()

	err := svc.productMongoRepo.UpdateRefBarcodeByKey(
		context.Background(),
		holdingCode,
		doc.ItemCode,
		doc.Barcode,
		refProductBarcode,
	)

	if err != nil {
		return err
	}

	return nil
}

func (svc ProductBarcodeConsumeService) UpdateProductType(holdingCode string, doc models.ProductType) error {
	err := svc.productMongoRepo.UpdateAllProductTypeByGUID(context.Background(), holdingCode, doc.GuidFixed, doc)

	if err != nil {
		return err
	}

	return nil
}

func (svc ProductBarcodeConsumeService) UpdateProductGroup(holdingCode string, doc models.ProductGroup) error {
	err := svc.productMongoRepo.UpdateAllProductGroup(context.Background(), holdingCode, doc)

	if err != nil {
		return err
	}

	return nil
}

func (svc ProductBarcodeConsumeService) UpdateProductUnit(holdingCode string, doc models.ProductUnit) error {
	err := svc.productMongoRepo.UpdateAllProductUnitByCode(context.Background(), holdingCode, doc)

	if err != nil {
		return err
	}

	return nil
}

func (svc ProductBarcodeConsumeService) UpdateProductOrderType(holdingCode string, doc models.ProductOrderType) error {
	err := svc.productMongoRepo.UpdateAllProductOrderTypeByGUID(context.Background(), holdingCode, doc.GuidFixed, doc)

	if err != nil {
		return err
	}

	return nil
}

func (svc ProductBarcodeConsumeService) UpSert(holdingCode string, businessCode string, barcode string, doc models.ProductBarcodeDoc) (*models.ProductBarcodePg, error) {
	holdingCode = strings.TrimSpace(holdingCode)
	businessCode = utils.NormalizeBusinessCode(businessCode)
	barcode = utils.NormalizeBusinessCode(barcode)
	if holdingCode == "" || businessCode == "" || barcode == "" {
		return nil, fmt.Errorf("holdingcode, businesscode and barcode are required")
	}
	doc.HoldingCode = holdingCode
	doc.BusinessCode = businessCode
	doc.Barcode = barcode

	pgDoc, err := svc.phaser.PhaseProductBarcodeDoc(&doc)
	if err != nil {
		return nil, err
	}

	pgDoc.HoldingCode = holdingCode
	pgDoc.BusinessCode = businessCode
	findbarcodePG, err := svc.productPgRepo.GetByCompanyBarcode(holdingCode, businessCode, barcode)
	if err != nil {
		return nil, err
	}

	if findbarcodePG != nil {
		err = svc.productPgRepo.UpdateInCompany(holdingCode, businessCode, barcode, pgDoc)
	} else {
		err = svc.productPgRepo.Create(pgDoc)
	}

	if err != nil {
		return nil, err
	}

	return pgDoc, nil
}

func (svc ProductBarcodeConsumeService) Delete(ctx context.Context, holdingCode string, businessCode string, barcode string) error {
	holdingCode = strings.TrimSpace(holdingCode)
	businessCode = utils.NormalizeBusinessCode(businessCode)
	barcode = utils.NormalizeBusinessCode(barcode)
	if holdingCode == "" || businessCode == "" || barcode == "" {
		return fmt.Errorf("holdingcode, businesscode and barcode are required")
	}

	err := svc.productPgRepo.DeleteInCompany(holdingCode, businessCode, barcode)
	if err != nil {
		return err
	}
	return nil
}
func (svc ProductBarcodeConsumeService) ReSync(holdingCode string, businessCode string) error {
	holdingCode = strings.TrimSpace(holdingCode)
	businessCode = utils.NormalizeBusinessCode(businessCode)
	if holdingCode == "" || businessCode == "" {
		return fmt.Errorf("holdingcode and businesscode are required")
	}

	// resync 100

	pageRequest := msModels.Pageable{
		Limit: 100,
		Page:  1,
		Sorts: []msModels.KeyInt{
			{
				Key:   "guidfixed",
				Value: -1,
			},
		},
	}

	for {
		barcodes, pages, err := svc.productMongoRepo.FindPageFilterInCompany(
			context.Background(),
			holdingCode,
			businessCode,
			nil,
			nil,
			pageRequest,
		)
		if err != nil {
			return err
		}

		for _, barcode := range barcodes {

			doc := models.ProductBarcodeDoc{
				ProductBarcodeData: models.ProductBarcodeData{
					HoldingCodeentity: commonModels.HoldingCodeentity{
						HoldingCode: holdingCode,
					},
					BusinessCode:       businessCode,
					ProductBarcodeInfo: barcode,
				},
			}

			if _, err := svc.UpSert(holdingCode, businessCode, barcode.Barcode, doc); err != nil {
				return err
			}
		}

		if pages.TotalPage > int64(pageRequest.Page) {
			pageRequest.Page++
		} else {
			break
		}
	}

	return nil
}
