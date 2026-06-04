package services

import (
	"context"
	"smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/product/productbarcode/usecases"
	"smlcloudplatform/pkg/microservice"

	commonModels "smlcloudplatform/internal/models"
	msModels "smlcloudplatform/pkg/microservice/models"
)

type IProductBarcodeConsumeService interface {
	UpdateRefBarcode(holdingCode string, doc models.ProductBarcodeDoc) error
	UpdateProductType(holdingCode string, doc models.ProductType) error
	UpdateProductGroup(holdingCode string, doc models.ProductGroup) error
	UpdateProductUnit(holdingCode string, doc models.ProductUnit) error
	UpdateProductOrderType(holdingCode string, doc models.ProductOrderType) error
	UpSert(holdingCode string, barcode string, doc models.ProductBarcodeDoc) (*models.ProductBarcodePg, error)
	Delete(ctx context.Context, holdingCode string, barcode string) error
	ReSync(holdingCode string) error
}

type ProductBarcodeConsumeService struct {
	productPgRepo         repositories.IProductBarcodePGRepository
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

	err := svc.productMongoRepo.UpdateRefBarcodeByGUID(context.Background(), holdingCode, doc.GuidFixed, refProductBarcode)

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
	err := svc.productMongoRepo.UpdateAllProductGroupByCode(context.Background(), holdingCode, doc)

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

func (svc ProductBarcodeConsumeService) UpSert(holdingCode string, barcode string, doc models.ProductBarcodeDoc) (*models.ProductBarcodePg, error) {

	pgDoc, err := svc.phaser.PhaseProductBarcodeDoc(&doc)
	if err != nil {
		return nil, err
	}

	findbarcodePG, err := svc.productPgRepo.Get(holdingCode, pgDoc.Barcode)
	if err != nil {
		return nil, err
	}

	if findbarcodePG != nil {
		err = svc.productPgRepo.Update(holdingCode, pgDoc.Barcode, pgDoc)
	} else {
		err = svc.productPgRepo.Create(pgDoc)
	}

	if err != nil {
		return nil, err
	}

	return pgDoc, nil
}

func (svc ProductBarcodeConsumeService) Delete(ctx context.Context, holdingCode string, barcode string) error {

	err := svc.productPgRepo.Delete(holdingCode, barcode)
	if err != nil {
		return err
	}
	return nil
}
func (svc ProductBarcodeConsumeService) ReSync(holdingCode string) error {

	// resync 100

	pageRequest := msModels.Pageable{
		Limit: 100,
		Page:  1,
		Sorts: []msModels.KeyInt{
			{
				Key:   "guid_fixed",
				Value: -1,
			},
		},
	}

	for {
		barcodes, pages, err := svc.productMongoRepo.FindPage(context.Background(), holdingCode, nil, pageRequest)
		if err != nil {
			return err
		}

		for _, barcode := range barcodes {

			doc := models.ProductBarcodeDoc{
				ProductBarcodeData: models.ProductBarcodeData{
					HoldingCodeentity: commonModels.HoldingCodeentity{
						HoldingCode: holdingCode,
					},
					ProductBarcodeInfo: barcode,
				},
			}

			svc.UpSert(holdingCode, barcode.Barcode, doc)
		}

		if pages.TotalPage > int64(pageRequest.Page) {
			pageRequest.Page++
		} else {
			break
		}
	}

	return nil
}
