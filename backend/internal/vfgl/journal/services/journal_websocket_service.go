package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	documentimageModel "smlcloudplatform/internal/documentwarehouse/documentimage/models"
	documentimageRepo "smlcloudplatform/internal/documentwarehouse/documentimage/repositories"
	"smlcloudplatform/internal/vfgl/journal/models"
	"smlcloudplatform/internal/vfgl/journal/repositories"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
)

type IJournalWebsocketService interface {
	PubDoc(shopID string, processID string, screen string, message interface{}) error
	SubDoc(shopID string, processID string, screen string) (<-chan *redis.Message, string, error)

	UnSub(subID string) error

	SaveLastMessage(shopID string, processID string, screen string, message string) error
	GetLastMessage(shopID string, processID string, screen string) (string, error)
	ClearLastMessage(shopID string, processID string) error

	SetWebsocket(shopID string, processID string, screen string) error
	DelWebsocket(shopID string, processID string, screen string) error
	ExistsWebsocket(shopID string, processID string) (bool, error)
	ExpireWebsocket(shopID string, processID string) error

	DocRefPool(shopID string, username string, ws *websocket.Conn) error
	SetDocRefPool(shopID string, username string, docRef string) error
	ExistsDocRefPool(shopID string, docRef string) (bool, error)
	GetDocRefPool(shopID string, docRef string) (string, error)
	GetAllDocRefPool(shopID string) (map[string]string, error)
	DelDocRefPool(shopID string, username string, docRef string) error

	ExistsDocRefUserPool(shopID string, username string) (bool, error)
	GetDocRefUserPool(shopID string, username string) (string, error)

	DocRefSelect(shopID string, username string, docRef string) (bool, error)
	DocRefDeSelect(shopID string, username string) (bool, error)
	DocRefSelectForce(shopID string, username string, docRef string, forceSelect bool) (bool, error)
	DocRefNextSelect(shopID string, username string, status int8) (documentimageModel.DocumentImageInfo, error)
}

type JournalWebsocketService struct {
	cacheChannelDoc     string
	cacheChannelDocRef  string
	cachePoolDocRef     string
	cachePoolDocRefUser string
	cacheMessageName    string
	cacheWebsocketName  string
	cacheExpire         time.Duration
	docImageRepo        documentimageRepo.IDocumentImageRepository
	repoCache           repositories.IJournalCacheRepository
}

func NewJournalWebsocketService(docImageRepo documentimageRepo.IDocumentImageRepository, repoCache repositories.IJournalCacheRepository, cacheExpire time.Duration) *JournalWebsocketService {

	return &JournalWebsocketService{
		cacheChannelDoc:     "chdoc",
		cacheChannelDocRef:  "chdocref",
		cachePoolDocRef:     "wsdocref",
		cachePoolDocRefUser: "wsdocuser",
		cacheMessageName:    "wsmsg",
		cacheWebsocketName:  "wssc",
		docImageRepo:        docImageRepo,
		repoCache:           repoCache,
		cacheExpire:         cacheExpire,
	}
}

func (svc JournalWebsocketService) PubDoc(shopID string, processID string, screen string, message interface{}) error {
	channel := svc.getChannelDoc(shopID, processID, screen, svc.cacheChannelDoc)
	return svc.repoCache.Pub(channel, message)
}

func (svc JournalWebsocketService) SubDoc(shopID string, processID string, screen string) (<-chan *redis.Message, string, error) {
	channel := svc.getChannelDoc(shopID, processID, screen, svc.cacheChannelDoc)
	return svc.repoCache.Sub(channel)
}

func (svc JournalWebsocketService) PubDocRef(shopID string, message interface{}) error {
	channel := svc.getChannelDocRef(shopID, svc.cacheChannelDocRef)
	return svc.repoCache.Pub(channel, message)
}

func (svc JournalWebsocketService) SubDocRef(shopID string) (<-chan *redis.Message, string, error) {
	channel := svc.getChannelDocRef(shopID, svc.cacheChannelDocRef)
	return svc.repoCache.Sub(channel)
}

func (svc JournalWebsocketService) UnSub(subID string) error {
	return svc.repoCache.Unsub(subID)
}

func (svc JournalWebsocketService) SaveLastMessage(shopID string, processID string, screen string, message string) error {

	keyVal := screen
	data := map[string]interface{}{
		keyVal: message,
	}
	cacheKeyName := svc.getTagID(shopID, processID, svc.cacheMessageName)
	return svc.repoCache.HSet(cacheKeyName, data)
}

func (svc JournalWebsocketService) GetLastMessage(shopID string, processID string, screen string) (string, error) {

	cacheKeyName := svc.getTagID(shopID, processID, svc.cacheMessageName)
	keyVal := screen
	return svc.repoCache.HGet(cacheKeyName, keyVal)
}

