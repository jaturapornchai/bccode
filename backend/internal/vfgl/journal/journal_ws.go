package journal

import (
	"encoding/json"
	"net/http"
	documentimageRepo "smlcloudplatform/internal/documentwarehouse/documentimage/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/vfgl/journal/config"
	"smlcloudplatform/internal/vfgl/journal/models"
	"smlcloudplatform/internal/vfgl/journal/repositories"
	"smlcloudplatform/internal/vfgl/journal/services"
	"smlcloudplatform/pkg/microservice"
	"time"
	"unicode/utf8"

	msConfig "smlcloudplatform/internal/config"

	"github.com/gorilla/websocket"
)

type JournalWs struct {
	ms           *microservice.Microservice
	cfg          msConfig.IConfig
	svcWebsocket services.IJournalWebsocketService
}

func NewJournalWs(ms *microservice.Microservice, cfg msConfig.IConfig) JournalWs {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	docImageRepo := documentimageRepo.NewDocumentImageRepository(pst)
	cacheRepo := repositories.NewJournalCacheRepository(cache)
	svcWebsocket := services.NewJournalWebsocketService(docImageRepo, cacheRepo, time.Duration(30)*time.Minute)

	return JournalWs{
		ms:           ms,
		cfg:          cfg,
		svcWebsocket: svcWebsocket,
	}
}

func (h JournalWs) RegisterHttp() {
	h.ms.GET("/gl/journal/ws/image", h.WebsocketImage)
	h.ms.GET("/gl/journal/ws/form", h.WebsocketForm)

	h.ms.GET("/gl/journal/ws/docref", h.WebsocketDocRefPool)

	h.ms.GET("/gl/journal/docref/selected", h.GetAllDocRefPool)
	h.ms.POST("/gl/journal/docref/select", h.SelectDocRefPool)
	h.ms.POST("/gl/journal/docref/deselect", h.DeSelectDocRefPool)
	h.ms.POST("/gl/journal/docref/next", h.NextSelectDocumentRef)

	h.ms.GET("/gl/journal/user-docref", h.GetUserDocRef)
	h.ms.GET("/gl/journal/docref-user", h.GetDocRefUser)

}

func (h JournalWs) WebsocketImage(ctx microservice.IContext) error {

	screenName := config.WEBSOCKET_SCREEN_IMAGE

	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	username := userInfo.Username

	if username == "" {
		ctx.Response(http.StatusBadRequest, "username parameter missing")
		return nil
	}

	socketID := utils.NewGUID()

	ws, err := h.ms.Websocket(socketID, ctx.ResponseWriter(), ctx.Request())

	if err != nil {
		return err
	}

	err = h.svcWebsocket.SetWebsocket(shopID, username, screenName)

	if err != nil {
		return err
	}

	err = h.svcWebsocket.ExpireWebsocket(shopID, username)

	if err != nil {
		return err
	}

	cacheMsg, subID, err := h.svcWebsocket.SubDoc(shopID, username, screenName)

	if err != nil {
		return err
	}

	singClose := make(chan struct{})
	defer func() {
		ws.Close()
		h.ms.WebsocketClose(socketID)
		h.svcWebsocket.UnSub(subID)
		h.svcWebsocket.DelWebsocket(shopID, username, screenName)
		h.ClearDocRef(shopID, username)
	}()

	go func() {

		defer func() {
			singClose <- struct{}{}
		}()

		for {
			_, _, err = ws.ReadMessage()
			if err != nil {
				return
			}
		}
	}()

	// Send to client
	lastMessageForm, err := h.svcWebsocket.GetLastMessage(shopID, username, config.WEBSOCKET_SCREEN_FORM)
	if err != nil {
		h.ms.Logger.Error(err.Error())
	}

	lastMessage, err := h.svcWebsocket.GetLastMessage(shopID, username, screenName)
	if err != nil {
		h.ms.Logger.Error(err.Error())
	}

	tempMessage := map[string]interface{}{
		"docref": "",
	}

	if lastMessageForm != "" {
		tempForm := map[string]interface{}{}
		err = json.Unmarshal([]byte(lastMessageForm), &tempForm)
		if err != nil {
			h.ms.Logger.Error(err.Error())
		}

		if tempDocref, ok := tempForm["docref"]; ok {
			tempMessage["docref"] = tempDocref
		}
	}

	if lastMessage != "" {
		tempImage := map[string]interface{}{}

		err = json.Unmarshal([]byte(lastMessage), &tempImage)
		if err != nil {
			h.ms.Logger.Error(err.Error())
		}

		if tempEvent, ok := tempImage["event"]; ok {
			tempMessage["event"] = tempEvent
		}

		if tempPayload, ok := tempImage["payload"]; ok {
			tempMessage["payload"] = tempPayload
		}
	}

	tempMessageText, err := json.Marshal(tempMessage)
	if err != nil {
		h.ms.Logger.Error(err.Error())
	}

	err = ws.WriteMessage(websocket.TextMessage, []byte(tempMessageText))

	if err != nil {
		h.ms.Logger.Error(err.Error())
	}

	for {
		h.svcWebsocket.ExpireWebsocket(shopID, username)

		// if err != nil {
		// 	return nil
		// }

		temp := <-cacheMsg
		if temp != nil {

			err = ws.WriteMessage(websocket.TextMessage, []byte(temp.Payload))
			if err != nil {
				return err
			}

			h.svcWebsocket.SaveLastMessage(shopID, username, screenName, temp.Payload)
		}

	}
}

