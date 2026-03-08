package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/reportquery"
	"smlcloudplatform/internal/transaction/smltransaction/models"
	"smlcloudplatform/internal/transaction/smltransaction/repositories"
	"smlcloudplatform/internal/utils"
	"strings"
	"text/template"
	"time"

	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type ISMLTransactionHttpService interface {
	CreateSMLTransaction(shopID string, authUsername string, smlRequest models.SMLTransactionRequest) (string, error)
	SaveInBatch(shopID string, authUsername string, dataReq models.SMLTransactionBulkRequest) ([]string, error)
	DeleteSMLTransaction(shopID string, authUsername string, smlKeyRequest models.SMLTransactionKeyRequest) ([]string, error)
	QueryFilter(filters bson.M, pageable micromodels.Pageable) ([]map[string]interface{}, mongopagination.PaginationData, error)

	QueryFilter2(paramQuery map[string]interface{}, pageable micromodels.Pageable) ([]map[string]interface{}, mongopagination.PaginationData, error)
}

type SMLTransactionHttpService struct {
	repo         repositories.ISMLTransactionRepository
	mqRepo       repositories.ISMLTransactionMessageQueueRepository
	indexCreated map[string]struct{}
}

func NewSMLTransactionHttpService(repo repositories.ISMLTransactionRepository, mqRepo repositories.ISMLTransactionMessageQueueRepository) *SMLTransactionHttpService {

	insSvc := &SMLTransactionHttpService{
		repo:         repo,
		mqRepo:       mqRepo,
		indexCreated: map[string]struct{}{},
	}

	return insSvc
}

func (svc SMLTransactionHttpService) CreateSMLTransaction(shopID string, authUsername string, smlRequest models.SMLTransactionRequest) (string, error) {
	// guid, err := svc.save(shopID, authUsername, smlRequest)

	// if err != nil {
	// 	return "", err
	// }

	// svc.mqRepo.Save(smlRequest)

	// return guid, nil

	collectionName := svc.getCollectionName(smlRequest.Collection)

	_, collectionIndexExists := svc.indexCreated[collectionName]
	if !collectionIndexExists {
		_, err := svc.repo.CreateIndex(collectionName, smlRequest.KeyID)
		if err == nil {
			svc.indexCreated[collectionName] = struct{}{}
		}
	}

	_, ok := smlRequest.Body[smlRequest.KeyID]

	if !ok || smlRequest.Body[smlRequest.KeyID] == nil {
		return "", fmt.Errorf("KeyID is not found in body")
	}

	deleteFilter := bson.M{smlRequest.KeyID: smlRequest.Body[smlRequest.KeyID]}
	err := svc.repo.Delete(collectionName, shopID, authUsername, deleteFilter)

	if err != nil {
		return "", err
	}

	tempData := svc.createBody(shopID, authUsername, smlRequest.Body)
	guid, err := svc.repo.Create(collectionName, tempData)

	if err != nil {
		return "", err
	}

	err = svc.mqRepo.Save(smlRequest)

	if err != nil {
		return "", err
	}

	return guid, nil
}

func (svc SMLTransactionHttpService) save(shopID string, authUsername string, smlRequest models.SMLTransactionRequest) (string, error) {
	collectionName := svc.getCollectionName(smlRequest.Collection)
	findDoc, err := svc.repo.FindByDocIndentityKey(collectionName, shopID, smlRequest.KeyID, smlRequest.Body[smlRequest.KeyID])

	if err != nil {
		return "", err
	}

	_, ok := findDoc[smlRequest.KeyID]

	if ok || findDoc[smlRequest.KeyID] != nil {
		guid, err := svc.update(shopID, authUsername, findDoc, smlRequest)

		if err != nil {
			return "", err
		}

		svc.mqRepo.Save(smlRequest)
		return guid, nil
	}

	guid, err := svc.create(shopID, authUsername, smlRequest)

	if err != nil {
		return "", err
	}

	return guid, nil
}

func (svc SMLTransactionHttpService) update(shopID string, authUsername string, findDoc map[string]interface{}, smlRequest models.SMLTransactionRequest) (string, error) {
	collectionName := svc.getCollectionName(smlRequest.Collection)

	guidFixed := fmt.Sprintf("%v", findDoc["guidfixed"])
	docData := smlRequest.Body
	docData["shopid"] = findDoc["shopid"]
	docData["guidfixed"] = guidFixed
	docData["createdby"] = findDoc["createdby"]
	docData["createdat"] = findDoc["createdat"]

	docData["updatedby"] = authUsername
	docData["updatedat"] = time.Now()

	err := svc.repo.Update(collectionName, shopID, guidFixed, docData)

	if err != nil {
		return "", err
	}

	return guidFixed, nil
}

func (svc SMLTransactionHttpService) create(shopID string, authUsername string, smlRequest models.SMLTransactionRequest) (string, error) {
	collectionName := svc.getCollectionName(smlRequest.Collection)

	docData := smlRequest.Body
	newGuidFixed := utils.NewGUID()

	docData["shopid"] = shopID
	docData["guidfixed"] = newGuidFixed
	docData["createdby"] = authUsername
	docData["createdat"] = time.Now()

	_, err := svc.repo.Create(collectionName, docData)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc SMLTransactionHttpService) createBody(shopID string, authUsername string, bodyRequest map[string]interface{}) map[string]interface{} {

	docData := bodyRequest
	newGuidFixed := utils.NewGUID()

	docData["shopid"] = shopID
	docData["guidfixed"] = newGuidFixed
	docData["createdby"] = authUsername
	docData["createdat"] = time.Now()

	return docData
}

