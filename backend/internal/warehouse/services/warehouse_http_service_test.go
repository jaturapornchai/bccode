package services_test

import (
	"context"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/warehouse/models"
	"smlcloudplatform/internal/warehouse/services"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"testing"
	"time"

	"github.com/smlsoft/mongopagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockWarehouseRepository mocks the WarehouseRepository
type MockWarehouseRepository struct {
	mock.Mock
}

func (m *MockWarehouseRepository) Count(ctx context.Context, holdingCode string) (int, error) {
	args := m.Called(ctx, holdingCode)
	return args.Int(0), args.Error(1)
}

func (m *MockWarehouseRepository) Create(ctx context.Context, doc models.WarehouseDoc) (string, error) {
	args := m.Called(ctx, doc)
	return args.String(0), args.Error(1)
}

func (m *MockWarehouseRepository) CreateInBatch(ctx context.Context, docList []models.WarehouseDoc) error {
	args := m.Called(ctx, docList)
	return args.Error(0)
}

func (m *MockWarehouseRepository) Update(ctx context.Context, holdingCode, guid string, doc models.WarehouseDoc) error {
	args := m.Called(ctx, holdingCode, guid, doc)
	return args.Error(0)
}

func (m *MockWarehouseRepository) DeleteByGuidfixed(ctx context.Context, holdingCode, guid, username string) error {
	args := m.Called(ctx, holdingCode, guid, username)
	return args.Error(0)
}

func (m *MockWarehouseRepository) Delete(ctx context.Context, holdingCode, username string, filters map[string]interface{}) error {
	args := m.Called(ctx, holdingCode, username, filters)
	return args.Error(0)
}

func (m *MockWarehouseRepository) FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.WarehouseInfo, mongopagination.PaginationData, error) {
	args := m.Called(ctx, holdingCode, searchInFields, pageable)
	return args.Get(0).([]models.WarehouseInfo), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *MockWarehouseRepository) Find(ctx context.Context, holdingCode string, searchInFields []string, q string) ([]models.WarehouseInfo, error) {
	args := m.Called(ctx, holdingCode, searchInFields, q)
	return args.Get(0).([]models.WarehouseInfo), args.Error(1)
}

func (m *MockWarehouseRepository) FindByGuid(ctx context.Context, holdingCode, guid string) (models.WarehouseDoc, error) {
	args := m.Called(ctx, holdingCode, guid)
	return args.Get(0).(models.WarehouseDoc), args.Error(1)
}

func (m *MockWarehouseRepository) FindInItemGuid(ctx context.Context, holdingCode, columnName string, itemGuidList []string) ([]models.WarehouseItemGuid, error) {
	args := m.Called(ctx, holdingCode, columnName, itemGuidList)
	return args.Get(0).([]models.WarehouseItemGuid), args.Error(1)
}

func (m *MockWarehouseRepository) FindByDocIndentityGuid(ctx context.Context, holdingCode, indentityField string, indentityValue interface{}) (models.WarehouseDoc, error) {
	args := m.Called(ctx, holdingCode, indentityField, indentityValue)
	return args.Get(0).(models.WarehouseDoc), args.Error(1)
}

func (m *MockWarehouseRepository) FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.WarehouseInfo, mongopagination.PaginationData, error) {
	args := m.Called(ctx, holdingCode, filters, searchInFields, pageable)
	return args.Get(0).([]models.WarehouseInfo), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *MockWarehouseRepository) FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.WarehouseInfo, int, error) {
	args := m.Called(ctx, holdingCode, filters, searchInFields, projects, pageableLimit)
	return args.Get(0).([]models.WarehouseInfo), args.Int(1), args.Error(2)
}

func (m *MockWarehouseRepository) FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.WarehouseDeleteActivity, mongopagination.PaginationData, error) {
	args := m.Called(ctx, holdingCode, lastUpdatedDate, filters, pageable)
	return args.Get(0).([]models.WarehouseDeleteActivity), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *MockWarehouseRepository) FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.WarehouseActivity, mongopagination.PaginationData, error) {
	args := m.Called(ctx, holdingCode, lastUpdatedDate, filters, pageable)
	return args.Get(0).([]models.WarehouseActivity), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *MockWarehouseRepository) FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.WarehouseDeleteActivity, error) {
	args := m.Called(ctx, holdingCode, lastUpdatedDate, filters, pageableStep)
	return args.Get(0).([]models.WarehouseDeleteActivity), args.Error(1)
}

func (m *MockWarehouseRepository) FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.WarehouseActivity, error) {
	args := m.Called(ctx, holdingCode, lastUpdatedDate, filters, pageableStep)
	return args.Get(0).([]models.WarehouseActivity), args.Error(1)
}

func (m *MockWarehouseRepository) Transaction(ctx context.Context, queryFunc func(ctx context.Context) error) error {
	args := m.Called(ctx, queryFunc)
	return args.Error(0)
}

func (m *MockWarehouseRepository) EnsureIndexes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// noopWarehouseMessageQueueRepository satisfies IWarehouseMessageQueueRepository so the service's
// fire-and-forget `go func() { ... svc.repoMq.Update(...) }()` post-save calls have a real (non-nil
// interface) receiver instead of panicking on a nil interface method call.
type noopWarehouseMessageQueueRepository struct{}

func (noopWarehouseMessageQueueRepository) Create(doc models.WarehouseDoc) error { return nil }
func (noopWarehouseMessageQueueRepository) Update(doc models.WarehouseDoc) error { return nil }
func (noopWarehouseMessageQueueRepository) Delete(doc models.WarehouseDoc) error { return nil }
func (noopWarehouseMessageQueueRepository) CreateInBatch(docList []models.WarehouseDoc) error {
	return nil
}
func (noopWarehouseMessageQueueRepository) UpdateInBatch(docList []models.WarehouseDoc) error {
	return nil
}
func (noopWarehouseMessageQueueRepository) DeleteInBatch(docList []models.WarehouseDoc) error {
	return nil
}

func TestUpdateWarehouse(t *testing.T) {
	holdingCode := "testHoldingCode"
	authUsername := "testUser"

	t.Run("successfully update warehouse master fields", func(t *testing.T) {
		mockRepo := new(MockWarehouseRepository)
		svc := services.NewWarehouseHttpService(mockRepo, noopWarehouseMessageQueueRepository{}, nil, nil)

		existing := models.WarehouseDoc{}
		existing.ID = primitive.NewObjectID()
		existing.GuidFixed = "123"
		existing.Code = "WH01"

		mockRepo.On("FindByGuid", mock.Anything, holdingCode, "123").Return(existing, nil)
		mockRepo.On("Update", mock.Anything, holdingCode, "123", mock.Anything).Return(nil)

		doc := models.Warehouse{
			Code: "WH01",
			Names: &[]common.NameX{
				*common.NewNameXWithCodeName("en", "warehouse 1"),
			},
		}

		err := svc.UpdateWarehouse(holdingCode, "123", authUsername, doc)
		assert.NoError(t, err)
	})

	t.Run("error when warehouse not found", func(t *testing.T) {
		mockRepo := new(MockWarehouseRepository)
		svc := services.NewWarehouseHttpService(mockRepo, noopWarehouseMessageQueueRepository{}, nil, nil)

		mockRepo.On("FindByGuid", mock.Anything, holdingCode, "999").Return(models.WarehouseDoc{}, nil)

		doc := models.Warehouse{Code: "WH99"}

		err := svc.UpdateWarehouse(holdingCode, "999", authUsername, doc)
		assert.Error(t, err)
	})
}