func (h JournalWs) WebsocketForm(ctx microservice.IContext) error {

	screenName := config.WEBSOCKET_SCREEN_FORM
	sendScreenName := config.WEBSOCKET_SCREEN_IMAGE
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	username := userInfo.Username

	if username == "" {
		ctx.Response(http.StatusBadRequest, "username parameter missing")
		return nil
	}

	socketID := utils.NewID()

	ws, err := h.ms.Websocket(socketID, ctx.ResponseWriter(), ctx.Request())

	if err != nil {
		return err
	}

	err = h.svcWebsocket.SetWebsocket(shopID, username, screenName)

	if err != nil {
		return err
	}

	err = h.svcWebsocket.ExpireWebsocket(shopID, username)

	if err != nil {
		return err
	}

	cacheMsg, subID, err := h.svcWebsocket.SubDoc(shopID, username, screenName)

	if err != nil {
		return err
	}

	sigClose := make(chan struct{})
	defer func() {
		ws.Close()
		h.ms.WebsocketClose(socketID)
		h.svcWebsocket.UnSub(subID)
		h.svcWebsocket.DelWebsocket(shopID, username, screenName)
		h.ClearDocRef(shopID, username)
	}()

	// Receive from client
	go func(ws *websocket.Conn, sigClose chan struct{}) {
		defer func() {
			sigClose <- struct{}{}
		}()

		for {
			h.svcWebsocket.ExpireWebsocket(shopID, username)

			journalEvent := models.JournalEvent{}
			err := ws.ReadJSON(&journalEvent)

			if err != nil {
				return
			}

			tempRef, _ := json.Marshal(journalEvent)
			h.svcWebsocket.PubDoc(shopID, username, sendScreenName, tempRef)

			switch journalEvent.Event {
			case "save":
				h.svcWebsocket.ClearLastMessage(shopID, username)
				//clear
				h.ClearDocRef(shopID, username)
			}
		}

	}(ws, sigClose)

	// Send to client
	lastMessage, err := h.svcWebsocket.GetLastMessage(shopID, username, screenName)
	if err != nil {
		h.ms.Logger.Error(err.Error())
	}

	if lastMessage != "" {
		err = ws.WriteMessage(websocket.TextMessage, []byte(lastMessage))

		if err != nil {
			h.ms.Logger.Error(err.Error())
		}
	}

	for {
		select {
		case temp := <-cacheMsg:
			if temp != nil {
				err = ws.WriteMessage(websocket.TextMessage, []byte(temp.Payload))
				if err != nil {
					return err
				}
			}
		case <-sigClose:
			return nil
		}

	}

}

func (h JournalWs) WebsocketDocRefPool(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	username := userInfo.Username

	if username == "" {
		ctx.Response(http.StatusBadRequest, "username is missing")
		return nil
	}

	socketID := utils.NewID()

	ws, err := h.ms.Websocket(socketID, ctx.ResponseWriter(), ctx.Request())

	if err != nil {
		return err
	}

	return h.svcWebsocket.DocRefPool(shopID, username, ws)

}

func (h JournalWs) ClearDocRef(shopID string, username string) error {

	// isExists, err := h.svcWebsocket.ExistsWebsocket(shopID, username)
	// if err != nil {
	// 	h.ms.Logger.Error(err.Error())
	// }

	// if isExists {
	// 	return nil
	// }

	// lastMessage, err := h.svcWebsocket.GetLastMessage(shopID, username, "form")
	// if err != nil {
	// 	h.ms.Logger.Error(err.Error())
	// }

	// if lastMessage != "" {
	// 	journalRef := models.JournalRef{}
	// 	json.Unmarshal([]byte(lastMessage), &journalRef)

	// 	err = h.svcWebsocket.DelDocRefPool(shopID, username, journalRef.DocRef)
	// 	if err != nil {
	// 		h.ms.Logger.Error(err.Error())
	// 	}
	// }

	return nil
}

