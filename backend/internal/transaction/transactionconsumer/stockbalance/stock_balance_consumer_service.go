package stockbalance

import (
	"context"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/stockbalancedetail/repositories"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"
)

type IStockReceiveTransactionConsumerService interface {
	GetStockBalanceDetail(holdingCode string, docNo string) (*[]models.StockBalanceTransactionDetailPG, error)
	Upsert(holdingCode string, docNo string, doc models.StockBalanceTransactionPG) error
	Delete(holdingCode string, docNo string) error
}

type StockReceiveTransactionConsumerService struct {
	repo                              IStockReceiveTransactionPGRepository
	stockBalanceDetailMongoRepository repositories.IStockBalanceDetailRepository
}

func NewStockReceiveTransactionConsumerService(repo IStockReceiveTransactionPGRepository, stockBalanceDetailMongoRepository repositories.IStockBalanceDetailRepository) IStockReceiveTransactionConsumerService {

	return &StockReceiveTransactionConsumerService{
		repo:                              repo,
		stockBalanceDetailMongoRepository: stockBalanceDetailMongoRepository,
	}
}

func (s *StockReceiveTransactionConsumerService) Upsert(holdingCode string, docNo string, doc models.StockBalanceTransactionPG) error {
	findDoc, err := s.repo.Get(holdingCode, docNo)
	if err != nil {
		err = s.repo.Create(doc)
		if err != nil {
			return err
		}
	} else {

		isEqual := findDoc.CompareTo(&doc)

		if !isEqual {
			err = s.repo.Update(holdingCode, docNo, doc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *StockReceiveTransactionConsumerService) Delete(holdingCode string, docNo string) error {
	err := s.repo.DeleteData(holdingCode, docNo, models.StockBalanceTransactionPG{
		TransactionPG: models.TransactionPG{
			HoldingCodeentity: pkgModels.HoldingCodeentity{
				HoldingCode: holdingCode,
			},
			DocNo: docNo,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *StockReceiveTransactionConsumerService) GetStockBalanceDetail(holdingCode string, docNo string) (*[]models.StockBalanceTransactionDetailPG, error) {

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
				Key:   "guid_fixed",
				Value: -1,
			},
		},
	}

	details := make([]models.StockBalanceTransactionDetailPG, 0)
	// docList := make([]stockBalanceModels.StockBalanceDetailInfo, 0)

	for {

		timeOut := time.Duration(600) * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeOut)
		defer cancel()

		docs, pages, err := s.stockBalanceDetailMongoRepository.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageRequest)
		if err != nil {
			return nil, err
		}

		//docList = append(docList, docs...)

		for _, doc := range docs {
			details = append(details, models.StockBalanceTransactionDetailPG{
				TransactionDetailPG: models.TransactionDetailPG{
					GuidFixed:           doc.GuidFixed,
					DocNo:               doc.DocNo,
					HoldingCode:         holdingCode,
					LineNumber:          int8(doc.LineNumber),
					DocRef:              doc.DocRef,
					Barcode:             doc.Barcode,
					ItemType:            doc.ItemType,
					ItemGuid:            doc.ItemGuid,
					VatType:             doc.VatType,
					TaxType:             doc.TaxType,
					StandValue:          doc.StandValue,
					DivideValue:         doc.DivideValue,
					WhCode:              doc.WhCode,
					LocationCode:        doc.LocationCode,
					Qty:                 doc.Qty,
					Price:               doc.Price,
					PriceExcludeVat:     doc.PriceExcludeVat,
					SumAmount:           doc.SumAmount,
					SumAmountExcludeVat: doc.SumAmountExcludeVat,
					Discount:            doc.Discount,
					DiscountAmount:      doc.DiscountAmount,
				},
			})
		}

		if pages.TotalPage > int64(pageRequest.Page) {
			pageRequest.Page++
		} else {
			break
		}
	}

	return &details, nil
}
