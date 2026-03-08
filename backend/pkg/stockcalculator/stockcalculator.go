package stockcalculator

import (
	"math"
	"smlcloudplatform/pkg/round"
)

type IStockCalculator interface {
	ApplyStock(qty float64, totalCostExcludeVat float64) (costPerUnit float64, totalCost float64, balanceQty float64, balanceAmount float64, balanceAverageCost float64)
	ApplyStockWithCurrentCost(qty float64) (costPerUnit float64, totalCost float64, balanceQty float64, balanceAmount float64, balanceAverageCost float64)
	ApplyCost(totalCostExcludeVat float64) (costPerUnit float64, totalCost float64, balanceQty float64, balanceAmount float64, balanceAverageCost float64)
	ReduceStock(qty float64) (costPerUnit float64, totalCost float64, balanceQty float64, balanceAmount float64, balanceAverageCost float64)
	ReduceStockWithCost(qty float64, cost float64) (costPerUnit float64, totalCost float64, balanceQty float64, balanceAmount float64, balanceAverageCost float64)
	ReduceCost(cost float64) (costPerUnit float64, totalCost float64, balanceQty float64, balanceAmount float64, balanceAverageCost float64)
	BalanceAmount() float64
	BalanceQty() float64
	AverageCost() float64
	AverageCostCalc(balanceAmount float64, balanceQty float64) float64
}

type StockCalculator struct {
	ShopID        string
	Barcode       string
	AmountDigit   int8
	balanceQty    float64
	balanceAmount float64
	averageCost   float64
}

func NewStockCalculator(shopID string, barcode string, amountDigit int8, balanceQtyFirst float64, balanceAmountFirst float64) IStockCalculator {

	if amountDigit <= 0 {
		amountDigit = 2
	}

	averageCost := round.Round(balanceAmountFirst/balanceQtyFirst, amountDigit)
	if math.IsNaN(averageCost) {
		averageCost = 0
	}

	return &StockCalculator{
		ShopID:        shopID,
		Barcode:       barcode,
		AmountDigit:   amountDigit,
		balanceQty:    balanceQtyFirst,
		balanceAmount: balanceAmountFirst,
		averageCost:   averageCost,
	}
}

func (sc *StockCalculator) ApplyStock(qty float64, totalCostExcludeVat float64) (costPerUnit float64, totalCost float64, balanceQty float64, balanceAmount float64, balanceAverageCost float64) {
	totalCostExcludeVat = round.Round(totalCostExcludeVat, sc.AmountDigit)

	//sc.balanceQty += qty
	sc.balanceQty = round.Round(sc.balanceQty+qty, sc.AmountDigit)

	// sc.balanceAmount += totalCostExcludeVat
	sc.balanceAmount = round.Round(sc.balanceAmount+totalCostExcludeVat, sc.AmountDigit)
	average := 0.0
	if qty != 0 {
		average = round.Round(totalCostExcludeVat/qty, sc.AmountDigit)
	}

	if sc.balanceQty != 0 {
		sc.averageCost = sc.AverageCostCalc(sc.balanceAmount, sc.balanceQty) // round.Round(sc.balanceAmount/sc.balanceQty, sc.AmountDigit)
	} else {
		sc.averageCost = 0
		sc.balanceAmount = 0
	}

	return average, totalCostExcludeVat, sc.balanceQty, sc.balanceAmount, sc.averageCost
}

func (sc *StockCalculator) ApplyStockWithCurrentCost(qty float64) (costPerUnit float64, totalCost float64, balanceQty float64, balanceAmount float64, balanceAverageCost float64) {
	perUnitCost := sc.averageCost
	applyCost := round.Round(perUnitCost*qty, sc.AmountDigit)

	sc.balanceQty = round.Round(sc.balanceQty+qty, sc.AmountDigit)
	sc.balanceAmount = round.Round(sc.balanceAmount+applyCost, sc.AmountDigit)
	sc.averageCost = sc.AverageCostCalc(sc.balanceAmount, sc.balanceQty)

	if sc.balanceQty == 0 {
		sc.averageCost = 0
		sc.balanceAmount = 0
	}

	return perUnitCost, applyCost, sc.balanceQty, sc.balanceAmount, sc.averageCost
}

