package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/debtaccount/debtor/models"
	"smlcloudplatform/internal/debtaccount/debtor/repositories"
	groupModels "smlcloudplatform/internal/debtaccount/debtorgroup/models"
	groupRepositories "smlcloudplatform/internal/debtaccount/debtorgroup/repositories"
	"smlcloudplatform/internal/logger"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/samber/lo"
	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IDebtorHttpService interface {
	CreateDebtor(holdingCode string, authUsername string, doc models.DebtorRequest) (string, error)
	UpdateDebtor(holdingCode string, guid string, authUsername string, doc models.DebtorRequest) error
	DeleteDebtor(holdingCode string, guid string, authUsername string) error
	DeleteDebtorByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoDebtor(holdingCode string, guid string) (models.DebtorInfo, error)
	InfoDebtorByCode(holdingCode string, code string) (models.DebtorInfo, error)
	InfoDebtorByLine(holdingCode string, code string) (models.DebtorInfo, error)
	SearchDebtor(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DebtorInfo, mongopagination.PaginationData, error)
	SearchDebtorStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DebtorInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.DebtorRequest) (common.BulkImport, error)
	InfoAuthDebtor(holdingCode string, username string, password string) (models.DebtorInfo, error)
	SearchPointTransactions(holdingCode string, debtorCode string, pageableStep micromodels.PageableStep) ([]models.PointTransactionInfo, int, error)
	RecalPointByPointsCode(holdingCode string, pointsCode string, authUsername string) error
	AddPointManually(holdingCode string, pointsCode string, pointAmount float64, description string, authUsername string) error
	BulkAddPoints(holdingCode string, pointsList []models.OpeningBalancePointRequest, authUsername string) (models.BulkImportPointResult, error)
	DeleteManualPointTransaction(holdingCode string, pointsCode string, docNo string, authUsername string) error

	GetModuleName() string
}

type DebtorHttpService struct {
	repo           repositories.IDebtorRepository
	repoMq         repositories.IDebtorMessageQueueRepository
	repoGroup      groupRepositories.IDebtorGroupRepository
	pointTransRepo repositories.IPointTransactionRepository
	syncCacheRepo  mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.DebtorActivity, models.DebtorDeleteActivity]
	hashPassword      func(password string) (string, error)
	checkHashPassword func(password, hash string) bool
	contextTimeout    time.Duration
}

func NewDebtorHttpService(
	repo repositories.IDebtorRepository,
	repoMq repositories.IDebtorMessageQueueRepository,
	repoGroup groupRepositories.IDebtorGroupRepository,
	pointTransRepo repositories.IPointTransactionRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	hashPassword func(password string) (string, error),
	checkHashPassword func(password, hash string) bool,
) *DebtorHttpService {
	contextTimeout := time.Duration(15) * time.Second

	insSvc := &DebtorHttpService{
		repo:              repo,
		repoMq:            repoMq,
		repoGroup:         repoGroup,
		pointTransRepo:    pointTransRepo,
		syncCacheRepo:     syncCacheRepo,
		hashPassword:      hashPassword,
		checkHashPassword: checkHashPassword,
		contextTimeout:    contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.DebtorActivity, models.DebtorDeleteActivity](repo)

	return insSvc
}

func (svc DebtorHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc DebtorHttpService) InfoAuthDebtor(holdingCode string, username string, password string) (models.DebtorInfo, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	if username == "" || password == "" {
		return models.DebtorInfo{}, errors.New("username or password incorrect")
	}

	findDoc, err := svc.repo.FindAuthByUsername(ctx, holdingCode, username)

	if err != nil {
		return models.DebtorInfo{}, err
	}

	if findDoc.Auth.Username == "" {
		return models.DebtorInfo{}, errors.New("username or password incorrect")
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.DebtorInfo{}, errors.New("username or password incorrect")
	}

	if findDoc.Auth.Password == "" {
		return models.DebtorInfo{}, errors.New("username or password incorrect")
	}

	invalidPassword := !svc.checkHashPassword(password, findDoc.Auth.Password)

	if invalidPassword {
		return models.DebtorInfo{}, errors.New("username or password incorrect")
	}

	findDoc.Auth.Password = ""

	return findDoc.DebtorInfo, nil

}