// List Document Ref selected godoc
// @Summary		List Document Ref selected
// @Description	For List Document Ref selected
// @Tags		WSDocumentRef
// @Success		200	{object}	common.ApiResponse
// @Failure		400 {object}	models.AuthResponseFailed
// @Accept 		json
// @Security     AccessToken
// @Router		/selected [get]
func (h JournalWs) GetAllDocRefPool(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	result, err := h.svcWebsocket.GetAllDocRefPool(shopID)

	docRefPool := []models.DocRefPool{}
	for tempDocRef, tempUsername := range result {
		docRefPool = append(docRefPool, models.DocRefPool{
			DocRef:   tempDocRef,
			Username: tempUsername,
		})
	}

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    docRefPool,
	})
	return nil
}

func (h JournalWs) GetUserDocRef(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	username := userInfo.Username

	result, err := h.svcWebsocket.GetDocRefUserPool(shopID, username)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data: &models.JournalRef{
			DocRef: result,
		},
	})
	return nil
}

func (h JournalWs) GetDocRefUser(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	docRef := ctx.QueryParam("docref")

	if docRef == "" {
		ctx.ResponseError(http.StatusBadRequest, "docref is empty")
		return nil
	}

	if utf8.RuneCountInString(docRef) > 100 {
		ctx.ResponseError(http.StatusBadRequest, "docref invalid")
		return nil
	}

	result, err := h.svcWebsocket.GetDocRefPool(shopID, docRef)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data: map[string]string{
			"username": result,
		},
	})
	return nil
}

// List Document Ref selected godoc
// @Summary		List Document Ref selected
// @Description	For List Document Ref selected
// @Tags		WSDocumentRef
// @Param		User  body      models.JournalRef  true  "JournalRef body"
// @Success		200	{object}	common.ApiResponse
// @Failure		400 {object}	models.AuthResponseFailed
// @Accept 		json
// @Security     AccessToken
// @Router		/gl/journal/docref/select [post]
func (h JournalWs) SelectDocRefPool(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	username := userInfo.Username

	input := ctx.ReadInput()

	docReq := &models.JournalRef{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	forceSelectRaw := ctx.QueryParam("force")
	forceSelect := false
	if forceSelectRaw == "1" {
		forceSelect = true
	}

	result, err := h.svcWebsocket.DocRefSelectForce(shopID, username, docReq.DocRef, forceSelect)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    result,
	})
	return nil
}

// List Document Ref selected godoc
// @Summary		List Document Ref selected
// @Description	For List Document Ref selected
// @Tags		WSDocumentRef
// @Param		User  body      models.JournalRef  true  "JournalRef body"
// @Success		200	{object}	common.ApiResponse
// @Failure		400 {object}	models.AuthResponseFailed
// @Accept 		json
// @Security     AccessToken
// @Router		/gl/journal/docref/deselect [post]
func (h JournalWs) DeSelectDocRefPool(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	username := userInfo.Username

	// input := ctx.ReadInput()

	// docReq := &models.JournalRef{}
	// err := json.Unmarshal([]byte(input), &docReq)

	// if err != nil {
	// 	ctx.ResponseError(400, err.Error())
	// 	return err
	// }

	result, err := h.svcWebsocket.DocRefDeSelect(shopID, username)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    result,
	})
	return nil
}

func (h JournalWs) WebsocketConnectCount(ctx microservice.IContext) error {

	ctx.Response(http.StatusOK, h.ms.WebsocketCount())

	return nil
}

// List Document Ref selected godoc
// @Summary		List Document Ref selected
// @Description	For List Document Ref selected
// @Tags		WSDocumentRef
// @Success		200	{object}	common.ApiResponse
// @Failure		400 {object}	models.AuthResponseFailed
// @Accept 		json
// @Security     AccessToken
// @Router		/gl/journal/docref/next [post]
func (h JournalWs) NextSelectDocumentRef(ctx microservice.IContext) error {
	// userInfo := ctx.UserInfo()
	// shopID := userInfo.ShopID
	// username := userInfo.Username

	// doc, err := h.svcWebsocket.DocRefNextSelect(shopID, username, 0)

	// if err != nil {
	// 	ctx.ResponseError(http.StatusBadRequest, err.Error())
	// 	return err
	// }

	// _, err = h.svcWebsocket.DocRefSelectForce(shopID, username, doc.DocumentRef, true)

	// if err != nil {
	// 	ctx.ResponseError(http.StatusBadRequest, err.Error())
	// 	return err
	// }

	// ctx.Response(http.StatusOK, common.ApiResponse{
	// 	Success: true,
	// 	Data: map[string]string{
	// 		"documentref": doc.DocumentRef,
	// 	},
	// })
	return nil
}
