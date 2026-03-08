package usecases

import "smlcloudplatform/internal/transaction/models"

type ITransactionPhaser[T any] interface {
	PhaseSingleDoc(input string) (*T, error)
	PhaseMultipleDoc(input string) (*[]T, error)
}

type IStockTransactionPhaser[T any] interface {
	PhaseSingleDoc(doc T) (*models.StockTransaction, error)
}

type ICreditorTransactionPhaser[T any] interface {
	PhaseSingleDoc(doc T) (*models.CreditorTransactionPG, error)
	// PhaseMultipleDoc(doc T) (*[]models.CreditorTransactionPG, error)
}

type IDebtorTransactionPhaser[T any] interface {
	PhaseSingleDoc(doc T) (*models.DebtorTransactionPG, error)
	// PhaseMultipleDoc(doc T) (*[]models.DebtorTransactionPG, error)
}
