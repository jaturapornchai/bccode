package creditoradmin

import (
	"context"
	creditorProcessModels "smlcloudplatform/internal/creditorprocess/models"
	creditorProcessRepositories "smlcloudplatform/internal/creditorprocess/repositories"
	creditorRepositories "smlcloudplatform/internal/debtaccount/creditor/repositories"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type ICreditorAdminService interface {
	ReSyncCreditor(holdingCode string) error
	ReCalcCreditorBalance(holdingCode string) error
}

type CreditorAdminService struct {
	kafkaRepo             creditorRepositories.ICreditorMessageQueueRepository
	mongoRepo             ICreditorAdminMongoRepository
	creditorProcessMQRepo creditorProcessRepositories.ICreditorProcessMessageQueueRepository
	timeoutDuration       time.Duration
}

func NewCreditorAdminService(pst microservice.IPersisterMongo, producer microservice.IProducer) ICreditorAdminService {

	mongoRepo := NewCreditorAdminMongoRepository(pst)
	kafkaRepo := creditorRepositories.NewCreditorMessageQueueRepository(producer)

	creditorProcessMQRepo := creditorProcessRepositories.NewCreditorProcessMessageQueueRepository(producer)
	return &CreditorAdminService{
		mongoRepo:             mongoRepo,
		kafkaRepo:             kafkaRepo,
		creditorProcessMQRepo: creditorProcessMQRepo,
		timeoutDuration:       time.Duration(30) * time.Second,
	}
}

func (svc CreditorAdminService) ReSyncCreditor(holdingCode string) error {

	ctx, cancel := context.WithTimeout(context.Background(), svc.timeoutDuration)
	defer cancel()

	// find creditor by holdingcode
	creditors, err := svc.mongoRepo.FindCreditorByHoldingCode(ctx, holdingCode)
	if err != nil {
		return err
	}

	// Send Message to MQ
	err = svc.kafkaRepo.CreateInBatch(creditors)
	if err != nil {
		return err
	}

	return nil
}

func (svc CreditorAdminService) ReCalcCreditorBalance(holdingCode string) error {

	ctx, cancel := context.WithTimeout(context.Background(), svc.timeoutDuration)
	defer cancel()

	creditors, err := svc.mongoRepo.FindCreditorByHoldingCode(ctx, holdingCode)
	if err != nil {
		return err
	}

	var requestStockProcessLists []creditorProcessModels.CreditorProcessRequest

	if len(creditors) > 0 {
		for _, creditor := range creditors {

			if creditor.Code != "" {
				requestStockProcessLists = append(requestStockProcessLists, creditorProcessModels.CreditorProcessRequest{
					HoldingCode:  holdingCode,
					CreditorCode: creditor.Code,
				})
			}
		}
	}

	err = svc.creditorProcessMQRepo.CreateInBatch(requestStockProcessLists)

	if err != nil {
		return err
	}

	return nil
}
