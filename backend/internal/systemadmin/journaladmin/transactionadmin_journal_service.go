package journal

import (
	"context"
	journalRepositories "smlcloudplatform/internal/vfgl/journal/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"
	"time"
)

type IJournalTransactionAdminService interface {
	ReSyncJournalTransactionDoc(shopID string) error
	ReSyncJournalDeleteTransactionDoc(shopID string) error
}

type JournalTransactionAdminService struct {
	mongoRepo       IJournalTransactionAdminRepository
	kafkaRepo       journalRepositories.IJournalMqRepository
	timeoutDuration time.Duration
}

func NewJournalTransactionAdminService(pst microservice.IPersisterMongo, kfProducer microservice.IProducer) IJournalTransactionAdminService {

	mongorepo := NewJournalTransactionAdminRepository(pst)
	kafkarepo := journalRepositories.NewJournalMqRepository(kfProducer)

	return &JournalTransactionAdminService{
		mongoRepo:       mongorepo,
		kafkaRepo:       kafkarepo,
		timeoutDuration: time.Duration(30) * time.Second,
	}
}

func (s *JournalTransactionAdminService) ReSyncJournalTransactionDoc(shopID string) error {

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

		docs, pages, err := s.mongoRepo.FindJournalTransactionDocByShopID(ctx, shopID, false, pageRequest)
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

func (s *JournalTransactionAdminService) ReSyncJournalDeleteTransactionDoc(shopID string) error {
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

		docs, pages, err := s.mongoRepo.FindJournalTransactionDocByShopID(ctx, shopID, true, pageRequest)
		if err != nil {
			return err
		}

		err = s.kafkaRepo.DeleteInBatch(docs)
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