func (sc *StockCalculator) ApplyCost(totalCostExcludeVat float64) (costPerUnit float64, totalCost float64, balanceQty float64, balanceAmount float64, balanceAverageCost float64) {

	sc.balanceAmount = round.Round(sc.balanceAmount+totalCostExcludeVat, sc.AmountDigit)
	sc.averageCost = round.Round(sc.balanceAmount/sc.balanceQty, sc.AmountDigit)

	return 0.0, totalCostExcludeVat, sc.balanceQty, sc.balanceAmount, sc.averageCost
}

func (sc *StockCalculator) ReduceStock(qty float64) (costPerUnit float64, totalCost float64, balanceQty float64, balanceAmount float64, balanceAverageCost float64) {
	// sc.balanceQty -= qty
	sc.balanceQty = round.Round(sc.balanceQty-qty, sc.AmountDigit)

	averageCostOut := sc.averageCost
	sumOfCostOut := averageCostOut * qty
	costOut := round.Round(sumOfCostOut, sc.AmountDigit)
	// sc.balanceAmount -= costOut

	if sc.balanceQty == 0 {
		costOut = sc.balanceAmount
		sc.balanceAmount = round.Round(sc.balanceAmount-costOut, sc.AmountDigit)
		sc.averageCost = 0
	} else {
		sc.balanceAmount = round.Round(sc.balanceAmount-costOut, sc.AmountDigit)
		sc.averageCost = sc.AverageCostCalc(sc.balanceAmount, sc.balanceQty)
	}

	return averageCostOut, costOut, sc.balanceQty, sc.balanceAmount, sc.averageCost
}

func (sc *StockCalculator) ReduceStockWithCost(qty float64, cost float64) (costPerUnit float64, totalCost float64, balanceQty float64, balanceAmount float64, balanceAverageCost float64) {

	// sc.balanceQty -= qty
	sc.balanceQty = round.Round(sc.balanceQty-qty, sc.AmountDigit)

	averageCostOut := cost
	costOut := round.Round(averageCostOut*qty, sc.AmountDigit)
	// sc.balanceAmount -= costOut

	if sc.balanceQty == 0 {
		costOut = sc.balanceAmount
		sc.balanceAmount = round.Round(sc.balanceAmount-costOut, sc.AmountDigit)
		sc.averageCost = 0
	} else {
		sc.balanceAmount = round.Round(sc.balanceAmount-costOut, sc.AmountDigit)
		sc.averageCost = sc.AverageCostCalc(sc.balanceAmount, sc.balanceQty)
	}

	return averageCostOut, costOut, sc.balanceQty, sc.balanceAmount, sc.averageCost
}

func (sc *StockCalculator) ReduceCost(cost float64) (costPerUnit float64, totalCost float64, balanceQty float64, balanceAmount float64, balanceAverageCost float64) {

	sc.balanceAmount = round.Round(sc.balanceAmount-cost, sc.AmountDigit)
	sc.averageCost = sc.AverageCostCalc(sc.balanceAmount, sc.balanceQty)

	return 0.0, cost, sc.balanceQty, sc.balanceAmount, sc.averageCost
}

func (sc *StockCalculator) BalanceAmount() float64 {
	return sc.balanceAmount
}

func (sc *StockCalculator) BalanceQty() float64 {
	return sc.balanceQty
}

func (sc *StockCalculator) AverageCost() float64 {
	return sc.averageCost
}

func (sc *StockCalculator) AverageCostCalc(balanceAmount float64, balanceQty float64) float64 {

	if balanceQty == 0 || balanceAmount == 0 {
		return 0.0
	}

	balanceAmountCalc := balanceAmount
	if balanceAmountCalc < 0 {
		balanceAmountCalc = balanceAmountCalc * -1
	}

	balanceQtyCalc := balanceQty
	if balanceQtyCalc < 0 {
		balanceQtyCalc = balanceQtyCalc * -1
	}

	averageCost := round.Round(balanceAmountCalc/balanceQtyCalc, sc.AmountDigit)
	return averageCost
}
