package productadmin

import (
	"context"
	"smlcloudplatform/internal/logger"
	productBarcodeRepositories "smlcloudplatform/internal/product/productbarcode/repositories"
	stockProcessModels "smlcloudplatform/internal/stockprocess/models"
	stockProcessRepository "smlcloudplatform/internal/stockprocess/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"
	"time"
)

type IProductAdminService interface {
	ReSyncProductBarcode(holdingCode string) error
	ReCalcStockBalance(holdingCode string, barcode string) error
	DeleteProductBarcodeAll(holdingCode string, userName string) error
}

type ProductAdminService struct {
	kafkaRepo          productBarcodeRepositories.IProductBarcodeMessageQueueRepository
	mongoRepo          IProductAdminMongoRepository
	stockProcessMGRepo stockProcessRepository.IStockProcessMessageQueueRepository
	timeoutDuration    time.Duration
}

func NewProductAdminService(pst microservice.IPersisterMongo, kfProducer microservice.IProducer) IProductAdminService {
	kafkaRepo := productBarcodeRepositories.NewProductBarcodeMessageQueueRepository(kfProducer)
	mongoRepo := NewProductAdminMongoRepository(pst)

	stockProcessMGRepo := stockProcessRepository.NewStockProcessMessageQueueRepository(kfProducer)

	return &ProductAdminService{
		kafkaRepo:          kafkaRepo,
		mongoRepo:          mongoRepo,
		stockProcessMGRepo: stockProcessMGRepo,
		timeoutDuration:    time.Duration(300) * time.Second,
	}
}

func (svc ProductAdminService) ReSyncProductBarcode(holdingCode string) error {

	// find product barcode by holding_code
	ctx, cancel := context.WithTimeout(context.Background(), svc.timeoutDuration)
	defer cancel()

	pageRequest := msModels.Pageable{
		Limit: 20,
		Page:  1,
		Sorts: []msModels.KeyInt{
			{
				Key:   "guid_fixed",
				Value: -1,
			},
		},
	}

	for {
		barcodes, pages, err := svc.mongoRepo.FindPage(ctx, holdingCode, nil, pageRequest)
		if err != nil {
			return err
		}

		logger.GetLogger().Info("ReSyncProductBarcode page ", pageRequest.Page, " totalpage ", pages.TotalPage, " len ", len(barcodes))
		// barcodeDocs := []models.ProductBarcodeDoc{}
		// for _, barcode := range barcodes {

		// 	barcodeDocs = append(barcodeDocs, barcode)
		// }

		// Send Message to MQ
		err = svc.kafkaRepo.CreateInBatch(barcodes)
		if err != nil {
			return err
		}

		if pages.TotalPage > int64(pageRequest.Page) {
			pageRequest.Page++
		} else {
			break
		}
	}

	return nil

}

func (svc ProductAdminService) ReCalcStockBalance(holdingCode string, barcode string) error {

	ctx, cancel := context.WithTimeout(context.Background(), svc.timeoutDuration)
	defer cancel()

	var requestStockProcessLists []stockProcessModels.StockProcessRequest

	if barcode != "" {
		findBarcode, err := svc.mongoRepo.FindProductAndBarcode(ctx, holdingCode, barcode)
		if err != nil {
			return err
		}

		if findBarcode.GuidFixed != "" {

			if findBarcode.Barcode != "" {
				requestStockProcessLists = append(requestStockProcessLists, stockProcessModels.StockProcessRequest{
					HoldingCode: holdingCode,
					Barcode:     findBarcode.Barcode,
				})
			}
		}
	} else {
		barcodes, err := svc.mongoRepo.FindProductBarcodeByHoldingCode(ctx, holdingCode)
		if err != nil {
			return err
		}

		for _, item := range barcodes {

			if item.Barcode != "" {
				requestStockProcessLists = append(requestStockProcessLists, stockProcessModels.StockProcessRequest{
					HoldingCode: holdingCode,
					Barcode:     item.Barcode,
				})
			}
		}
	}

	if len(requestStockProcessLists) == 0 {
		return nil
	}

	err := svc.stockProcessMGRepo.CreateInBatch(requestStockProcessLists)

	if err != nil {
		return err
	}

	return nil
}

func (svc ProductAdminService) DeleteProductBarcodeAll(holdingCode string, userName string) error {

	ctx, cancel := context.WithTimeout(context.Background(), svc.timeoutDuration)
	defer cancel()

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
		barcodes, pages, err := svc.mongoRepo.FindPage(ctx, holdingCode, nil, pageRequest)
		if err != nil {
			return err
		}

		ids := []string{}
		for _, barcode := range barcodes {

			ids = append(ids, barcode.ID.Hex())
		}

		err = svc.mongoRepo.DeleteProductBarcodeByHoldingCode(ctx, holdingCode, userName, ids)
		if err != nil {
			return err
		}

		// Send Message to MQ
		err = svc.kafkaRepo.DeleteInBatch(barcodes)
		if err != nil {
			return err
		}

		if pages.TotalPage > int64(pageRequest.Page) {
			pageRequest.Page++
		} else {
			break
		}
	}

	return nil
}