func (svc DebtorHttpService) CreateDebtor(holdingCode string, authUsername string, doc models.DebtorRequest) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if findDoc.Code != "" {
		return "", errors.New("code is exists")
	}

	if doc.Auth.Username != "" {
		findDocAuth, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "auth.username", doc.Code)

		if err != nil {
			return "", err
		}

		if findDocAuth.Auth.Username != "" {
			return "", errors.New("auth username is exists")
		}
	}

	// Check PointsCode uniqueness if provided
	if doc.PointsCode != "" {
		findDocPointsCode, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "points_code", doc.PointsCode)

		if err != nil {
			return "", err
		}

		if findDocPointsCode.PointsCode != "" {
			return "", errors.New("pointscode is exists")
		}
	}

	newGuidFixed := utils.NewGUID()

	docData := models.DebtorDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.Debtor = doc.Debtor
	docData.GroupGUIDs = &doc.Groups

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	if doc.Auth.Password != "" {
		hashedPassword, err := svc.hashPassword(doc.Auth.Password)
		if err != nil {
			return "", err
		}
		docData.Auth.Password = hashedPassword
	}

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	// Recalculate point balance if pointscode is provided
	if doc.PointsCode != "" {
		go func() {
			err := svc.RecalPointByPointsCode(holdingCode, doc.PointsCode, authUsername)
			if err != nil {
				logger.GetLogger().Errorf("Recalculate point balance error for pointscode %s :: %s", doc.PointsCode, err.Error())
			}
		}()
	}

	go func() {
		svc.saveMasterSync(holdingCode)
		err = svc.repoMq.Create(docData)
		if err != nil {
			logger.GetLogger().Errorf("Create creditor message queue error :: %s", err.Error())
		}
	}()

	return newGuidFixed, nil
}

func (svc DebtorHttpService) UpdateDebtor(holdingCode string, guid string, authUsername string, doc models.DebtorRequest) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	if doc.Auth.Username != "" {
		findDocAuth, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "auth.username", doc.Auth.Username)

		if err != nil {
			return err
		}

		if findDoc.Auth.Username != findDocAuth.Auth.Username && findDocAuth.Auth.Username != "" {
			return errors.New("auth username is exists")
		}
	}

	// Check PointsCode uniqueness if provided
	if doc.PointsCode != "" {
		findDocPointsCode, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "points_code", doc.PointsCode)

		if err != nil {
			return err
		}

		// Only check for duplicates if the pointscode is different from current record
		if findDoc.PointsCode != findDocPointsCode.PointsCode && findDocPointsCode.PointsCode != "" {
			return errors.New("pointscode is exists")
		}
	}

	dataDoc := findDoc

	dataDoc.Debtor = doc.Debtor
	dataDoc.GroupGUIDs = &doc.Groups

	dataDoc.UpdatedBy = authUsername
	dataDoc.UpdatedAt = time.Now()

	if doc.Auth.Password != "" {
		hashedPassword, err := svc.hashPassword(doc.Auth.Password)
		if err != nil {
			return err
		}
		dataDoc.Auth.Password = hashedPassword
	}

	err = svc.repo.Update(ctx, holdingCode, guid, dataDoc)

	if err != nil {
		return err
	}

	// Recalculate point balance if pointscode is provided or changed
	if doc.PointsCode != "" && doc.PointsCode != findDoc.PointsCode {
		go func() {
			err := svc.RecalPointByPointsCode(holdingCode, doc.PointsCode, authUsername)
			if err != nil {
				logger.GetLogger().Errorf("Recalculate point balance error for pointscode %s :: %s", doc.PointsCode, err.Error())
			}
		}()
	}

	go func() {
		svc.saveMasterSync(holdingCode)
		err = svc.repoMq.Update(dataDoc)
		if err != nil {
			logger.GetLogger().Errorf("Update creditor message queue error :: %s", err.Error())
		}
	}()

	return nil
}

func (svc DebtorHttpService) DeleteDebtor(holdingCode string, guid string, authUsername string) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, holdingCode, guid, authUsername)
	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(holdingCode)
		err = svc.repoMq.Delete(findDoc)
		if err != nil {
			logger.GetLogger().Errorf("Delete creditor message queue error :: %s", err.Error())
		}
	}()

	return nil
}

func (svc DebtorHttpService) DeleteDebtorByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDocs, err := svc.repo.FindByGuids(ctx, holdingCode, GUIDs)

	if err != nil {
		return err
	}

	deleteFilterQuery := map[string]interface{}{
		"guid_fixed": bson.M{"$in": GUIDs},
	}

	err = svc.repo.Delete(ctx, holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(holdingCode)
		err = svc.repoMq.DeleteInBatch(findDocs)
		if err != nil {
			logger.GetLogger().Errorf("Delete creditor message queue error :: %s", err.Error())
		}
	}()

	return nil
}

