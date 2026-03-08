package usecase

type ITrialBalanceSheetReportUsecase interface {
	IsAmountDebitSide(accountCategory int16, amount float64) bool
	DisplayAmount(amount float64) float64
}

type TrialBalanceSheetReportUsecase struct{}

func (u *TrialBalanceSheetReportUsecase) IsAmountDebitSide(accountCategory int16, amount float64) bool {
	if accountCategory == 1 || accountCategory == 5 {
		// debit side default
		if amount > 0 {
			return true
		}
	} else {
		if amount < 0 {
			return true
		}
	}
	return false
}

func (u *TrialBalanceSheetReportUsecase) DisplayAmount(amount float64) float64 {
	if amount < 0 {
		return amount * -1
	}
	return amount
}
