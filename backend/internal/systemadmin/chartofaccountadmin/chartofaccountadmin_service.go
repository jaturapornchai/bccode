package chartofaccountadmin

import (
	"context"
	chartOfAccountRepositories "smlcloudplatform/internal/vfgl/chartofaccount/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"
	"time"
)

type IChartOfAccountAdminService interface {
	ReSyncChartOfAccountDoc(holdingCode string) error
}

type ChartOfAccountAdminService struct {
	kafkaRepo       chartOfAccountRepositories.IChartOfAccountMQRepository
	mongoRepo       IChartOfAccountAdminRepository
	timeoutDuration time.Duration
}

func NewChartOfAccountAdminService(pst microservice.IPersisterMongo, kfProducer microservice.IProducer) IChartOfAccountAdminService {
	kafkaRepo := chartOfAccountRepositories.NewChartOfAccountMQRepository(kfProducer)
	mongoRepo := NewChartOfAccountAdminRepository(pst)

	return &ChartOfAccountAdminService{
		kafkaRepo:       kafkaRepo,
		mongoRepo:       mongoRepo,
		timeoutDuration: time.Duration(30) * time.Second,
	}
}
func (s *ChartOfAccountAdminService) ReSyncChartOfAccountDoc(holdingCode string) error {

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
		ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
		defer cancel()

		docs, pages, err := s.mongoRepo.FindChartOfAccountDocByHoldingCode(ctx, holdingCode, false, pageRequest)
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
