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

type IWarehouseBinHttpService interface {
	CreateBin(holdingCode, authUsername string, doc models.WarehouseBin) (string, error)
	UpdateBin(holdingCode, guid, authUsername string, doc models.WarehouseBin) error
	DeleteBin(holdingCode, guid, authUsername string) error
	InfoBin(holdingCode, guid string) (models.WarehouseBinInfo, error)
	SearchBin(holdingCode, locationGuid string, pageable micromodels.Pageable) ([]models.WarehouseBinInfo, mongopagination.PaginationData, error)
	GetModuleName() string
}

type WarehouseBinHttpService struct {
	repo           repositories.IWarehouseBinRepository
	locationRepo   repositories.IWarehouseLocationRepository
	syncCacheRepo  mastersync.IMasterSyncCacheRepository
	contextTimeout time.Duration
}

func NewWarehouseBinHttpService(repo repositories.IWarehouseBinRepository, locationRepo repositories.IWarehouseLocationRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *WarehouseBinHttpService {
	return &WarehouseBinHttpService{
		repo:           repo,
		locationRepo:   locationRepo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: time.Duration(15) * time.Second,
	}
}

func (svc WarehouseBinHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc WarehouseBinHttpService) CreateBin(holdingCode, authUsername string, doc models.WarehouseBin) (string, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	locationDoc, err := svc.locationRepo.FindByGuid(ctx, holdingCode, doc.LocationGuid)
	if err != nil {
		return "", err
	}
	if locationDoc.ID == primitive.NilObjectID {
		return "", errors.New("location not found")
	}
	if locationDoc.WarehouseGuid != doc.WarehouseGuid {
		return "", errors.New("location does not belong to the given warehouse")
	}

	existing, _ := svc.repo.FindByLocationAndCode(ctx, holdingCode, doc.LocationGuid, doc.Code)
	if existing.GuidFixed != "" {
		return "", errors.New("bin code is exists")
	}

	if doc.Barcode != "" {
		existingBarcode, _ := svc.repo.FindByBarcode(ctx, holdingCode, doc.Barcode)
		if existingBarcode.GuidFixed != "" {
			return "", errors.New("bin barcode is exists")
		}
	}

	newGuidFixed := utils.NewGUID()

	docData := models.WarehouseBinDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.WarehouseBin = doc
	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)
	if err != nil {
		return "", err
	}

	svc.saveMasterSync(holdingCode)
	return newGuidFixed, nil
}

func (svc WarehouseBinHttpService) UpdateBin(holdingCode, guid, authUsername string, doc models.WarehouseBin) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)
	if err != nil {
		return err
	}
	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	locationDoc, err := svc.locationRepo.FindByGuid(ctx, holdingCode, doc.LocationGuid)
	if err != nil {
		return err
	}
	if locationDoc.ID == primitive.NilObjectID {
		return errors.New("location not found")
	}
	if locationDoc.WarehouseGuid != doc.WarehouseGuid {
		return errors.New("location does not belong to the given warehouse")
	}

	if doc.Code != findDoc.Code || doc.LocationGuid != findDoc.LocationGuid {
		existing, _ := svc.repo.FindByLocationAndCode(ctx, holdingCode, doc.LocationGuid, doc.Code)
		if existing.GuidFixed != "" && existing.GuidFixed != guid {
			return errors.New("bin code is exists")
		}
	}

	if doc.Barcode != "" && doc.Barcode != findDoc.Barcode {
		existingBarcode, _ := svc.repo.FindByBarcode(ctx, holdingCode, doc.Barcode)
		if existingBarcode.GuidFixed != "" && existingBarcode.GuidFixed != guid {
			return errors.New("bin barcode is exists")
		}
	}

	dataDoc := findDoc
	dataDoc.WarehouseBin = doc
	dataDoc.UpdatedBy = authUsername
	dataDoc.UpdatedAt = time.Now()

	if err := svc.repo.Update(ctx, holdingCode, guid, dataDoc); err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)
	return nil
}

func (svc WarehouseBinHttpService) DeleteBin(holdingCode, guid, authUsername string) error {
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

func (svc WarehouseBinHttpService) InfoBin(holdingCode, guid string) (models.WarehouseBinInfo, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)
	if err != nil {
		return models.WarehouseBinInfo{}, err
	}
	if findDoc.ID == primitive.NilObjectID {
		return models.WarehouseBinInfo{}, errors.New("document not found")
	}

	return findDoc.WarehouseBinInfo, nil
}

func (svc WarehouseBinHttpService) SearchBin(holdingCode, locationGuid string, pageable micromodels.Pageable) ([]models.WarehouseBinInfo, mongopagination.PaginationData, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{"code", "name"}
	return svc.repo.FindPageByLocation(ctx, holdingCode, locationGuid, searchInFields, pageable)
}

func (svc WarehouseBinHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		if err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName()); err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc WarehouseBinHttpService) GetModuleName() string {
	return "warehousebin"
}