func (svc SMLTransactionHttpService) SaveInBatch(shopID string, authUsername string, dataReq models.SMLTransactionBulkRequest) ([]string, error) {
	collectionName := svc.getCollectionName(dataReq.Collection)
	_, collectionIndexExists := svc.indexCreated[collectionName]
	if !collectionIndexExists {
		_, err := svc.repo.CreateIndex(collectionName, dataReq.KeyID)
		if err == nil {
			svc.indexCreated[collectionName] = struct{}{}
		}
	}

	identityKeys := []string{}
	tempData := []map[string]interface{}{}
	for _, smlRequest := range dataReq.Body {
		_, ok := smlRequest[dataReq.KeyID]

		if !ok || smlRequest[dataReq.KeyID] == nil {
			continue
		}

		identityKeys = append(identityKeys, smlRequest[dataReq.KeyID].(string))
		createData := svc.createBody(shopID, authUsername, smlRequest)
		tempData = append(tempData, createData)
	}

	if len(identityKeys) == 0 {
		return []string{}, fmt.Errorf("body request is empty")
	}

	err := svc.repo.Transaction(func(ctx context.Context) error {

		filters := map[string]interface{}{
			dataReq.KeyID: bson.M{"$in": identityKeys},
		}
		err := svc.repo.Delete(collectionName, shopID, authUsername, filters)

		if err != nil {
			return err
		}

		err = svc.repo.CreateInBatch(collectionName, tempData)

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return []string{}, err
	}

	err = svc.mqRepo.Delete(models.SMLTransactionKeyRequest{
		Collection: dataReq.Collection,
		KeyID:      dataReq.KeyID,
		DeleteKeys: identityKeys,
	})

	if err != nil {
		return []string{}, err
	}

	err = svc.mqRepo.BulkSave(models.SMLTransactionBulkRequest{
		Collection: dataReq.Collection,
		KeyID:      dataReq.KeyID,
		Body:       dataReq.Body,
	})

	if err != nil {
		return identityKeys, err
	}

	return identityKeys, nil
}

func (svc SMLTransactionHttpService) SaveInBatchOld(shopID string, authUsername string, dataReq models.SMLTransactionBulkRequest) ([]string, error) {

	guids := []string{}
	tempSaveSuccess := []map[string]interface{}{}
	err := svc.repo.Transaction(func(ctx context.Context) error {
		for _, smlRequest := range dataReq.Body {
			guidFixed, err := svc.CreateSMLTransaction(shopID, authUsername, models.SMLTransactionRequest{
				Collection: dataReq.Collection,
				KeyID:      dataReq.KeyID,
				Body:       smlRequest,
			})

			if err != nil {
				return err
			}

			guids = append(guids, guidFixed)
			tempSaveSuccess = append(tempSaveSuccess, smlRequest)
		}

		return nil
	})

	if err != nil {
		return []string{}, err
	}

	err = svc.mqRepo.BulkSave(models.SMLTransactionBulkRequest{
		Collection: dataReq.Collection,
		KeyID:      dataReq.KeyID,
		Body:       tempSaveSuccess,
	})

	if err != nil {
		return guids, err
	}

	return guids, nil
}

func (svc SMLTransactionHttpService) DeleteSMLTransaction(shopID string, authUsername string, smlKeyRequest models.SMLTransactionKeyRequest) ([]string, error) {
	collectionName := svc.getCollectionName(smlKeyRequest.Collection)

	filters := map[string]interface{}{
		smlKeyRequest.KeyID: bson.M{"$in": smlKeyRequest.DeleteKeys},
	}
	err := svc.repo.Delete(collectionName, shopID, authUsername, filters)

	if err != nil {
		return []string{}, err
	}

	err = svc.mqRepo.Delete(smlKeyRequest)

	if err != nil {
		return []string{}, err
	}

	return smlKeyRequest.DeleteKeys, nil
}

func (svc SMLTransactionHttpService) QueryFilter(filters bson.M, pageable micromodels.Pageable) ([]map[string]interface{}, mongopagination.PaginationData, error) {
	// collectionName := svc.getCollectionName("products")
	collectionName := "products"

	paramx := map[string]interface{}{
		"@shop": "2Eh6e3pfWvXTp0yV3CyFEhKPjdI",
	}
	filterx, err := reportquery.ReplacePlaceholdersInMap(filters, &paramx)

	if err != nil {
		return []map[string]interface{}{}, mongopagination.PaginationData{}, err
	}

	docList, pagination, err := svc.repo.Filter(collectionName, filterx, pageable)

	if err != nil {
		return docList, pagination, err
	}

	return docList, pagination, nil
}

func (svc SMLTransactionHttpService) QueryFilter2(paramQuery map[string]interface{}, pageable micromodels.Pageable) ([]map[string]interface{}, mongopagination.PaginationData, error) {
	collectionName := "products"

	tempQuery := map[string]interface{}{}
	for key, value := range paramQuery {
		escaped, err := json.Marshal(value)
		if err != nil {
			panic(err)
		}
		tempQuery[key] = string(escaped)
	}

	rawQuery := `{
		"itemcode": "{{.itemcode}}"
	}`

	tmpl, err := template.New("myTemplate").Parse(rawQuery)

	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, tempQuery)
	if err != nil {
		panic(err)
	}

	replacedText := buf.String()

	var filters bson.M
	// Convert the JSON object to a bson.M object
	err = bson.UnmarshalExtJSON([]byte(replacedText), true, &filters)

	if err != nil {
		panic(err)
	}

	docList, pagination, err := svc.repo.Filter(collectionName, filters, pageable)

	if err != nil {
		return docList, pagination, err
	}

	return docList, pagination, nil
}

func (svc SMLTransactionHttpService) getCollectionName(collectionName string) string {
	return "smlx" + strings.Trim(collectionName, " ")
}
