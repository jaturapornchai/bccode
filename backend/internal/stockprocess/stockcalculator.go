package stockprocess

import (
	"smlcloudplatform/internal/logger"
	productBarcodeRepositories "smlcloudplatform/internal/product/productbarcode/repositories"
	stockModel "smlcloudplatform/internal/stockprocess/models"
	"smlcloudplatform/internal/stockprocess/repositories"
	"smlcloudplatform/pkg/stockcalculator"
)

type IStockCalculator interface {
	// CalculateStockPrice(stockData []StockData) float64
	CalculatorStock(shopID string, barcode string) error
	GetStockDataList(shopID string, barcode string) ([]stockModel.StockData, error)
	WriteUpdateStockDataChanged(stockData []stockModel.StockData) error
}

type StockCalculator struct {
	stockMovementRepo  repositories.IStockProcessPGRepository
	productBarcodeRepo productBarcodeRepositories.IProductBarcodePGRepository
}

func NewStockCalculator(
	repo repositories.IStockProcessPGRepository,
	productBarcodeRepo productBarcodeRepositories.IProductBarcodePGRepository,
) IStockCalculator {

	return &StockCalculator{
		stockMovementRepo:  repo,
		productBarcodeRepo: productBarcodeRepo,
	}
}

func (sc *StockCalculator) CalculatorStock(shopID string, barcode string) error {

	stockDataList, err := sc.GetStockDataList(shopID, barcode)
	if err != nil {
		return err
	}

	productBarcode, err := sc.productBarcodeRepo.Get(shopID, barcode)
	if err != nil {
		return err
	}

	var stockDataChangeLists []stockModel.StockData

	if len(stockDataList) > 0 {
		calculator := stockcalculator.NewStockCalculator(shopID, barcode, 2, 0, 0)
		for i, data := range stockDataList {
			var costPerUnit, totalCost, balanceQty, balanceAmount, balanceAveragePerUnit float64
			if data.CalcFlag == 1 {

				if data.HasCostFromOtherDoc() {

					var qty, amountExcludeVat float64
					var costPerUnitOut float64
					qty = data.CalcQty

					for _, costData := range stockDataList {
						if data.TransFlag == 72 && costData.DocNo == data.DocNo && costData.LineNumber == data.LineNumber {
							amountExcludeVat = costData.TotalCost
							break
						} else if costData.DocNo == data.DocRef {
							costPerUnitOut = costData.CostPerUnit
							amountExcludeVat = costPerUnitOut * qty
							break
						}
					}

					costPerUnit, totalCost, balanceQty, balanceAmount, balanceAveragePerUnit = calculator.ApplyStock(qty, amountExcludeVat)

				} else if data.TransFlag == 66 {
					// ปรับปรุงเฉพาะจำนวน ใช้ต้นทุนปัจจุบัน
					costPerUnit, totalCost, balanceQty, balanceAmount, balanceAveragePerUnit = calculator.ApplyStockWithCurrentCost(data.CalcQty)
				} else if data.TransFlag == 866 {
					// ปรับปรุงเฉพาะต้นทุน
					costPerUnit, totalCost, balanceQty, balanceAmount, balanceAveragePerUnit = calculator.ApplyCost(data.SumAmountExcludeVat)
				} else {
					costPerUnit, totalCost, balanceQty, balanceAmount, balanceAveragePerUnit = calculator.ApplyStock(data.CalcQty, data.SumAmountExcludeVat)
				}

				isDataChange :=
					data.CostPerUnit != costPerUnit ||
						data.TotalCost != totalCost ||
						data.BalanceQty != balanceQty ||
						data.BalanceAmount != balanceAmount ||
						data.BalanceAverage != balanceAveragePerUnit

				if isDataChange {
					stockDataList[i].CostPerUnit = costPerUnit
					stockDataList[i].TotalCost = totalCost
					stockDataList[i].BalanceQty = balanceQty
					stockDataList[i].BalanceAmount = balanceAmount
					stockDataList[i].BalanceAverage = balanceAveragePerUnit

					stockDataChangeLists = append(stockDataChangeLists, stockDataList[i])
				}
				logger.GetLogger().Debugf("ApplyStock: %+v", stockDataList[i])
			} else {

				// var costPerUnit, totalCost, balanceQty, balanceAmount, balanceAveragePerUnit float64
				if data.HasCostFromOtherDoc() {

					var costOut float64

					for _, costData := range stockDataList {
						if costData.DocNo == data.DocRef {
							costOut = costData.CostPerUnit
							break
						}
					}
					// find cost from other doc
					costPerUnit, totalCost, balanceQty, balanceAmount, balanceAveragePerUnit = calculator.ReduceStockWithCost(data.CalcQty, costOut)
				} else if data.TransFlag == 868 {
					// ลดแบบไม่กระทบยอด
					costPerUnit, totalCost, balanceQty, balanceAmount, balanceAveragePerUnit = calculator.ReduceCost(data.SumAmountExcludeVat)
				} else {
					costPerUnit, totalCost, balanceQty, balanceAmount, balanceAveragePerUnit = calculator.ReduceStock(data.CalcQty)
				}

				isDataChange :=
					data.CostPerUnit != costPerUnit ||
						data.TotalCost != totalCost ||
						data.BalanceQty != balanceQty ||
						data.BalanceAmount != balanceAmount ||
						data.BalanceAverage != balanceAveragePerUnit

				if isDataChange {

					stockDataList[i].CostPerUnit = costPerUnit
					stockDataList[i].TotalCost = totalCost
					stockDataList[i].BalanceQty = balanceQty
					stockDataList[i].BalanceAmount = balanceAmount
					stockDataList[i].BalanceAverage = balanceAveragePerUnit

					stockDataChangeLists = append(stockDataChangeLists, stockDataList[i])
				}
				logger.GetLogger().Debugf("ReduceStock: %+v", stockDataList[i])
			}

			logger.GetLogger().Debugf("After Place Movement: %+v", calculator)
		}

		// write update movement
		if len(stockDataChangeLists) > 0 {
			err = sc.WriteUpdateStockDataChanged(stockDataChangeLists)
			if err != nil {
				return err
			}
		}

		isProductBalanceNotEqual := calculator.BalanceQty() != productBarcode.BalanceQty || calculator.BalanceAmount() != productBarcode.BalanceAmount || calculator.AverageCost() != productBarcode.AverageCost

		if isProductBalanceNotEqual {

			productBarcode.BalanceQty = calculator.BalanceQty()
			productBarcode.BalanceAmount = calculator.BalanceAmount()
			productBarcode.AverageCost = calculator.AverageCost()

			err = sc.productBarcodeRepo.Update(shopID, barcode, productBarcode)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func (sc *StockCalculator) GetStockDataList(shopID string, barcode string) ([]stockModel.StockData, error) {
	return sc.stockMovementRepo.GetStockTransactionList(shopID, barcode)
}

func (sc *StockCalculator) WriteUpdateStockDataChanged(stockData []stockModel.StockData) error {
	return sc.stockMovementRepo.UpdateStockTransactionChange(stockData)
}