func (svc JournalWebsocketService) ClearLastMessage(shopID string, processID string) error {
	cacheKeyName := svc.getTagID(shopID, processID, svc.cacheMessageName)
	return svc.repoCache.Del(cacheKeyName)
}

func (svc JournalWebsocketService) SetWebsocket(shopID string, processID string, screen string) error {
	cacheKeyName := svc.getTagID(shopID, processID, svc.cacheWebsocketName)

	keyVal := screen
	data := map[string]interface{}{
		keyVal: 1,
	}

	return svc.repoCache.HSet(cacheKeyName, data)
}

func (svc JournalWebsocketService) DelWebsocket(shopID string, processID string, screen string) error {
	cacheKeyName := svc.getTagID(shopID, processID, svc.cacheWebsocketName)
	keyVal := screen
	return svc.repoCache.HDel(cacheKeyName, keyVal)
}

func (svc JournalWebsocketService) ExistsWebsocket(shopID string, processID string) (bool, error) {
	cacheKeyName := svc.getTagID(shopID, processID, svc.cacheWebsocketName)
	return svc.repoCache.Exists(cacheKeyName)
}

func (svc JournalWebsocketService) ExpireWebsocket(shopID string, processID string) error {
	cacheKeyName := svc.getTagID(shopID, processID, svc.cacheWebsocketName)
	return svc.repoCache.Expire(cacheKeyName, svc.cacheExpire)
}

// doc ref
func (svc JournalWebsocketService) SetDocRefPool(shopID string, username string, docRef string) error {

	if len(docRef) < 1 {
		return errors.New("doc ref is empty")
	}

	cacheKeyDocRef := svc.getTagID(shopID, "", svc.cachePoolDocRef)
	cacheKeyUser := svc.getTagID(shopID, "", svc.cachePoolDocRefUser)

	isDocRefSelected, err := svc.repoCache.HExists(cacheKeyDocRef, docRef)
	if err != nil {
		return err
	}

	if isDocRefSelected {
		return errors.New("doc ref is selected")
	}

	isDocUserSelected, err := svc.repoCache.HExists(cacheKeyUser, username)
	if err != nil {
		return err
	}

	if isDocUserSelected {
		return errors.New("user is selected")
	}

	dataUser := map[string]interface{}{
		username: docRef,
	}

	svc.repoCache.HSet(cacheKeyUser, dataUser)

	if err != nil {
		return err
	}

	dataDocRef := map[string]interface{}{
		docRef: username,
	}

	return svc.repoCache.HSet(cacheKeyDocRef, dataDocRef)
}

func (svc JournalWebsocketService) ExistsDocRefPool(shopID string, docRef string) (bool, error) {
	cacheKeyName := svc.getTagID(shopID, "", svc.cachePoolDocRef)
	return svc.repoCache.HExists(cacheKeyName, docRef)
}

func (svc JournalWebsocketService) GetDocRefPool(shopID string, docRef string) (string, error) {
	cacheKeyName := svc.getTagID(shopID, "", svc.cachePoolDocRef)
	return svc.repoCache.HGet(cacheKeyName, docRef)
}

func (svc JournalWebsocketService) GetAllDocRefPool(shopID string) (map[string]string, error) {
	cacheKeyName := svc.getTagID(shopID, "", svc.cachePoolDocRef)
	return svc.repoCache.HGetAll(cacheKeyName)
}

func (svc JournalWebsocketService) DelDocRefPool(shopID string, username string, docRef string) error {
	cacheKeyDocRef := svc.getTagID(shopID, "", svc.cachePoolDocRef)
	cacheKeyUser := svc.getTagID(shopID, "", svc.cachePoolDocRefUser)

	err := svc.repoCache.HDel(cacheKeyUser, username)

	if err != nil {
		return err
	}

	return svc.repoCache.HDel(cacheKeyDocRef, docRef)
}

// user
func (svc JournalWebsocketService) ExistsDocRefUserPool(shopID string, username string) (bool, error) {
	cacheKeyName := svc.getTagID(shopID, "", svc.cachePoolDocRefUser)
	return svc.repoCache.HExists(cacheKeyName, username)
}

func (svc JournalWebsocketService) GetDocRefUserPool(shopID string, username string) (string, error) {
	cacheKeyName := svc.getTagID(shopID, "", svc.cachePoolDocRefUser)
	return svc.repoCache.HGet(cacheKeyName, username)
}

func (svc JournalWebsocketService) getChannelDocRef(shopID string, prefix string) string {
	tempChannel := fmt.Sprintf("%s-%s", prefix, shopID)
	return tempChannel
}

