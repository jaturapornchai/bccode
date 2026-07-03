package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/warehouse/models"
	"smlcloudplatform/internal/warehouse/repositories"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IWarehouseLocationHttpService interface {
	CreateLocation(holdingCode, authUsername string, doc models.WarehouseLocation) (string, error)
	UpdateLocation(holdingCode, guid, authUsername string, doc models.WarehouseLocation) error
	DeleteLocation(holdingCode, guid, authUsername string) error
	InfoLocation(holdingCode, guid string) (models.WarehouseLocationInfo, error)
	SearchLocation(holdingCode, warehouseGuid string, pageable micromodels.Pageable) ([]models.WarehouseLocationInfo, mongopagination.PaginationData, error)
	GetModuleName() string
}

type WarehouseLocationHttpService struct {
	repo           repositories.IWarehouseLocationRepository
	warehouseRepo  repositories.IWarehouseRepository
	syncCacheRepo  mastersync.IMasterSyncCacheRepository
	contextTimeout time.Duration
}

func NewWarehouseLocationHttpService(repo repositories.IWarehouseLocationRepository, warehouseRepo repositories.IWarehouseRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *WarehouseLocationHttpService {
	return &WarehouseLocationHttpService{
		repo:           repo,
		warehouseRepo:  warehouseRepo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: time.Duration(15) * time.Second,
	}
}

func (svc WarehouseLocationHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

// validateLocationCompanyScope enforces that a Location's CompanyGuids is a subset of the parent
// warehouse-level CompanyGuids. An empty warehouseCompanyGuids means "all companies" — any location
// scope is a valid subset. An empty location CompanyGuids means "inherit all the warehouse allows".
func validateLocationCompanyScope(locationCompanyGuids []string, warehouseCompanyGuids []string) error {
	if len(warehouseCompanyGuids) == 0 {
		return nil
	}
	allowed := make(map[string]bool, len(warehouseCompanyGuids))
	for _, g := range warehouseCompanyGuids {
		allowed[g] = true
	}
	for _, g := range locationCompanyGuids {
		if !allowed[g] {
			return fmt.Errorf("ที่เก็บสินค้าใช้ได้เฉพาะบริษัทที่คลังอนุญาตเท่านั้น (บริษัท %s ไม่อยู่ในสิทธิ์ของคลัง)", g)
		}
	}
	return nil
}

func (svc WarehouseLocationHttpService) CreateLocation(holdingCode, authUsername string, doc models.WarehouseLocation) (string, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	warehouseDoc, err := svc.warehouseRepo.FindByGuid(ctx, holdingCode, doc.WarehouseGuid)
	if err != nil {
		return "", err
	}
	if warehouseDoc.ID == primitive.NilObjectID {
		return "", errors.New("warehouse not found")
	}

	if err := validateLocationCompanyScope(doc.CompanyGuids, warehouseDoc.CompanyGuids); err != nil {
		return "", err
	}

	// FindByWarehouseAndCode returns a mongo "no documents" error plus a zero-value doc when the
	// code is free — that's the expected case, so only the doc's GuidFixed decides duplicate-ness.
	existing, _ := svc.repo.FindByWarehouseAndCode(ctx, holdingCode, doc.WarehouseGuid, doc.Code)
	if existing.GuidFixed != "" {
		return "", errors.New("location code is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.WarehouseLocationDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.WarehouseLocation = doc
	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)
	if err != nil {
		return "", err
	}

	svc.saveMasterSync(holdingCode)
	return newGuidFixed, nil
}

func (svc WarehouseLocationHttpService) UpdateLocation(holdingCode, guid, authUsername string, doc models.WarehouseLocation) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)
	if err != nil {
		return err
	}
	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	warehouseDoc, err := svc.warehouseRepo.FindByGuid(ctx, holdingCode, doc.WarehouseGuid)
	if err != nil {
		return err
	}
	if warehouseDoc.ID == primitive.NilObjectID {
		return errors.New("warehouse not found")
	}

	if err := validateLocationCompanyScope(doc.CompanyGuids, warehouseDoc.CompanyGuids); err != nil {
		return err
	}

	if doc.Code != findDoc.Code || doc.WarehouseGuid != findDoc.WarehouseGuid {
		existing, _ := svc.repo.FindByWarehouseAndCode(ctx, holdingCode, doc.WarehouseGuid, doc.Code)
		if existing.GuidFixed != "" && existing.GuidFixed != guid {
			return errors.New("location code is exists")
		}
	}

	dataDoc := findDoc
	dataDoc.WarehouseLocation = doc
	dataDoc.UpdatedBy = authUsername
	dataDoc.UpdatedAt = time.Now()

	if err := svc.repo.Update(ctx, holdingCode, guid, dataDoc); err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)
	return nil
}

func (svc WarehouseLocationHttpService) DeleteLocation(holdingCode, guid, authUsername string) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)
	if err != nil {
		return err
	}
	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	if err := svc.repo.DeleteByGuidfixed(ctx, holdingCode, guid, authUsername); err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)
	return nil
}

func (svc WarehouseLocationHttpService) InfoLocation(holdingCode, guid string) (models.WarehouseLocationInfo, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)
	if err != nil {
		return models.WarehouseLocationInfo{}, err
	}
	if findDoc.ID == primitive.NilObjectID {
		return models.WarehouseLocationInfo{}, errors.New("document not found")
	}

	return findDoc.WarehouseLocationInfo, nil
}

func (svc WarehouseLocationHttpService) SearchLocation(holdingCode, warehouseGuid string, pageable micromodels.Pageable) ([]models.WarehouseLocationInfo, mongopagination.PaginationData, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{"code"}
	return svc.repo.FindPageByWarehouse(ctx, holdingCode, warehouseGuid, searchInFields, pageable)
}

func (svc WarehouseLocationHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		if err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName()); err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc WarehouseLocationHttpService) GetModuleName() string {
	return "warehouselocation"
}
