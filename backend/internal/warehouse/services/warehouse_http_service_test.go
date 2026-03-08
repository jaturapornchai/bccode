package services_test

import (
	"context"
	"errors"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/warehouse/models"
	"smlcloudplatform/internal/warehouse/services"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"testing"
	"time"

	"github.com/smlsoft/mongopagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWarehouseRepository mocks the WarehouseRepository
type MockWarehouseRepository struct {
	mock.Mock
}

func (m *MockWarehouseRepository) Count(ctx context.Context, shopID string) (int, error) {
	args := m.Called(ctx, shopID)
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

func (m *MockWarehouseRepository) Update(ctx context.Context, shopID, guid string, doc models.WarehouseDoc) error {
	args := m.Called(ctx, shopID, guid, doc)
	return args.Error(0)
}

func (m *MockWarehouseRepository) DeleteByGuidfixed(ctx context.Context, shopID, guid, username string) error {
	args := m.Called(ctx, shopID, guid, username)
	return args.Error(0)
}

func (m *MockWarehouseRepository) Delete(ctx context.Context, shopID, username string, filters map[string]interface{}) error {
	args := m.Called(ctx, shopID, username, filters)
	return args.Error(0)
}

func (m *MockWarehouseRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.WarehouseInfo, mongopagination.PaginationData, error) {
	args := m.Called(ctx, shopID, searchInFields, pageable)
	return args.Get(0).([]models.WarehouseInfo), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *MockWarehouseRepository) FindByGuid(ctx context.Context, shopID, guid string) (models.WarehouseDoc, error) {
	args := m.Called(ctx, shopID, guid)
	return args.Get(0).(models.WarehouseDoc), args.Error(1)
}

func (m *MockWarehouseRepository) FindInItemGuid(ctx context.Context, shopID, columnName string, itemGuidList []string) ([]models.WarehouseItemGuid, error) {
	args := m.Called(ctx, shopID, columnName, itemGuidList)
	return args.Get(0).([]models.WarehouseItemGuid), args.Error(1)
}

func (m *MockWarehouseRepository) FindByDocIndentityGuid(ctx context.Context, shopID, indentityField string, indentityValue interface{}) (models.WarehouseDoc, error) {
	args := m.Called(ctx, shopID, indentityField, indentityValue)
	return args.Get(0).(models.WarehouseDoc), args.Error(1)
}

func (m *MockWarehouseRepository) FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.WarehouseInfo, mongopagination.PaginationData, error) {
	args := m.Called(ctx, shopID, filters, searchInFields, pageable)
	return args.Get(0).([]models.WarehouseInfo), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *MockWarehouseRepository) FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.WarehouseInfo, int, error) {
	args := m.Called(ctx, shopID, filters, searchInFields, projects, pageableLimit)
	return args.Get(0).([]models.WarehouseInfo), args.Int(1), args.Error(2)
}

func (m *MockWarehouseRepository) FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.WarehouseDeleteActivity, mongopagination.PaginationData, error) {
	args := m.Called(ctx, shopID, lastUpdatedDate, filters, pageable)
	return args.Get(0).([]models.WarehouseDeleteActivity), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *MockWarehouseRepository) FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.WarehouseActivity, mongopagination.PaginationData, error) {
	args := m.Called(ctx, shopID, lastUpdatedDate, filters, pageable)
	return args.Get(0).([]models.WarehouseActivity), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *MockWarehouseRepository) FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.WarehouseDeleteActivity, error) {
	args := m.Called(ctx, shopID, lastUpdatedDate, filters, pageableStep)
	return args.Get(0).([]models.WarehouseDeleteActivity), args.Error(1)
}

func (m *MockWarehouseRepository) FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.WarehouseActivity, error) {
	args := m.Called(ctx, shopID, lastUpdatedDate, filters, pageableStep)
	return args.Get(0).([]models.WarehouseActivity), args.Error(1)
}

func (m *MockWarehouseRepository) FindLocationPage(ctx context.Context, shopID string, pageable micromodels.Pageable) ([]models.LocationInfo, mongopagination.PaginationData, error) {
	args := m.Called(ctx, shopID, pageable)
	return args.Get(0).([]models.LocationInfo), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *MockWarehouseRepository) FindShelfPage(ctx context.Context, shopID string, pageable micromodels.Pageable) ([]models.ShelfInfo, mongopagination.PaginationData, error) {
	args := m.Called(ctx, shopID, pageable)
	return args.Get(0).([]models.ShelfInfo), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *MockWarehouseRepository) FindWarehouseByLocation(ctx context.Context, shopID, warehouseCode, locationCode string) (models.WarehouseDoc, error) {
	args := m.Called(ctx, shopID, warehouseCode, locationCode)
	return args.Get(0).(models.WarehouseDoc), args.Error(1)
}

func (m *MockWarehouseRepository) FindWarehouseByShelf(ctx context.Context, shopID, warehouseCode, locationCode, shelfCode string) (models.WarehouseDoc, error) {
	args := m.Called(ctx, shopID, warehouseCode, locationCode, shelfCode)
	return args.Get(0).(models.WarehouseDoc), args.Error(1)
}

func (m *MockWarehouseRepository) Transaction(ctx context.Context, queryFunc func(ctx context.Context) error) error {
	args := m.Called(ctx, queryFunc)
	return args.Error(0)
}

