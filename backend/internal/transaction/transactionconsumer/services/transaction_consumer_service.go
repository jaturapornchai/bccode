package services

import "smlcloudplatform/pkg/microservice"

type ITransactionDocConsumer interface {
	ConsumeOnCreateOrUpdate(ctx microservice.IContext) error
	ConsumeOnBulkCreateOrUpdate(ctx microservice.IContext) error
	ConsumeOnDelete(ctx microservice.IContext) error
	ConsumeOnBulkDelete(ctx microservice.IContext) error
	RegisterConsumer(ms *microservice.Microservice)
}

// type ITransactionConsumerService interface {
// 	Upsert(shopID string, docNo string, doc models.StockTransaction) error
// 	Delete(shopID string, docNo string) error
// 	// UpsertPurchaseTrx(shopID string, docNo string, doc purchaseModels.PurchaseDoc) error
// }

// func NewTransactionConsumerService(pst microservice.IPersister, producer microservice.IProducer) ITransactionConsumerService {

// 	pgRepo := repositories.NewTransactionPGRepository(pst)
// 	stockProcessMGRepo := stockProcessRepository.NewPurchaseMessageQueueRepository(producer)
// 	return &TransactionConsumerService{
// 		transactionPGRepo:  pgRepo,
// 		stockProcessMGRepo: stockProcessMGRepo,
// 	}
// }

// type TransactionConsumerService struct {
// 	phaser             usecases.ITransactionPhaser
// 	transactionPGRepo  repositories.ITransactionPGRepository
// 	stockProcessMGRepo stockProcessRepository.IStockProcessMessageQueueRepository
// }

// func (svc *TransactionConsumerService) Upsert(shopID string, docNo string, doc models.StockTransaction) error {

// 	findTrx, err := svc.transactionPGRepo.Get(shopID, docNo)
// 	if err != nil {
// 		if err != gorm.ErrRecordNotFound {
// 			return err
// 		}
// 	}

// 	if findTrx == nil {
// 		err = svc.transactionPGRepo.Create(doc)
// 		if err != nil {
// 			return err
// 		}
// 	} else {

// 		for idx, tmp := range *doc.Details {
// 			for _, detail := range *findTrx.Details {
// 				if tmp.DocNo == detail.DocNo && tmp.Barcode == detail.Barcode && tmp.LineNumber == detail.LineNumber {
// 					// tmpAccBook := *docPg.AccountBook
// 					(*doc.Details)[idx].ID = detail.ID
// 					(*doc.Details)[idx].SumOfCost = detail.SumOfCost
// 					(*doc.Details)[idx].AverageCost = detail.AverageCost
// 				}
// 			}
// 		}

// 		isEqual := doc.CompareTo(findTrx)

// 		if !isEqual {

// 			err = svc.transactionPGRepo.Update(shopID, docNo, doc)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	var requestStockProcessLists []stockProcessModels.StockProcessRequest

// 	if len(*doc.Details) > 0 {
// 		for _, item := range *doc.Details {

// 			if item.Barcode != "" {
// 				requestStockProcessLists = append(requestStockProcessLists, stockProcessModels.StockProcessRequest{
// 					ShopID:  shopID,
// 					Barcode: item.Barcode,
// 				})
// 			}
// 		}
// 	}

// 	svc.stockProcessMGRepo.CreateInBatch(requestStockProcessLists)

// 	return nil
// }

// func (svc *TransactionConsumerService) Delete(shopID string, docNo string) error {
// 	err := svc.transactionPGRepo.Delete(shopID, docNo)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
