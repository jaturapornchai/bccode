package services

import (
	"context"
	"testing"
	"time"

	"smlcloudplatform/internal/productsection/sectionbranch/models"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Create a mock for the SectionBranchRepository
type MockSectionBranchRepository struct {
	mock.Mock
}

func (m *MockSectionBranchRepository) Count(ctx context.Context, shopID string) (int, error) {
	args := m.Called(ctx, shopID)
	return args.Int(0), args.Error(1)
}

func (m *MockSectionBranchRepository) Create(ctx context.Context, doc models.SectionBranchDoc) (string, error) {
	args := m.Called(ctx, doc)
	return args.String(0), args.Error(1)
}

func (m *MockSectionBranchRepository) CreateInBatch(ctx context.Context, docList []models.SectionBranchDoc) error {
	args := m.Called(ctx, docList)
	return args.Error(0)
}

func (m *MockSectionBranchRepository) Update(ctx context.Context, shopID string, guid string, doc models.SectionBranchDoc) error {
	args := m.Called(ctx, shopID, guid, doc)
	return args.Error(0)
}

func (m *MockSectionBranchRepository) DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error {
	args := m.Called(ctx, shopID, guid, username)
	return args.Error(0)
}

func (m *MockSectionBranchRepository) Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error {
	args := m.Called(ctx, shopID, username, filters)
	return args.Error(0)
}

func (m *MockSectionBranchRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.SectionBranchInfo, mongopagination.PaginationData, error) {
	args := m.Called(ctx, shopID, searchInFields, pageable)
	return args.Get(0).([]models.SectionBranchInfo), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *MockSectionBranchRepository) FindByGuid(ctx context.Context, shopID string, guid string) (models.SectionBranchDoc, error) {
	args := m.Called(ctx, shopID, guid)
	return args.Get(0).(models.SectionBranchDoc), args.Error(1)
}

func (m *MockSectionBranchRepository) FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.SectionBranchItemGuid, error) {
	args := m.Called(ctx, shopID, columnName, itemGuidList)
	return args.Get(0).([]models.SectionBranchItemGuid), args.Error(1)
}

func (m *MockSectionBranchRepository) FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.SectionBranchDoc, error) {
	args := m.Called(ctx, shopID, indentityField, indentityValue)
	return args.Get(0).(models.SectionBranchDoc), args.Error(1)
}

func (m *MockSectionBranchRepository) FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.SectionBranchInfo, mongopagination.PaginationData, error) {
	args := m.Called(ctx, shopID, filters, searchInFields, pageable)
	return args.Get(0).([]models.SectionBranchInfo), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *MockSectionBranchRepository) FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.SectionBranchInfo, int, error) {
	args := m.Called(ctx, shopID, filters, searchInFields, projects, pageableLimit)
	return args.Get(0).([]models.SectionBranchInfo), args.Int(1), args.Error(2)
}

func (m *MockSectionBranchRepository) FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SectionBranchDeleteActivity, mongopagination.PaginationData, error) {
	args := m.Called(ctx, shopID, lastUpdatedDate, filters, pageable)
	return args.Get(0).([]models.SectionBranchDeleteActivity), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *MockSectionBranchRepository) FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SectionBranchActivity, mongopagination.PaginationData, error) {
	args := m.Called(ctx, shopID, lastUpdatedDate, filters, pageable)
	return args.Get(0).([]models.SectionBranchActivity), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *MockSectionBranchRepository) FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SectionBranchDeleteActivity, error) {
	args := m.Called(ctx, shopID, lastUpdatedDate, filters, pageableStep)
	return args.Get(0).([]models.SectionBranchDeleteActivity), args.Error(1)
}

func (m *MockSectionBranchRepository) FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SectionBranchActivity, error) {
	args := m.Called(ctx, shopID, lastUpdatedDate, filters, pageableStep)
	return args.Get(0).([]models.SectionBranchActivity), args.Error(1)
}

func mockNewGUID() string {
	return "testGuidFixed"
}

func TestSaveSectionBranch(t *testing.T) {
	mockRepo := new(MockSectionBranchRepository)
	// ... setup other mocks ...

	// Create an instance of SectionBranchHttpService
	svc := NewSectionBranchHttpService(mockRepo, mockNewGUID, nil)

	t.Run("Test SaveSectionBranch - Create", func(t *testing.T) {
		// Setup
		shopID := "testShopID"
		authUsername := "testUser"
		branchCode := "testBranchCode"

		doc := models.SectionBranch{
			BranchCode: branchCode,
		}

		emptyDoc := models.SectionBranchDoc{}
		mockRepo.On("FindByDocIndentityGuid", shopID, "branchcode", branchCode).Return(emptyDoc, nil)
		mockRepo.On("Create", mock.Anything).Return("testGuidFixed", nil)

		// Execute
		guidFixed, err := svc.SaveSectionBranch(shopID, authUsername, doc)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, guidFixed)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Test SaveSectionBranch - Update", func(t *testing.T) {
		// Setup
		shopID := "testShopID"
		authUsername := "testUser"
		branchCode := "testBranchCode"

		doc := models.SectionBranch{
			BranchCode: branchCode,
		}

		existingDoc := models.SectionBranchDoc{}

		existingDoc.ShopID = shopID
		existingDoc.GuidFixed = "testGuidFixed"
		existingDoc.SectionBranch = doc

		mockRepo.On("FindByDocIndentityGuid", shopID, "branchcode", branchCode).Return(existingDoc, nil)
		mockRepo.On("Update", shopID, existingDoc.GuidFixed, mock.Anything).Return(nil)

		// Execute
		guidFixed, err := svc.SaveSectionBranch(shopID, authUsername, doc)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, existingDoc.GuidFixed, guidFixed)
		// mockRepo.AssertExpectations(t)
	})
}
