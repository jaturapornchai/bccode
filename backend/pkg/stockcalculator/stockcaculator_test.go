package stockcalculator_test

import (
	"smlcloudplatform/pkg/stockcalculator"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStockCalculator(t *testing.T) {

	calculator := stockcalculator.NewStockCalculator("TESTSHOP", "TESTBARCODE", 2, 0, 0)

	assert.Equal(t, 0.0, calculator.BalanceAmount())
	assert.Equal(t, 0.0, calculator.BalanceQty())
	assert.Equal(t, 0.0, calculator.AverageCost())

	calculator.ApplyStock(100, 1000)
	assert.Equal(t, 1000.0, calculator.BalanceAmount())
	assert.Equal(t, 100.0, calculator.BalanceQty())
	assert.Equal(t, 10.0, calculator.AverageCost())

	calculator.ApplyStock(2, 22)
	assert.Equal(t, 1022.0, calculator.BalanceAmount())
	assert.Equal(t, 102.0, calculator.BalanceQty())
	assert.Equal(t, 10.02, calculator.AverageCost())

}

func TestStockCalculator2(t *testing.T) {

	calculator := stockcalculator.NewStockCalculator("TESTSHOP", "TESTBARCODE", 2, 0, 0)

	assert.Equal(t, 0.0, calculator.BalanceAmount())
	assert.Equal(t, 0.0, calculator.BalanceQty())
	assert.Equal(t, 0.0, calculator.AverageCost())

	calculator.ApplyStock(100, 1000)
	assert.Equal(t, 1000.0, calculator.BalanceAmount())
	assert.Equal(t, 100.0, calculator.BalanceQty())
	assert.Equal(t, 10.0, calculator.AverageCost())

	calculator.ApplyStock(2, 22)
	assert.Equal(t, 1022.0, calculator.BalanceAmount())
	assert.Equal(t, 102.0, calculator.BalanceQty())
	assert.Equal(t, 10.02, calculator.AverageCost())

	calculator.ApplyStock(3, 30)
	assert.Equal(t, 1052.0, calculator.BalanceAmount())
	assert.Equal(t, 105.0, calculator.BalanceQty())
	assert.Equal(t, 10.02, calculator.AverageCost())

}

func TestApplyStockWithoutCost(t *testing.T) {

	calculator := stockcalculator.NewStockCalculator("TESTSHOP", "TESTBARCODE", 2, 292.00, 1864.77)
	assert.Equal(t, 6.39, calculator.AverageCost())
	assert.Equal(t, 1864.77, calculator.BalanceAmount())
	assert.Equal(t, 292.00, calculator.BalanceQty())

	calculator.ApplyStock(5, 0)

	assert.Equal(t, 1864.77, calculator.BalanceAmount())
	assert.Equal(t, float64(297), calculator.BalanceQty())
	assert.Equal(t, 6.28, calculator.AverageCost())

}

func TestStockCalculatorApplyCost(t *testing.T) {

	calculator := stockcalculator.NewStockCalculator("TESTSHOP", "TESTBARCODE", 2, 129.00, 828.76)
	assert.Equal(t, 6.42, calculator.AverageCost())

	// ApplyCost
	calculator.ApplyCost(3.59)
	assert.Equal(t, 832.35, calculator.BalanceAmount())
	assert.Equal(t, float64(129), calculator.BalanceQty())
	assert.Equal(t, 6.45, calculator.AverageCost())
}

func TestApplyStockWithCurentCost(t *testing.T) {

	calculator := stockcalculator.NewStockCalculator("TESTSHOP", "TESTBARCODE", 2, 147, 939.57)
	assert.Equal(t, 6.39, calculator.AverageCost())

	// ApplyStock
	calculator.ApplyStockWithCurrentCost(30)
	assert.Equal(t, 1131.27, calculator.BalanceAmount())
	assert.Equal(t, 177.00, calculator.BalanceQty())
	assert.Equal(t, 6.39, calculator.AverageCost())

}

func TestReduceStock(t *testing.T) {
	calculator := stockcalculator.NewStockCalculator("TESTSHOP", "TESTBARCODE", 2, 177, 1131.87)
	assert.Equal(t, 6.39, calculator.AverageCost())

	calculator.ReduceStock(30)

	assert.Equal(t, 940.17, calculator.BalanceAmount())
	assert.Equal(t, 147.00, calculator.BalanceQty())
	assert.Equal(t, 6.40, calculator.AverageCost())
}

func TestReduceCost(t *testing.T) {

	calculator := stockcalculator.NewStockCalculator("TESTSHOP", "TESTBARCODE", 2, 177, 1131.87)
	assert.Equal(t, 6.39, calculator.AverageCost())

	calculator.ReduceCost(191.70)

	assert.Equal(t, 940.17, calculator.BalanceAmount())
	assert.Equal(t, 177.00, calculator.BalanceQty())
	assert.Equal(t, 5.31, calculator.AverageCost())
}

func TestRecudeStockWithCost(t *testing.T) {

	calculator := stockcalculator.NewStockCalculator("TESTSHOP", "TESTBARCODE", 2, 297, 1864.77)
	assert.Equal(t, 6.28, calculator.AverageCost())

	calculator.ReduceStockWithCost(60, 6.39)

	assert.Equal(t, 1481.37, calculator.BalanceAmount())
	assert.Equal(t, 237.00, calculator.BalanceQty())
	assert.Equal(t, 6.25, calculator.AverageCost())
}
