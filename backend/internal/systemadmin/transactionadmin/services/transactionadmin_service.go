package services

type ITransactionAdminService interface{}

type TransactionAdminService struct{}

func NewTransactionAdminService() ITransactionAdminService {
	return &TransactionAdminService{}
}