func (svc DebtorHttpService) InfoDebtor(holdingCode string, guid string) (models.DebtorInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.DebtorInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.DebtorInfo{}, errors.New("document not found")
	}

	if findDoc.GroupGUIDs != nil {
		findGroups, err := svc.repoGroup.FindByGuids(ctx, holdingCode, *findDoc.GroupGUIDs)

		if err != nil {
			return models.DebtorInfo{}, err
		}

		custGroupInfo := lo.Map[groupModels.DebtorGroupDoc, groupModels.DebtorGroupInfo](
			findGroups,
			func(docGroup groupModels.DebtorGroupDoc, idx int) groupModels.DebtorGroupInfo {
				return docGroup.DebtorGroupInfo
			})

		findDoc.DebtorInfo.Groups = &custGroupInfo
	}

	docInfo := findDoc.DebtorInfo
	docInfo.Auth.Password = ""

	return docInfo, nil

}

func (svc DebtorHttpService) InfoDebtorByCode(holdingCode string, code string) (models.DebtorInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", code)

	if err != nil {
		return models.DebtorInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.DebtorInfo{}, errors.New("document not found")
	}

	if findDoc.GroupGUIDs != nil {
		findGroups, err := svc.repoGroup.FindByGuids(ctx, holdingCode, *findDoc.GroupGUIDs)

		if err != nil {
			return models.DebtorInfo{}, err
		}

		custGroupInfo := lo.Map[groupModels.DebtorGroupDoc, groupModels.DebtorGroupInfo](
			findGroups,
			func(docGroup groupModels.DebtorGroupDoc, idx int) groupModels.DebtorGroupInfo {
				return docGroup.DebtorGroupInfo
			})

		findDoc.DebtorInfo.Groups = &custGroupInfo
	}

	docInfo := findDoc.DebtorInfo
	docInfo.Auth.Password = ""

	return docInfo, nil

}

func (svc DebtorHttpService) InfoDebtorByLine(holdingCode string, code string) (models.DebtorInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "line.lineuid", code)

	if err != nil {
		return models.DebtorInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.DebtorInfo{}, errors.New("document not found")
	}

	findGroups, err := svc.repoGroup.FindByGuids(ctx, holdingCode, *findDoc.GroupGUIDs)

	if err != nil {
		return models.DebtorInfo{}, err
	}

	custGroupInfo := lo.Map[groupModels.DebtorGroupDoc, groupModels.DebtorGroupInfo](
		findGroups,
		func(docGroup groupModels.DebtorGroupDoc, idx int) groupModels.DebtorGroupInfo {
			return docGroup.DebtorGroupInfo
		})

	findDoc.DebtorInfo.Groups = &custGroupInfo

	docInfo := findDoc.DebtorInfo
	docInfo.Auth.Password = ""

	return docInfo, nil

}

func (svc DebtorHttpService) SearchDebtor(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DebtorInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
		"groups",
		"fund_code",
		"addressforbilling.address.0",
		"addressforbilling.phoneprimary",
		"addressforbilling.phonesecondary",
	}

	// Check if holding_code filter exists (from shopsid parameter)
	var docList []models.DebtorInfo
	var pagination mongopagination.PaginationData
	var err error

	if _, hasHoldingCodeFilter := filters["holding_code"]; hasHoldingCodeFilter {
		// Use FindPageFilterNoHoldingCode when holding_code filter exists
		docList, pagination, err = svc.repo.FindPageFilterNoHoldingCode(ctx, filters, searchInFields, pageable)
	} else {
		// Use regular FindPageFilter when no holding_code filter
		docList, pagination, err = svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)
	}

	if err != nil {
		return []models.DebtorInfo{}, pagination, err
	}

	for idx, doc := range docList {
		if doc.GroupGUIDs != nil {
			findCustGroups, err := svc.repoGroup.FindByGuids(ctx, holdingCode, *doc.GroupGUIDs)
			if err != nil {
				return []models.DebtorInfo{}, pagination, err
			}

			custGroupInfo := lo.Map[groupModels.DebtorGroupDoc, groupModels.DebtorGroupInfo](
				findCustGroups,
				func(docGroup groupModels.DebtorGroupDoc, idx int) groupModels.DebtorGroupInfo {
					return docGroup.DebtorGroupInfo
				})

			docList[idx].Groups = &custGroupInfo
		}
	}

	for i := 0; i < len(docList); i++ {
		docList[i].Auth.Password = ""
	}

	return docList, pagination, nil
}

