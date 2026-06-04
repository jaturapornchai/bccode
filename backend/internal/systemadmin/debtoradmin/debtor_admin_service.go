package debtoradmin

import (
	"context"
	debtorRepositories "smlcloudplatform/internal/debtaccount/debtor/repositories"
	debtorProcessModels "smlcloudplatform/internal/debtorprocess/models"
	debtorProcessRepository "smlcloudplatform/internal/debtorprocess/repositories"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IDebtorAdminService interface {
	ReSyncDebtor(holdingCode string) error
	ReCalcDebtorBalance(holdingCode string) error
}

type DebtorAdminService struct {
	repoMQ              debtorRepositories.IDebtorMessageQueueRepository
	repo                IDebtorAdminMongoRepository
	debtorProcessMQRepo debtorProcessRepository.IDebtorProcessMessageQueueRepository
	timeoutDuration     time.Duration
}

func NewDebtorAdminService(pst microservice.IPersisterMongo, producer microservice.IProducer) IDebtorAdminService {

	repo := NewDebtorAdminMongoRepository(pst)
	repoMQ := debtorRepositories.NewDebtorMessageQueueRepository(producer)

	debtorProcessMQRepo := debtorProcessRepository.NewDebtorProcessMessageQueueRepository(producer)

	return &DebtorAdminService{
		repoMQ:              repoMQ,
		repo:                repo,
		debtorProcessMQRepo: debtorProcessMQRepo,
		timeoutDuration:     time.Duration(30) * time.Second,
	}
}

func (svc DebtorAdminService) ReSyncDebtor(holdingCode string) error {

	ctx, cancel := context.WithTimeout(context.Background(), svc.timeoutDuration)
	defer cancel()

	debtors, err := svc.repo.FindDebtorByHoldingCode(ctx, holdingCode)
	if err != nil {
		return err
	}

	err = svc.repoMQ.CreateInBatch(debtors)
	if err != nil {
		return err
	}

	return nil
}

func (svc DebtorAdminService) ReCalcDebtorBalance(holdingCode string) error {

	ctx, cancel := context.WithTimeout(context.Background(), svc.timeoutDuration)
	defer cancel()

	debtors, err := svc.repo.FindDebtorByHoldingCode(ctx, holdingCode)
	if err != nil {
		return err
	}

	var requestStockProcessLists []debtorProcessModels.DebtorProcessRequest

	if len(debtors) > 0 {
		for _, debtor := range debtors {

			if debtor.Code != "" {
				requestStockProcessLists = append(requestStockProcessLists, debtorProcessModels.DebtorProcessRequest{
					HoldingCode: holdingCode,
					DebtorCode:  debtor.Code,
				})
			}
		}
	}

	err = svc.debtorProcessMQRepo.CreateInBatch(requestStockProcessLists)

	if err != nil {
		return err
	}

	return nil
}
