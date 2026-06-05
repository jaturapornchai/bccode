package purchasereceive

import (
	"context"
	purchasePartialRepository "smlcloudplatform/internal/transaction/purchasepartial/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"
	"time"
)

type IPurchaseReceiveTransactionAdminService interface {
	ReSyncPurchaseReceiveDoc(holdingCode string) error
	ReSyncPurchaseReceiveDeleteDoc(holdingCode string) error
}

type PurchaseReceiveTransactionAdminService struct {
	mongoRepo       IPurchaseReceiveTransactionAdminRepositories
	kafkaRepo       purchasePartialRepository.IPurchasepartialMessageQueueRepository
	timeoutDuration time.Duration
}

func NewPurchaseReceiveTransactionAdminService(pst microservice.IPersisterMongo, kfProducer microservice.IProducer) IPurchaseReceiveTransactionAdminService {

	kafkaRepo := purchasePartialRepository.NewPurchasepartialMessageQueueRepository(kfProducer)
	mongoRepo := NewPurchaseReceiveTransactionAdminRepositories(pst)
	return &PurchaseReceiveTransactionAdminService{
		kafkaRepo:       kafkaRepo,
		mongoRepo:       mongoRepo,
		timeoutDuration: time.Duration(30) * time.Second,
	}
}

func (s *PurchaseReceiveTransactionAdminService) ReSyncPurchaseReceiveDoc(holdingCode string) error {

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

		docs, pages, err := s.mongoRepo.FindPage(ctx, holdingCode, nil, pageRequest)
		if err != nil {
			return err
		}

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

func (s *PurchaseReceiveTransactionAdminService) ReSyncPurchaseReceiveDeleteDoc(holdingCode string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindPurchaseReceiveDocDeleteByHoldingCode(ctx, holdingCode)

	if err != nil {
		return err
	}

	err = s.kafkaRepo.DeleteInBatch(docs)
	if err != nil {
		return err
	}
	return nil
}
