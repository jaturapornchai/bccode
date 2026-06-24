package aichat

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type openaiModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type openaiModelList struct {
	Object string        `json:"object"`
	Data   []openaiModel `json:"data"`
}

// OpenAIGatewayListModels handles GET /goapi/api/aichat/v1/models
func OpenAIGatewayListModels(c echo.Context) error {
	now := time.Now().Unix()
	return c.JSON(http.StatusOK, openaiModelList{
		Object: "list",
		Data: []openaiModel{
			{ID: "nongkung", Object: "model", Created: now, OwnedBy: "bc-account"},
			{ID: "nongkung-react", Object: "model", Created: now, OwnedBy: "bc-account"},
		},
	})
}
