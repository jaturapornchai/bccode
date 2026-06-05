package repositories

import (
	stockModel "smlcloudplatform/internal/stockprocess/models"
	"smlcloudplatform/pkg/microservice"
)

type IStockProcessPGRepository interface {
	GetStockTransactionList(holdingCode string, barcode string) ([]stockModel.StockData, error)
	UpdateStockTransactionChange(stockData []stockModel.StockData) error

	ExecuteUpdateProductBarcodeStockBalance(holdingCode string, barcode string) error
}

type StockProcessPGRepository struct {
	pst microservice.IPersister
}

func NewStockProcessPGRepository(pst microservice.IPersister) IStockProcessPGRepository {
	return &StockProcessPGRepository{
		pst: pst,
	}
}

func (repo *StockProcessPGRepository) GetStockTransactionList(holdingCode string, barcode string) ([]stockModel.StockData, error) {

	var stockDatas []stockModel.StockData

	sql := `SELECT
	STKD.id,
	STK.holdingcode, STK.docno,  STK.docdate, STK.transflag, STK.inquirytype, STKD.docref
	, STKD.barcode, PDB.mainbarcoderef, STKD.unitcode, STKD.qty
	, PDB.standvalue, PDB.dividevalue
	, STKD.calcflag
	, ((STKD.qty*PDB.standvalue)/PDB.dividevalue) AS calcqty
	, STKD.price, STKD.sumamount
	, STKD.sumamountexcludevat, STKD.priceexcludevat
	, STKD.linenumber, STKD.sumofcost, STKD.averagecost
	, STKD.vattype ,STKD.taxtype
	, STKD.costperunit, STKD.totalcost
	, STKD.balanceamount, STKD.balanceaverage, STKD.balanceqty

	FROM stock_transaction AS STK
	JOIN stock_transaction_detail AS STKD on STKD.docno = STK.docno AND STKD.holdingcode = STK.holdingcode
	JOIN productbarcode AS PDB ON PDB.barcode = STKD.barcode AND STKD.holdingcode = PDB.holdingcode
	WHERE STK.holdingcode = @holdingcode AND STK.iscancel = false AND PDB.mainbarcoderef = (select mainbarcoderef from productbarcode where barcode = @barcode and productbarcode.holdingcode = STK.holdingcode )
	ORDER BY STK.docdate, STK.docno, STKD.calcflag, STKD.linenumber`

	//repo.pst.Where(&stockDatas, "holdingcode = ? AND barcode = ?", holdingCode, barcode)
	conditions := map[string]interface{}{
		"holdingcode": holdingCode,
		"barcode":     barcode,
	}
	_, err := repo.pst.Raw(sql, conditions, &stockDatas)

	if err != nil {
		return nil, err
	}
	return stockDatas, nil

}

func (repo *StockProcessPGRepository) UpdateStockTransactionChange(stockData []stockModel.StockData) error {

	err := repo.pst.Transaction(func(pst *microservice.Persister) error {

		for _, data := range stockData {
			err := pst.DBClient().Exec("UPDATE stock_transaction_detail "+
				" SET costperunit = ? , totalcost = ?, balanceqty = ?, balanceamount = ?, balanceaverage = ? "+
				" WHERE id = ?", data.CostPerUnit, data.TotalCost, data.BalanceQty, data.BalanceAmount, data.BalanceAverage, data.ID).Error
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

func (repo *StockProcessPGRepository) ExecuteUpdateProductBarcodeStockBalance(holdingCode string, barcode string) error {

	sql := `WITH stock AS (select
		barcode, holdingcode, balanceqty
		, (SELECT SUM(STKD.qty*calcflag) FROM stock_transaction_detail AS STKD
			WHERE STKD.holdingcode = productbarcode.holdingcode AND STKD.barcode = productbarcode.barcode) as trx_balance_qty
		from productbarcode
		where holdingcode= @holdingcode and barcode = @barcode
		)
		UPDATE productbarcode set balanceqty = stock.trx_balance_qty
		FROM stock  WHERE productbarcode.barcode = stock.barcode AND productbarcode.holdingcode= stock.holdingcode AND productbarcode.balanceqty <>  stock.trx_balance_qty
		 `
	conditions := map[string]interface{}{
		"holdingcode": holdingCode,
		"barcode":     barcode,
	}

	err := repo.pst.Exec(sql, conditions)
	if err != nil {
		return err
	}
	return nil
}