func (svc DebtorHttpService) SearchDebtorStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DebtorInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
		"groups",
		"fund_code",
		"addressforbilling.address.0",
		"addressforbilling.phoneprimary",
		"addressforbilling.phonesecondary",
	}

	selectFields := map[string]interface{}{}

	// Check if holding_code filter exists (from shopsid parameter)
	var docList []models.DebtorInfo
	var total int
	var err error

	if _, hasHoldingCodeFilter := filters["holding_code"]; hasHoldingCodeFilter {
		// Use FindStepNoHoldingCode when holding_code filter exists
		docList, total, err = svc.repo.FindStepNoHoldingCode(ctx, filters, searchInFields, selectFields, pageableStep)
	} else {
		// Use regular FindStep when no holding_code filter
		docList, total, err = svc.repo.FindStep(ctx, holdingCode, filters, searchInFields, selectFields, pageableStep)
	}

	if err != nil {
		return []models.DebtorInfo{}, 0, err
	}

	for idx, doc := range docList {
		if doc.GroupGUIDs != nil {
			findCustGroups, err := svc.repoGroup.FindByGuids(ctx, holdingCode, *doc.GroupGUIDs)
			if err != nil {
				return []models.DebtorInfo{}, 0, err
			}

			custGroupInfo := lo.Map[groupModels.DebtorGroupDoc, groupModels.DebtorGroupInfo](
				findCustGroups,
				func(docGroup groupModels.DebtorGroupDoc, idx int) groupModels.DebtorGroupInfo {
					return docGroup.DebtorGroupInfo
				})

			docList[idx].Groups = &custGroupInfo
		}
	}

	for i := 0; i < len(docList); i++ {
		docList[i].Auth.Password = ""
	}

	return docList, total, nil
}

func (svc DebtorHttpService) SaveInBatch(holdingCode string, authUsername string, dataListReq []models.DebtorRequest) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	dataList := []models.Debtor{}
	for i := range dataListReq {
		// สร้าง copy ของ Groups
		groupsCopy := make([]string, len(dataListReq[i].Groups))
		copy(groupsCopy, dataListReq[i].Groups)

		dataListReq[i].Debtor.GroupGUIDs = &groupsCopy
		dataList = append(dataList, dataListReq[i].Debtor)
	}

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.Debtor](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.Code)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, holdingCode, "code", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.Code)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Debtor, models.DebtorDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.Debtor) models.DebtorDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.DebtorDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.Debtor = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Debtor, models.DebtorDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.DebtorDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		},
		func(doc models.DebtorDoc) bool {
			return doc.Code != ""
		},
		func(holdingCode string, authUsername string, data models.Debtor, doc models.DebtorDoc) error {

			doc.Debtor = data
			doc.UpdatedBy = authUsername
			doc.UpdatedAt = time.Now()

			err = svc.repo.Update(ctx, holdingCode, doc.GuidFixed, doc)
			if err != nil {
				return nil
			}
			return nil
		},
	)

	if len(createDataList) > 0 {
		err = svc.repo.CreateInBatch(ctx, createDataList)

		if err != nil {
			return common.BulkImport{}, err
		}

	}

	createDataKey := []string{}

	for _, doc := range createDataList {
		createDataKey = append(createDataKey, doc.Code)
	}

	payloadDuplicateDataKey := []string{}
	for _, doc := range payloadDuplicateList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.Code)
	}

	updateDataKey := []string{}
	for _, doc := range updateSuccessDataList {

		updateDataKey = append(updateDataKey, doc.Code)
	}

	updateFailDataKey := []string{}
	for _, doc := range updateFailDataList {
		updateFailDataKey = append(updateFailDataKey, svc.getDocIDKey(doc))
	}

	go func() {
		svc.saveMasterSync(holdingCode)
		err = svc.repoMq.CreateInBatch(createDataList)
		if err != nil {
			logger.GetLogger().Errorf("Create creditor message queue error :: %s", err.Error())
		}
		svc.repoMq.UpdateInBatch(updateSuccessDataList)

		if err != nil {
			logger.GetLogger().Errorf("Update creditor message queue error :: %s", err.Error())
		}
	}()

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc DebtorHttpService) getDocIDKey(doc models.Debtor) string {
	return doc.Code
}

func (svc DebtorHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc DebtorHttpService) GetModuleName() string {
	return "debtor"
}

