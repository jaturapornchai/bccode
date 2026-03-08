package repositories

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/debtaccount/debtor/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
)

type IPointTransactionRepository interface {
	Create(ctx context.Context, doc models.PointTransactionDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.PointTransactionDoc) error
	UpdateDebtorPointBalance(ctx context.Context, shopID string, customerCode string, pointAmount float64) error
	UpdateDebtorPointBalanceByCode(ctx context.Context, shopID string, custCode string, pointAmount float64) error
	UpdateDebtorPointBalanceByPointsCode(ctx context.Context, shopID string, pointsCode string, pointAmount float64) error
	FindPointTransactionsByDebtorCode(ctx context.Context, shopID string, debtorCode string, pageableStep micromodels.PageableStep) ([]models.PointTransactionInfo, int, error)
	FindPointTransactionsByPointsCode(ctx context.Context, shopID string, pointsCode string, pageableStep micromodels.PageableStep) ([]models.PointTransactionInfo, int, error)
	DeletePointTransactionsByDocNo(ctx context.Context, shopID string, docNo string, authUsername string) error
	DeleteManualPointTransaction(ctx context.Context, shopID string, pointsCode string, docNo string, authUsername string) error
	RecalculatePointBalanceByPointsCode(ctx context.Context, shopID string, pointsCode string) error
}

type PointTransactionRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.PointTransactionDoc]
	repositories.SearchRepository[models.PointTransactionInfo]
	repositories.GuidRepository[models.PointTransactionItemGuid]
	repositories.ActivityRepository[models.PointTransactionActivity, models.PointTransactionDeleteActivity]
}

func NewPointTransactionRepository(pst microservice.IPersisterMongo) *PointTransactionRepository {
	insRepo := &PointTransactionRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.PointTransactionDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.PointTransactionInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.PointTransactionItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.PointTransactionActivity, models.PointTransactionDeleteActivity](pst)

	return insRepo
}

func (repo PointTransactionRepository) UpdateDebtorPointBalance(ctx context.Context, shopID string, customerCode string, pointAmount float64) error {
	// This is a generic function that tries both "code" and "pointscode" fields
	// for backward compatibility and UsePoint/GetPoint scenarios
	debtorRepo := NewDebtorRepository(repo.pst)

	// Try to find debtor by code first (for UsePoint - traditional customer lookup)
	debtor, err := debtorRepo.FindByDocIndentityGuid(ctx, shopID, "code", customerCode)
	if err != nil {
		return err
	}

	// If not found by code, try by pointscode (for GetPoint - new point system)
	if debtor.GuidFixed == "" {
		debtor, err = debtorRepo.FindByDocIndentityGuid(ctx, shopID, "pointscode", customerCode)
		if err != nil {
			return err
		}
	}

	// Check if debtor exists
	if debtor.GuidFixed == "" {
		return fmt.Errorf("debtor not found with code or pointscode: %s", customerCode)
	}

	// Update point balance
	debtor.PointBalance += pointAmount
	if debtor.PointBalance < 0 {
		debtor.PointBalance = 0
	}

	// Save updated debtor
	err = debtorRepo.Update(ctx, shopID, debtor.GuidFixed, debtor)
	return err
}

// UpdateDebtorPointBalanceByCode updates point balance by searching with debtor.code (for UsePoint)
func (repo PointTransactionRepository) UpdateDebtorPointBalanceByCode(ctx context.Context, shopID string, custCode string, pointAmount float64) error {
	debtorRepo := NewDebtorRepository(repo.pst)

	// Find debtor by code (for UsePoint - traditional customer lookup)
	debtor, err := debtorRepo.FindByDocIndentityGuid(ctx, shopID, "code", custCode)
	if err != nil {
		return err
	}

	// Check if debtor exists
	if debtor.GuidFixed == "" {
		return fmt.Errorf("debtor not found with code: %s", custCode)
	}

	// Update point balance
	debtor.PointBalance += pointAmount
	if debtor.PointBalance < 0 {
		debtor.PointBalance = 0
	}

	// Save updated debtor
	err = debtorRepo.Update(ctx, shopID, debtor.GuidFixed, debtor)
	return err
}

// UpdateDebtorPointBalanceByPointsCode updates point balance by searching with debtor.pointscode (for GetPoint)
func (repo PointTransactionRepository) UpdateDebtorPointBalanceByPointsCode(ctx context.Context, shopID string, pointsCode string, pointAmount float64) error {
	debtorRepo := NewDebtorRepository(repo.pst)

	// Find debtor by pointscode (for GetPoint - new point system)
	debtor, err := debtorRepo.FindByDocIndentityGuid(ctx, shopID, "pointscode", pointsCode)
	if err != nil {
		return err
	}

	// Check if debtor exists
	if debtor.GuidFixed == "" {
		return fmt.Errorf("debtor not found with pointscode: %s", pointsCode)
	}

	// Update point balance
	debtor.PointBalance += pointAmount
	if debtor.PointBalance < 0 {
		debtor.PointBalance = 0
	}

	// Save updated debtor
	err = debtorRepo.Update(ctx, shopID, debtor.GuidFixed, debtor)
	return err
}