func (svc JournalWebsocketService) getChannelDoc(shopID string, processID string, prefix string, screen string) string {
	tempID := svc.getTagID(shopID, processID, prefix)
	return fmt.Sprintf("%s:%s", tempID, screen)
}

func (JournalWebsocketService) getTagID(shopID string, processID string, prefix string) string {
	// tempID := utils.FastHash(fmt.Sprintf("%s%s", shopID, processID))
	tempID := fmt.Sprintf("%s-%s%s", prefix, shopID, processID)
	return tempID
}

func (svc JournalWebsocketService) DocRefDeSelect(shopID string, username string) (bool, error) {

	docRef, err := svc.GetDocRefUserPool(shopID, username)
	if err != nil {
		return false, err
	}

	if docRef == "" {
		return false, errors.New("user is not selected")
	}

	err = svc.DelDocRefPool(shopID, username, docRef)

	if err != nil {
		return false, err
	}

	// send websocket to user
	svc.ClearLastMessage(shopID, username)

	tempDocRef, _ := json.Marshal(models.JournalRef{
		DocRef: "",
	})

	svc.PubDoc(shopID, username, "form", tempDocRef)

	docRefEvent := models.DocRefEvent{
		DocRef:   docRef,
		Username: username,
		Status:   "deselected",
	}

	err = svc.pubDocRefSelect(shopID, docRefEvent)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (svc JournalWebsocketService) DocRefSelectForce(shopID string, username string, docRef string, forceSelect bool) (bool, error) {
	if forceSelect {
		svc.DocRefDeSelect(shopID, username)
	}

	return svc.DocRefSelect(shopID, username, docRef)
}

func (svc JournalWebsocketService) DocRefSelect(shopID string, username string, docRef string) (bool, error) {
	isExists, err := svc.ExistsDocRefPool(shopID, docRef)

	if err != nil {
		return false, err
	}

	if isExists {
		return false, errors.New("user is selected")
	}

	err = svc.SetDocRefPool(shopID, username, docRef)

	if err != nil {
		return false, err
	}

	// send websocket to user
	tempDocRef, _ := json.Marshal(models.JournalRef{
		DocRef: docRef,
	})
	svc.PubDoc(shopID, username, "form", tempDocRef)
	err = svc.SaveLastMessage(shopID, username, "form", string(tempDocRef))
	if err != nil {
		return false, err
	}

	docRefEvent := models.DocRefEvent{
		DocRef:   docRef,
		Username: username,
		Status:   "selected",
	}

	err = svc.pubDocRefSelect(shopID, docRefEvent)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (svc JournalWebsocketService) pubDocRefSelect(shopID string, docRefEvent models.DocRefEvent) error {
	tempData, err := json.Marshal(docRefEvent)
	if err != nil {
		return err
	}

	err = svc.PubDocRef(shopID, tempData)

	if err != nil {
		return err
	}

	return nil
}

func (svc JournalWebsocketService) DocRefPool(shopID string, username string, ws *websocket.Conn) error {

	cacheMsg, subID, err := svc.SubDocRef(shopID)

	if err != nil {
		return err
	}

	defer func(ws *websocket.Conn, svc JournalWebsocketService) {
		ws.Close()
		svc.UnSub(subID)
	}(ws, svc)

	for {
		temp := <-cacheMsg
		if temp != nil {

			err = ws.WriteMessage(websocket.TextMessage, []byte(temp.Payload))
			if err != nil {
				return err
			}
		}
	}
}

func (svc JournalWebsocketService) DocRefNextSelect(shopID string, username string, status int8) (documentimageModel.DocumentImageInfo, error) {
	docList, err := svc.GetAllDocRefPool(shopID)

	if err != nil {
		return documentimageModel.DocumentImageInfo{}, err
	}

	docRefList := []string{}
	for docRef := range docList {
		docRefList = append(docRefList, docRef)
	}

	filters := map[string]interface{}{
		"status": status,
	}

	if len(docRefList) > 0 {
		filters["documentref"] = bson.M{
			"$nin": docRefList,
		}
	}

	pageable := micromodels.Pageable{
		Query: "",
		Page:  1,
		Limit: 30,
	}

	tempNextDocImage, _, err := svc.docImageRepo.FindPageFilter(context.Background(), shopID, filters, []string{}, pageable)

	if err != nil {
		return documentimageModel.DocumentImageInfo{}, err
	}

	totalDoc := len(tempNextDocImage)

	randNum := rand.New(rand.NewSource(time.Now().UnixNano()))

	randIdx := randNum.Intn(totalDoc)

	return tempNextDocImage[randIdx], nil

}