func (svc DebtorHttpService) SearchPointTransactions(holdingCode string, debtorCode string, pageableStep micromodels.PageableStep) ([]models.PointTransactionInfo, int, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docList, total, err := svc.pointTransRepo.FindPointTransactionsByPointsCode(ctx, holdingCode, debtorCode, pageableStep)

	if err != nil {
		return []models.PointTransactionInfo{}, 0, err
	}

	return docList, total, nil
}

// RecalPointByPointsCode recalculates point balance for a pointscode based on all point transactions
func (svc DebtorHttpService) RecalPointByPointsCode(holdingCode string, pointsCode string, authUsername string) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	if pointsCode == "" {
		return errors.New("pointscode is required")
	}

	// Find the debtor by pointscode
	findDebtor, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "points_code", pointsCode)
	if err != nil {
		return err
	}

	if findDebtor.ID == primitive.NilObjectID {
		return errors.New("debtor not found with pointscode: " + pointsCode)
	}

	// Get all point transactions for this pointscode
	pageableStep := micromodels.PageableStep{
		Skip:  0,
		Limit: 10000, // Get all transactions
	}

	transactions, _, err := svc.pointTransRepo.FindPointTransactionsByPointsCode(ctx, holdingCode, pointsCode, pageableStep)
	if err != nil {
		return err
	}

	// Calculate balance from all transactions
	var calculatedBalance float64 = 0.0
	for _, transaction := range transactions {
		calculatedBalance += transaction.PointAmount
	}

	// Ensure balance is not negative
	if calculatedBalance < 0 {
		calculatedBalance = 0
	}

	// Update the debtor's point balance
	findDebtor.PointBalance = calculatedBalance
	findDebtor.UpdatedBy = authUsername
	findDebtor.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, findDebtor.GuidFixed, findDebtor)
	if err != nil {
		return err
	}

	return nil
}

// AddPointManually adds points manually to a customer (can be used for opening balance, adjustments, or promotions)
func (svc DebtorHttpService) AddPointManually(holdingCode string, pointsCode string, pointAmount float64, description string, authUsername string) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Validate inputs
	if pointsCode == "" {
		return fmt.Errorf("pointsCode is required")
	}
	if pointAmount <= 0 {
		return fmt.Errorf("pointAmount must be greater than 0")
	}

	// Find debtor by pointscode
	debtor, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "points_code", pointsCode)
	if err != nil {
		return fmt.Errorf("failed to find debtor: %w", err)
	}

	// Check if debtor exists
	if debtor.GuidFixed == "" {
		return fmt.Errorf("debtor not found with pointscode: %s", pointsCode)
	}

	currentBalance := debtor.PointBalance

	// Set default description if not provided
	if description == "" {
		description = "Manual point adjustment"
	}

	// Create point transaction for manual adjustment
	pointTransaction := models.PointTransactionDoc{
		PointTransactionData: models.PointTransactionData{
			HoldingCodeentity: common.HoldingCodeentity{HoldingCode: holdingCode},
			PointTransactionInfo: models.PointTransactionInfo{
				DocIdentity: common.DocIdentity{GuidFixed: utils.NewGUID()},
				PointTransaction: models.PointTransaction{
					TransactionDocNo: "ADJ-" + time.Now().Format("20060102150405"), // ADJ = Adjustment
					TransactionDate:  time.Now(),
					DebtorCode:       debtor.Code,
					PointsCode:       pointsCode,
					TransactionType:  3, // Manual Adjustment / Opening Balance
					PointAmount:      pointAmount,
					BalanceBefore:    currentBalance,
					BalanceAfter:     currentBalance + pointAmount,
					Description:      description,
				},
			},
		},
		ActivityDoc: common.ActivityDoc{
			CreatedBy: authUsername,
			CreatedAt: time.Now(),
		},
	}

	// Create transaction
	_, err = svc.pointTransRepo.Create(ctx, pointTransaction)
	if err != nil {
		return fmt.Errorf("failed to create point transaction: %w", err)
	}

	// Recalculate point balance from all transactions to ensure accuracy
	err = svc.pointTransRepo.RecalculatePointBalanceByPointsCode(ctx, holdingCode, pointsCode)
	if err != nil {
		return fmt.Errorf("failed to recalculate point balance: %w", err)
	}

	return nil
}

