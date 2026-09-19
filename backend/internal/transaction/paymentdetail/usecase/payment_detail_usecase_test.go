package usecase_test

import (
	"errors"
	"smlcloudplatform/internal/transaction/paymentdetail/models"
	"smlcloudplatform/internal/transaction/paymentdetail/usecase"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (mock *MockRepository) Get(holdingCode string, docNo string) (*models.TransactionPaymentDetail, error) {
	args := mock.Called(holdingCode, docNo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransactionPaymentDetail), args.Error(1)
}

func (mock *MockRepository) Create(doc models.TransactionPaymentDetail) error {
	args := mock.Called(doc)
	return args.Error(0)
}

func (mock *MockRepository) Update(holdingCode string, docNo string, doc models.TransactionPaymentDetail) error {
	args := mock.Called(holdingCode, docNo, doc)
	return args.Error(0)
}

func (mock *MockRepository) Delete(holdingCode string, docNo string, doc models.TransactionPaymentDetail) error {
	args := mock.Called(holdingCode, docNo, doc)
	return args.Error(0)
}

func TestUpsert(t *testing.T) {
	t.Run("error on get", func(t *testing.T) {
		mockRepo := new(MockRepository)
		u := usecase.NewPaymentDetailUsecase(mockRepo)
		mockRepo.On("Get", "shop1", "doc1").Return(nil, errors.New("record not found"))
		mockRepo.On("Create", mock.Anything).Return(nil)

		err := u.Upsert("shop1", "doc1", models.TransactionPaymentDetail{})
		assert.NoError(t, err)
	})

	t.Run("no error on get", func(t *testing.T) {
		mockRepo := new(MockRepository)
		u := usecase.NewPaymentDetailUsecase(mockRepo)
		mockRepo.On("Get", "shop1", "doc1").Return(&models.TransactionPaymentDetail{Amount: 100}, nil)
		mockRepo.On("Update", "shop1", "doc1", mock.Anything).Return(nil)

		err := u.Upsert("shop1", "doc1", models.TransactionPaymentDetail{Amount: 200})
		assert.NoError(t, err)
	})
}

func TestDelete(t *testing.T) {
	mockRepo := new(MockRepository)
	u := usecase.NewPaymentDetailUsecase(mockRepo)

	mockRepo.On("Delete", "shop1", "doc1", mock.Anything).Return(nil)

	err := u.Delete("shop1", "doc1")
	assert.NoError(t, err)
}
