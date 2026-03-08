package saleinvoice

import (
	"context"
	saleInvoiceRepository "smlcloudplatform/internal/transaction/saleinvoice/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type ISaleInvoiceTransactionAdminService interface {
	ReSyncSaleInvoiceDoc(shopID string) error
	ReSyncSaleInvoiceDeleteDoc(shopID string) error
	ReSyncSaleInvoiceDocByDate(shopID string, date string) error
	ReSyncSaleInvoiceDeleteDocByDate(shopID string, date time.Time) error
}

type SaleInvoiceTransactionAdminService struct {
	mongoRepo       ISaleInvoiceTransactionAdminRepository
	kafkaRepo       saleInvoiceRepository.ISaleInvoiceMessageQueueRepository
	timeoutDuration time.Duration
}

func NewSaleInvoiceTransactionAdminService(pst microservice.IPersisterMongo, kfProducer microservice.IProducer) ISaleInvoiceTransactionAdminService {

	mongoRepo := NewSaleInvoiceTransactionAdminRepository(pst)
	kafkaRepo := saleInvoiceRepository.NewSaleInvoiceMessageQueueRepository(kfProducer)

	return &SaleInvoiceTransactionAdminService{
		mongoRepo:       mongoRepo,
		kafkaRepo:       kafkaRepo,
		timeoutDuration: time.Duration(30) * time.Second,
	}
}

func (s *SaleInvoiceTransactionAdminService) ReSyncSaleInvoiceDoc(shopID string) error {

	pageRequest := msModels.Pageable{
		Limit: 20,
		Page:  1,
		Sorts: []msModels.KeyInt{
			{
				Key:   "guidfixed",
				Value: -1,
			},
		},
	}

	for {
		ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
		defer cancel()

		docs, pages, err := s.mongoRepo.FindPage(ctx, shopID, nil, pageRequest)
		if err != nil {
			return err
		}

		// barcodes, pages, err := svc.mongoRepo.FindPage(shopID, nil, pageRequest)
		// 	if err != nil {
		// 		return err
		// 	}

		err = s.kafkaRepo.CreateInBatch(docs)
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

func (s *SaleInvoiceTransactionAdminService) ReSyncSaleInvoiceDeleteDoc(shopID string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindSaleInvoiceDeleteByShopID(ctx, shopID)

	if err != nil {
		return err
	}

	err = s.kafkaRepo.DeleteInBatch(docs)
	if err != nil {
		return err
	}
	return nil
}

func (s *SaleInvoiceTransactionAdminService) ReSyncSaleInvoiceDocByDate(shopID string, datestr string) error {

	pageRequest := msModels.Pageable{
		Limit: 20,
		Page:  1,
		Sorts: []msModels.KeyInt{
			{
				Key:   "guidfixed",
				Value: -1,
			},
		},
	}

	// create a time range filter
	date, err := time.Parse("2006-01-02", datestr)
	if err != nil {
		return err
	}

	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endDate := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, time.UTC)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
		defer cancel()

		dateFilter := bson.M{
			"docdatetime": bson.M{
				"$gte": startDate,
				"$lt":  endDate,
			},
		}

		docs, pages, err := s.mongoRepo.FindPageFilter(ctx, shopID, dateFilter, nil, pageRequest)
		if err != nil {
			return err
		}

		// barcodes, pages, err := svc.mongoRepo.FindPage(shopID, nil, pageRequest)
		// 	if err != nil {
		// 		return err
		// 	}

		err = s.kafkaRepo.CreateInBatch(docs)
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

func (s *SaleInvoiceTransactionAdminService) ReSyncSaleInvoiceDeleteDocByDate(shopID string, date time.Time) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindSaleInvoiceDeleteByShopID(ctx, shopID)

	if err != nil {
		return err
	}

	err = s.kafkaRepo.DeleteInBatch(docs)
	if err != nil {
		return err
	}
	return nil
}