// BulkAddPoints bulk adds points for multiple customers (can be used for opening balance, adjustments, or promotions)
func (svc DebtorHttpService) BulkAddPoints(holdingCode string, pointsList []models.OpeningBalancePointRequest, authUsername string) (models.BulkImportPointResult, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	result := models.BulkImportPointResult{
		Total:        len(pointsList),
		FailedItems:  []models.BulkImportPointFailedItem{},
		SuccessItems: []string{},
	}

	// Process each item
	for _, item := range pointsList {
		// Validate item
		if item.PointsCode == "" {
			result.Failed++
			result.FailedItems = append(result.FailedItems, models.BulkImportPointFailedItem{
				PointsCode:  item.PointsCode,
				PointAmount: item.PointAmount,
				Reason:      "pointsCode is required",
			})
			continue
		}

		if item.PointAmount <= 0 {
			result.Failed++
			result.FailedItems = append(result.FailedItems, models.BulkImportPointFailedItem{
				PointsCode:  item.PointsCode,
				PointAmount: item.PointAmount,
				Reason:      "pointAmount must be greater than 0",
			})
			continue
		}

		// Find debtor by pointscode
		debtor, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "points_code", item.PointsCode)
		if err != nil {
			result.Failed++
			result.FailedItems = append(result.FailedItems, models.BulkImportPointFailedItem{
				PointsCode:  item.PointsCode,
				PointAmount: item.PointAmount,
				Reason:      fmt.Sprintf("failed to find debtor: %v", err),
			})
			continue
		}

		// Check if debtor exists
		if debtor.GuidFixed == "" {
			result.Failed++
			result.FailedItems = append(result.FailedItems, models.BulkImportPointFailedItem{
				PointsCode:  item.PointsCode,
				PointAmount: item.PointAmount,
				Reason:      "debtor not found",
			})
			continue
		}

		currentBalance := debtor.PointBalance

		// Set default description if not provided
		description := item.Description
		if description == "" {
			description = "Manual point adjustment (bulk)"
		}

		// Create point transaction for manual adjustment
		pointTransaction := models.PointTransactionDoc{
			PointTransactionData: models.PointTransactionData{
				HoldingCodeentity: common.HoldingCodeentity{HoldingCode: holdingCode},
				PointTransactionInfo: models.PointTransactionInfo{
					DocIdentity: common.DocIdentity{GuidFixed: utils.NewGUID()},
					PointTransaction: models.PointTransaction{
						TransactionDocNo: "ADJ-" + item.PointsCode + "-" + time.Now().Format("20060102150405"),
						TransactionDate:  time.Now(),
						DebtorCode:       debtor.Code,
						PointsCode:       item.PointsCode,
						TransactionType:  3, // Manual Adjustment / Opening Balance
						PointAmount:      item.PointAmount,
						BalanceBefore:    currentBalance,
						BalanceAfter:     currentBalance + item.PointAmount,
						Description:      description,
					},
				},
			},
			ActivityDoc: common.ActivityDoc{
				CreatedBy: authUsername,
				CreatedAt: time.Now(),
			},
		}

		// Create transaction
		_, err = svc.pointTransRepo.Create(ctx, pointTransaction)
		if err != nil {
			result.Failed++
			result.FailedItems = append(result.FailedItems, models.BulkImportPointFailedItem{
				PointsCode:  item.PointsCode,
				PointAmount: item.PointAmount,
				Reason:      fmt.Sprintf("failed to create transaction: %v", err),
			})
			continue
		}

		// Recalculate point balance from all transactions to ensure accuracy
		err = svc.pointTransRepo.RecalculatePointBalanceByPointsCode(ctx, holdingCode, item.PointsCode)
		if err != nil {
			result.Failed++
			result.FailedItems = append(result.FailedItems, models.BulkImportPointFailedItem{
				PointsCode:  item.PointsCode,
				PointAmount: item.PointAmount,
				Reason:      fmt.Sprintf("failed to recalculate balance: %v", err),
			})
			continue
		}

		// Success
		result.Success++
		result.SuccessItems = append(result.SuccessItems, item.PointsCode)
	}

	return result, nil
}

// DeleteManualPointTransaction deletes a manual point transaction and recalculates balance
func (svc DebtorHttpService) DeleteManualPointTransaction(holdingCode string, pointsCode string, docNo string, authUsername string) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Validate inputs
	if pointsCode == "" {
		return fmt.Errorf("pointsCode is required")
	}
	if docNo == "" {
		return fmt.Errorf("docNo is required")
	}

	// Delete the manual transaction (this function validates it's type 3 and recalculates balance)
	err := svc.pointTransRepo.DeleteManualPointTransaction(ctx, holdingCode, pointsCode, docNo, authUsername)
	if err != nil {
		return err
	}

	return nil
}
