package stockbalance

import (
	"context"
	trans_models "smlcloudplatform/internal/transaction/models"
	stockBalanceModels "smlcloudplatform/internal/transaction/stockbalance/models"
	stockReceiveProductRepositories "smlcloudplatform/internal/transaction/stockbalance/repositories"
	stockBalanceDetailModels "smlcloudplatform/internal/transaction/stockbalancedetail/models"
	stockReceiveProductDetailRepositories "smlcloudplatform/internal/transaction/stockbalancedetail/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"
)

type IStockBalanceProductTransactionAdminService interface {
	ReSyncStockBalanceProductDoc(holdingCode string) error
	ReSyncStockBalanceProductDeleteDoc(holdingCode string) error
}

type StockBalanceProductTransactionAdminService struct {
	mongoRepo       IStockBalanceTransactionAdminRepository
	stockDetailRepo stockReceiveProductDetailRepositories.IStockBalanceDetailRepository
	kafkaRepo       stockReceiveProductRepositories.IStockBalanceMessageQueueRepository
	timeoutDuration time.Duration
}

func NewStockBalanceProductTransactionAdminService(
	pst microservice.IPersisterMongo,
	kfProducer microservice.IProducer,
) IStockBalanceProductTransactionAdminService {

	mongoRepo := NewStockBalanceTransactionAdminRepository(pst)
	stockDetailRepo := stockReceiveProductDetailRepositories.NewStockBalanceDetailRepository(pst)
	kafkaRepo := stockReceiveProductRepositories.NewStockBalanceMessageQueueRepository(kfProducer)

	return &StockBalanceProductTransactionAdminService{
		mongoRepo:       mongoRepo,
		stockDetailRepo: stockDetailRepo,
		kafkaRepo:       kafkaRepo,
		timeoutDuration: time.Duration(30) * time.Second,
	}
}

func (s *StockBalanceProductTransactionAdminService) ReSyncStockBalanceProductDoc(holdingCode string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindStockBalanceDocByHoldingCode(ctx, holdingCode)
	if err != nil {
		return err
	}

	stockBalanceMessages := []stockBalanceModels.StockBalanceMessage{}

	for _, doc := range docs {
		stockBalanceMessage := stockBalanceModels.StockBalanceMessage{}
		stockBalanceMessage.HoldingCode = doc.HoldingCode
		stockBalanceMessage.GuidFixed = doc.GuidFixed
		stockBalanceMessage.StockBalance = doc.StockBalance

		// get stock balance details
		stockDetails, err := s.GetStockBalanceDetail(doc.HoldingCode, doc.DocNo)
		if err != nil {
			return err
		}

		docDetail := make([]trans_models.Detail, 0)

		for _, detail := range *stockDetails {
			docDetail = append(docDetail, detail.Detail)
		}

		stockBalanceMessage.Details = &docDetail
		stockBalanceMessages = append(stockBalanceMessages, stockBalanceMessage)
	}

	err = s.kafkaRepo.CreateInBatch(stockBalanceMessages)
	if err != nil {
		return err
	}

	return nil
}

func (s *StockBalanceProductTransactionAdminService) ReSyncStockBalanceProductDeleteDoc(holdingCode string) error {

	ctx, cancel := context.WithTimeout(context.Background(), s.timeoutDuration)
	defer cancel()

	docs, err := s.mongoRepo.FindStockBalanceDocDeleteByHoldingCode(ctx, holdingCode)

	if err != nil {
		return err
	}

	stockBalanceMessages := []stockBalanceModels.StockBalanceMessage{}

	for _, doc := range docs {
		stockBalanceMessage := stockBalanceModels.StockBalanceMessage{}
		stockBalanceMessage.StockBalance = doc.StockBalance

		// get stock balance details
		stockDetails, err := s.GetStockBalanceDetail(doc.HoldingCode, doc.DocNo)
		if err != nil {
			return err
		}

		docDetail := make([]trans_models.Detail, 0)

		for _, detail := range *stockDetails {
			docDetail = append(docDetail, detail.Detail)
		}

		stockBalanceMessage.Details = &docDetail
		stockBalanceMessages = append(stockBalanceMessages, stockBalanceMessage)

	}

	err = s.kafkaRepo.DeleteInBatch(stockBalanceMessages)
	if err != nil {
		return err
	}
	return nil
}

func (s *StockBalanceProductTransactionAdminService) GetStockBalanceDetail(holdingCode string, docNo string) (*[]stockBalanceDetailModels.StockBalanceDetailInfo, error) {

	filters := map[string]interface{}{
		"docno": docNo,
	}

	searchInFields := []string{
		"docno",
	}

	pageRequest := micromodels.Pageable{
		Limit: 20,
		Page:  1,
		Sorts: []micromodels.KeyInt{
			{
				Key:   "guidfixed",
				Value: -1,
			},
		},
	}

	docList := make([]stockBalanceDetailModels.StockBalanceDetailInfo, 0)

	for {

		timeOut := time.Duration(600) * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeOut)
		defer cancel()

		docs, pages, err := s.stockDetailRepo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageRequest)
		if err != nil {
			return nil, err
		}

		docList = append(docList, docs...)

		// for _, doc := range docs {
		// 	details = append(details, models.StockBalanceTransactionDetailPG{
		// 		TransactionDetailPG: models.TransactionDetailPG{
		// 			GuidFixed:           doc.GuidFixed,
		// 			DocNo:               doc.DocNo,
		// 			HoldingCode:              holdingCode,
		// 			LineNumber:          int8(doc.LineNumber),
		// 			DocRef:              doc.DocRef,
		// 			Barcode:             doc.Barcode,
		// 			ItemType:            doc.ItemType,
		// 			ItemGuid:            doc.ItemGuid,
		// 			VatType:             doc.VatType,
		// 			TaxType:             doc.TaxType,
		// 			StandValue:          doc.StandValue,
		// 			DivideValue:         doc.DivideValue,
		// 			WhCode:              doc.WhCode,
		// 			LocationCode:        doc.LocationCode,
		// 			Qty:                 doc.Qty,
		// 			Price:               doc.Price,
		// 			PriceExcludeVat:     doc.PriceExcludeVat,
		// 			SumAmount:           doc.SumAmount,
		// 			SumAmountExcludeVat: doc.SumAmountExcludeVat,
		// 			Discount:            doc.Discount,
		// 			DiscountAmount:      doc.DiscountAmount,
		// 		},
		// 	})
		// }

		if pages.TotalPage > int64(pageRequest.Page) {
			pageRequest.Page++
		} else {
			break
		}
	}

	return &docList, nil
}
