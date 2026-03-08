package datatransfer

import (
	"context"
	"smlcloudplatform/internal/payment/qrpayment/models"
	qrPaymentRepository "smlcloudplatform/internal/payment/qrpayment/repositories"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type QRPaymentDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IQRPaymentDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.QrPaymentDoc, mongopagination.PaginationData, error)
}

type QRPaymentDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.QrPaymentDoc]
}

func NewQRPaymentDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IQRPaymentDataTransferRepository {
	repo := &QRPaymentDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.QrPaymentDoc](mongodbPersister)
	return repo
}

func (repo QRPaymentDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.QrPaymentDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewQRPaymentDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {

	return &QRPaymentDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *QRPaymentDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewQRPaymentDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := qrPaymentRepository.NewQrPaymentRepository(pdt.transferConnection.GetTargetConnection())

	pageRequest := msModels.Pageable{
		Limit: 100,
		Page:  1,
	}

	for {

		docs, pages, err := sourceRepository.FindPage(ctx, shopID, nil, pageRequest)
		if err != nil {
			return err
		}

		if len(docs) > 0 {

			if targetShopID != "" {
				for i := range docs {
					docs[i].ShopID = targetShopID
					docs[i].ID = primitive.NewObjectID()
				}
			}

			err = targetRepository.CreateInBatch(ctx, docs)
			if err != nil {
				return err
			}
		}

		if pages.TotalPage > int64(pageRequest.Page) {
			pageRequest.Page++
		} else {
			break
		}
	}

	return nil
}