func TestUpdateLocation(t *testing.T) {
	shopID := "testShopID"
	authUsername := "testUser"
	warehouseCode := "WH01"
	locationCode := "L01"

	warehouseDoc := models.WarehouseDoc{}

	warehouseDoc.GuidFixed = "123"
	warehouseDoc.Location = &[]models.Location{
		{
			Code: "L01",
			Names: &[]common.NameX{
				*common.NewNameXWithCodeName("en", "location 1"),
			},
		},
	}

	t.Run("successfully update location within the same warehouse", func(t *testing.T) {
		mockRepo := new(MockWarehouseRepository)
		svc := services.NewWarehouseHttpService(mockRepo, nil, nil)

		mockRepo.On("FindWarehouseByLocation", shopID, warehouseCode, locationCode).Return(warehouseDoc, nil)
		mockRepo.On("Update", shopID, warehouseDoc.GuidFixed, mock.Anything).Return(nil)
		mockRepo.On("Transaction", mock.Anything).Return(nil)

		doc := models.LocationRequest{
			Code: "L01",
			Names: &[]common.NameX{
				*common.NewNameXWithCodeName("en", "location 1"),
			},
		}

		err := svc.UpdateLocation(shopID, authUsername, warehouseCode, locationCode, doc)
		assert.NoError(t, err)
	})

	t.Run("successfully move location to a different warehouse", func(t *testing.T) {
		mockRepo := new(MockWarehouseRepository)
		svc := services.NewWarehouseHttpService(mockRepo, nil, nil)

		targetWarehouseDoc := models.WarehouseDoc{}

		targetWarehouseDoc.GuidFixed = "456"
		targetWarehouseDoc.Code = "WH01"
		targetWarehouseDoc.Location = &[]models.Location{}

		mockRepo.On("FindWarehouseByLocation", shopID, warehouseCode, locationCode).Return(warehouseDoc, nil)
		mockRepo.On("FindByDocIndentityGuid", shopID, "code", "WH01").Return(targetWarehouseDoc, nil)
		mockRepo.On("Update", shopID, warehouseDoc.GuidFixed, mock.Anything).Return(nil)
		mockRepo.On("Update", shopID, targetWarehouseDoc.GuidFixed, mock.Anything).Return(nil)
		mockRepo.On("Transaction", mock.Anything).Return(nil)

		doc := models.LocationRequest{}

		doc.WarehouseCode = "WH01"
		doc.Code = "L01"
		doc.Names = &[]common.NameX{
			*common.NewNameXWithCodeName("lo1", "location 1"),
		}

		err := svc.UpdateLocation(shopID, authUsername, warehouseCode, locationCode, doc)
		assert.NoError(t, err)
	})

	t.Run("error when location not found", func(t *testing.T) {
		mockRepo := new(MockWarehouseRepository)
		svc := services.NewWarehouseHttpService(mockRepo, nil, nil)

		mockRepo.On("FindWarehouseByLocation", shopID, warehouseCode, locationCode).Return(models.WarehouseDoc{}, nil)

		doc := models.LocationRequest{
			WarehouseCode: "WH02",
			Code:          "L01",
			Names: &[]common.NameX{
				*common.NewNameXWithCodeName("en", "location 1"),
			},
		}

		err := svc.UpdateLocation(shopID, authUsername, warehouseCode, locationCode, doc)
		assert.Error(t, err)
		assert.Equal(t, errors.New("document not found"), err)
	})
	/*
		t.Run("error when target warehouse not found", func(t *testing.T) {
			mockRepo := new(MockWarehouseRepository)
			svc := WarehouseHttpService{repo: mockRepo}

			mockRepo.On("FindWarehouseByLocation", shopID, warehouseCode, locationCode).Return(warehouseDoc, nil)
			mockRepo.On("FindByDocIndentityGuid", shopID, "code", "WH02").Return(models.WarehouseDoc{}, nil)

			doc := models.LocationRequest{
				WarehouseCode: "WH02",
				Code:          "L01",
				Names:         "Updated Location",
			}

			err := svc.UpdateLocation(shopID, authUsername, warehouseCode, locationCode, doc)
			assert.Error(t, err)
			assert.Equal(t, errors.New("document not found"), err)
		})

		t.Run("error when target location code already exists", func(t *testing.T) {
			mockRepo := new(MockWarehouseRepository)
			svc := WarehouseHttpService{repo: mockRepo}

			targetWarehouseDoc := models.WarehouseDoc{
				GuidFixed: "456",
				Location: &[]models.Location{
					{
						Code:  "L01",
						Names: "Existing Location",
					},
				},
			}

			mockRepo.On("FindWarehouseByLocation", shopID, warehouseCode, locationCode).Return(warehouseDoc, nil)
			mockRepo.On("FindByDocIndentityGuid", shopID, "code", "WH02").Return(targetWarehouseDoc, nil)

			doc := models.LocationRequest{
				WarehouseCode: "WH02",
				Code:          "L01",
				Names:         "Updated Location",
			}

			err := svc.UpdateLocation(shopID, authUsername, warehouseCode, locationCode, doc)
			assert.Error(t, err)
			assert.Equal(t, errors.New("location code is exists"), err)
		})
	*/
}
