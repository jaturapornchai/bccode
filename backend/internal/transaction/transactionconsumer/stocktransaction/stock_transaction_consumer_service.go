package stocktransaction

import (
	"smlcloudplatform/internal/logger"
	stockProcessRepository "smlcloudplatform/internal/stockprocess/repositories"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/pkg/microservice"

	"gorm.io/gorm"
)

type IStockTransactionConsumerService interface {
	Upsert(holdingCode string, docNo string, doc models.StockTransaction) error
	Delete(holdingCode string, docNo string) error
}

func NewStockTransactionConsumerService(
	pst microservice.IPersister,
	producer microservice.IProducer,
) IStockTransactionConsumerService {

	pgRepo := NewStockTransactionPGRepository(pst)
	stockProcessMGRepo := stockProcessRepository.NewStockProcessMessageQueueRepository(producer)
	stockProcessPhaser := StockProcessTransactionPhaser{}

	return &StockTransactionConsumerService{
		transactionPGRepo:  pgRepo,
		stockProcessMGRepo: stockProcessMGRepo,
		stockProcessPhaser: stockProcessPhaser,
	}
}

type StockTransactionConsumerService struct {
	transactionPGRepo  IStockTransactionPGRepository
	stockProcessMGRepo stockProcessRepository.IStockProcessMessageQueueRepository
	stockProcessPhaser IStockProcessTransactionPhaser
}

func (svc *StockTransactionConsumerService) Upsert(holdingCode string, docNo string, doc models.StockTransaction) error {

	findTrx, err := svc.transactionPGRepo.Get(holdingCode, docNo)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if findTrx == nil {
		err = svc.transactionPGRepo.Create(doc)
		if err != nil {
			return err
		}
		err, stockProcesList := svc.stockProcessPhaser.PhaseStockTransactionProcess(doc)
		if err == nil {
			svc.stockProcessMGRepo.CreateInBatch(*stockProcesList)
			if err != nil {
				return err
			}
		}
	} else {

		isEqual := findTrx.CompareTo(&doc)

		if !isEqual {
			err = svc.transactionPGRepo.Update(holdingCode, docNo, doc)
			if err != nil {
				return err
			}

			err, stockProcesList := svc.stockProcessPhaser.PhaseStockTransactionProcess(doc)
			if err == nil {
				err = svc.stockProcessMGRepo.CreateInBatch(*stockProcesList)
				if err != nil {
					return err
				}
			}
		} else {
			logger.GetLogger().Info("Stock Transaction No Change")
		}
	}

	return nil
}

func (svc *StockTransactionConsumerService) Delete(holdingCode string, docNo string) error {

	getDoc, _ := svc.transactionPGRepo.Get(holdingCode, docNo)

	err := svc.transactionPGRepo.Delete(holdingCode, docNo)
	if err != nil {
		return err
	}

	if getDoc != nil {
		err, stockProcesList := svc.stockProcessPhaser.PhaseStockTransactionProcess(*getDoc)
		if err == nil {
			svc.stockProcessMGRepo.CreateInBatch(*stockProcesList)
		}
	}

	return nil
}