func (repo PointTransactionRepository) FindPointTransactionsByDebtorCode(ctx context.Context, shopID string, debtorCode string, pageableStep micromodels.PageableStep) ([]models.PointTransactionInfo, int, error) {
	filters := map[string]interface{}{
		"debtorcode": debtorCode,
	}

	searchInFields := []string{
		"transactiondocno",
		"description",
	}

	selectFields := map[string]interface{}{}

	// Add default sorting by TransactionDate in descending order (latest first) if no sorts are provided
	if len(pageableStep.Sorts) == 0 {
		pageableStep.Sorts = []micromodels.KeyInt{
			{Key: "transactiondate", Value: -1},
		}
	}

	docList, total, err := repo.SearchRepository.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.PointTransactionInfo{}, 0, err
	}

	return docList, total, nil
}

func (repo PointTransactionRepository) FindPointTransactionsByPointsCode(ctx context.Context, shopID string, pointsCode string, pageableStep micromodels.PageableStep) ([]models.PointTransactionInfo, int, error) {
	filters := map[string]interface{}{
		"pointscode": pointsCode,
	}

	searchInFields := []string{
		"transactiondocno",
		"description",
	}

	selectFields := map[string]interface{}{}

	// Add default sorting by TransactionDate in descending order (latest first) if no sorts are provided
	if len(pageableStep.Sorts) == 0 {
		pageableStep.Sorts = []micromodels.KeyInt{
			{Key: "transactiondate", Value: -1},
		}
	}

	docList, total, err := repo.SearchRepository.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.PointTransactionInfo{}, 0, err
	}

	return docList, total, nil
}

// DeletePointTransactionsByDocNo deletes all point transactions related to a specific document number
func (repo PointTransactionRepository) DeletePointTransactionsByDocNo(ctx context.Context, shopID string, docNo string, authUsername string) error {
	filters := map[string]interface{}{
		"transactiondocno": docNo,
	}

	err := repo.CrudRepository.Delete(ctx, shopID, authUsername, filters)
	return err
}

// DeleteManualPointTransaction deletes a manual point transaction (TransactionType 3) and recalculates balance
func (repo PointTransactionRepository) DeleteManualPointTransaction(ctx context.Context, shopID string, pointsCode string, docNo string, authUsername string) error {
	// Find the transaction first to validate it's a manual transaction
	filters := map[string]interface{}{
		"transactiondocno": docNo,
		"pointscode":       pointsCode,
	}

	searchInFields := []string{}
	selectFields := map[string]interface{}{}
	pageableStep := micromodels.PageableStep{
		Skip:  0,
		Limit: 1,
	}

	transactions, _, err := repo.SearchRepository.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)
	if err != nil {
		return fmt.Errorf("failed to find transaction: %w", err)
	}

	if len(transactions) == 0 {
		return fmt.Errorf("transaction not found with docNo: %s and pointsCode: %s", docNo, pointsCode)
	}

	transaction := transactions[0]

	// Validate that it's a manual transaction (Type 3)
	if transaction.TransactionType != 3 {
		return fmt.Errorf("only manual point transactions (type 3) can be deleted, found type: %d", transaction.TransactionType)
	}

	// Delete the transaction
	err = repo.DeletePointTransactionsByDocNo(ctx, shopID, docNo, authUsername)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	// Recalculate balance
	err = repo.RecalculatePointBalanceByPointsCode(ctx, shopID, pointsCode)
	if err != nil {
		return fmt.Errorf("failed to recalculate balance: %w", err)
	}

	return nil
}

// RecalculatePointBalanceByPointsCode recalculates point balance by summing all point transactions for a specific pointsCode
func (repo PointTransactionRepository) RecalculatePointBalanceByPointsCode(ctx context.Context, shopID string, pointsCode string) error {
	debtorRepo := NewDebtorRepository(repo.pst)

	// Find debtor by pointscode
	debtor, err := debtorRepo.FindByDocIndentityGuid(ctx, shopID, "pointscode", pointsCode)
	if err != nil {
		return err
	}

	// Check if debtor exists
	if debtor.GuidFixed == "" {
		return fmt.Errorf("debtor not found with pointscode: %s", pointsCode)
	}

	// Get all point transactions for this pointscode
	pageableStep := micromodels.PageableStep{
		Skip:  0,
		Limit: 10000, // Get all transactions (adjust if needed)
		Sorts: []micromodels.KeyInt{
			{Key: "transactiondate", Value: 1}, // Ascending order for calculation
		},
	}

	transactions, _, err := repo.FindPointTransactionsByPointsCode(ctx, shopID, pointsCode, pageableStep)
	if err != nil {
		return err
	}

	// Calculate total balance by summing all point amounts
	var calculatedBalance float64 = 0.0
	for _, transaction := range transactions {
		calculatedBalance += transaction.PointAmount
	}

	// Ensure balance is not negative
	if calculatedBalance < 0 {
		calculatedBalance = 0
	}

	// Update debtor's point balance
	debtor.PointBalance = calculatedBalance
	err = debtorRepo.Update(ctx, shopID, debtor.GuidFixed, debtor)
	return err
}
