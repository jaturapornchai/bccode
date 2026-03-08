package usecases

import (
	"encoding/json"
	"errors"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	saleReturnModel "smlcloudplatform/internal/transaction/saleinvoicereturn/models"
)

type StockTransactionSaleInvoicePhaser struct{}

func (p *StockTransactionSaleInvoicePhaser) PhaseSaleReturnDoc(doc saleReturnModel.SaleInvoiceReturnDoc) (*models.StockTransaction, error) {

	details := make([]models.StockTransactionDetail, len(*doc.Transaction.Details))

	for i, detail := range *doc.Transaction.Details {
		stockDetail := models.StockTransactionDetail{
			Barcode:    detail.Barcode,
			Qty:        detail.Qty,
			Price:      detail.Price,
			SumAmount:  detail.SumAmount,
			CalcFlag:   1,
			LineNumber: int8(i),
		}
		details[i] = stockDetail
	}

	stockTransaction := models.StockTransaction{
		ShopIdentity: pkgModels.ShopIdentity{
			ShopID: doc.ShopID,
		},
		TransFlag: 48,
		DocNo:     doc.DocNo,
		DocDate:   doc.DocDatetime,
		Details:   &details,
	}
	return &stockTransaction, nil
}

func (p StockTransactionSaleInvoicePhaser) PhaseSingleDoc(msg string) (*models.StockTransaction, error) {

	doc := saleReturnModel.SaleInvoiceReturnDoc{}
	err := json.Unmarshal([]byte(msg), &doc)
	if err != nil {
		return nil, errors.New("Cannot Unmarshal PurchaseDoc Message : " + err.Error())
	}
	trx, err := p.PhaseSaleReturnDoc(doc)
	if err != nil {
		return nil, errors.New("Error on Convert PurchaseDoc to StockTransaction : " + err.Error())
	}
	return trx, err
}

func (p StockTransactionSaleInvoicePhaser) PhaseMultipleDoc(input string) (*[]models.StockTransaction, error) {

	docs := []saleReturnModel.SaleInvoiceReturnDoc{}
	err := json.Unmarshal([]byte(input), &docs)
	if err != nil {
		return nil, errors.New("Cannot Unmarshal PurchaseDoc Message : " + err.Error())
	}

	stockTransactions := make([]models.StockTransaction, len(docs))

	for i, doc := range docs {
		trx, err := p.PhaseSaleReturnDoc(doc)
		if err != nil {
			return nil, errors.New("Error on Convert PurchaseDoc to StockTransaction : " + err.Error())
		}
		stockTransactions[i] = *trx
	}

	return &stockTransactions, nil
}
