package microservice

import (
	"encoding/json"
	"fmt"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/pkg/microservice/models"
	"time"
)

// startAsyncTaskConsumer read async task message from message queue and execute with handler
func (ms *Microservice) startAsyncTaskConsumer(path string, cacheConfig config.ICacherConfig, mqConfig config.IMQConfig, h ServiceHandleFunc) error {

	topic := escapeName(path)
	mq := NewMQ(mqConfig, ms.Logger)
	ms.Logger.Debugf("Create Topic \"%s\".", topic)
	err := mq.CreateTopicR(topic, 5, 1, time.Hour*24*30) // retain message for 30 days
	if err != nil {
		ms.Logger.Error("Failed on Create Topic.", err)
		return err
	}

	ms.Consume(mqConfig.URI(), topic, "atask", -1, func(ctx IContext) error {
		message := map[string]interface{}{}
		err := json.Unmarshal([]byte(ctx.ReadInput()), &message)
		if err != nil {
			return err
		}

		userInfoStr := message["userInfo"].(string)
		ref, _ := message["ref"].(string)
		input, _ := message["input"].(string)

		userInfo := models.UserInfo{}
		err = json.Unmarshal([]byte(userInfoStr), &userInfo)

		if err != nil {
			return err
		}

		return h(NewAsyncTaskContext(ms, cacheConfig, userInfo, ref, input))
	})

	return nil
}

// handleAsyncTaskRequest accept async task request and send it to message queue
func (ms *Microservice) handleAsyncTaskRequest(path string, cacheConfig config.ICacherConfig, mqConfig config.IMQConfig, ctx IContext) error {
	topic := escapeName(path)

	// 1. Read Input
	input := ctx.ReadInput()

	// 2. Generate REF
	ref := fmt.Sprintf("atask-%s", randString())

	// 3. Set Status in Cache
	cacher := ctx.Cacher(cacheConfig)
	status := map[string]interface{}{
		"status": "processing",
	}
	expire := time.Minute * 30
	cacher.Set(ref, status, expire)

	userInfoStr, err := json.Marshal(ctx.UserInfo())

	if err != nil {
		return err
	}

	// 4. Send Message to MQ
	prod := ctx.Producer(mqConfig)
	message := map[string]interface{}{
		"userInfo": string(userInfoStr),
		"ref":      ref,
		"input":    input,
	}
	prod.SendMessage(topic, "", message)

	// 5. Response REF
	res := map[string]string{
		"ref": ref,
	}
	ctx.Response(http.StatusOK, res)
	return nil
}

func (ms *Microservice) handleAsyncTaskResponse(path string, cacheConfig config.ICacherConfig, ctx IContext) error {
	// 1. ReadInput (REF from query string)
	ref := ctx.QueryParam("ref")

	// 2. Read Status from Cache
	cacher := ctx.Cacher(cacheConfig)
	statusJS, err := cacher.Get(ref)
	if err != nil {
		return err
	}

	// 3. Return Status
	status := map[string]interface{}{}
	err = json.Unmarshal([]byte(statusJS), &status)
	if err != nil {
		return err
	}
	ctx.Response(http.StatusOK, status)
	return nil
}

// AsyncPOST register async task service for HTTP POST
func (ms *Microservice) AsyncPOST(path string, cacheConfig config.ICacherConfig, mqConfig config.IMQConfig, h ServiceHandleFunc) error {
	ms.Logger.Debugf("Register HTTP Handler Async POST %s", path)
	err := ms.startAsyncTaskConsumer(path, cacheConfig, mqConfig, h)
	if err != nil {
		return err
	}
	ms.GET(path, func(ctx IContext) error {
		return ms.handleAsyncTaskResponse(path, cacheConfig, ctx)
	})

	ms.POST(path, func(ctx IContext) error {
		return ms.handleAsyncTaskRequest(path, cacheConfig, mqConfig, ctx)
	})
	return nil
}

// AsyncPUT register async task service for HTTP PUT
func (ms *Microservice) AsyncPUT(path string, cacheConfig config.ICacherConfig, mqConfig config.IMQConfig, h ServiceHandleFunc) {
	ms.startAsyncTaskConsumer(path, cacheConfig, mqConfig, h)
	ms.GET(path, func(ctx IContext) error {
		return ms.handleAsyncTaskResponse(path, cacheConfig, ctx)
	})
	ms.PUT(path, func(ctx IContext) error {
		return ms.handleAsyncTaskRequest(path, cacheConfig, mqConfig, ctx)
	})
}
