package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"smlcloudplatform/internal/mcptoken"
	"smlcloudplatform/pkg/microservice"
)

func (h *Http) registerTokenAPI() {
	group := h.ms.Echo().Group("/integration/gl/v2")
	group.GET("/reports/:report", h.tokenAPI(h.report, false))
	group.GET("/:resource", h.tokenAPI(h.list, false))
	group.GET("/:resource/:id", h.tokenAPI(h.get, false))
	group.POST("/command", h.tokenAPI(h.command, true))
}

func (h *Http) tokenAPI(run func(microservice.IContext) error, write bool) echo.HandlerFunc {
	return h.tokenAPIWithAuth(run, write, mcptoken.AuthenticateAPI, mcptoken.ResolveCompany)
}

func (h *Http) tokenAPIWithAuth(run func(microservice.IContext) error, write bool, auth func(context.Context, string) (mcptoken.Principal, error), resolve func(context.Context, mcptoken.Principal, string) (mcptoken.Principal, error)) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "no-store")
		parts := strings.Fields(c.Request().Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Response().Header().Set("WWW-Authenticate", `Bearer realm="BC GL API"`)
			return c.JSON(401, map[string]any{"success": false, "message": "Invalid API token"})
		}
		ctx, cancel := context.WithTimeout(c.Request().Context(), 35*time.Second)
		defer cancel()
		p, err := auth(ctx, parts[1])
		if err != nil || p.Kind != "api" || (p.Mode != "readonly" && p.Mode != "readwrite") {
			c.Response().Header().Set("WWW-Authenticate", `Bearer realm="BC GL API"`)
			return c.JSON(401, map[string]any{"success": false, "message": "Invalid API token"})
		}
		if write && p.Mode != "readwrite" {
			return c.JSON(403, map[string]any{"success": false, "message": "Token is readonly"})
		}
		var many []string
		if header := c.Request().Header.Get("X-BC-Company-Codes"); header != "" {
			many = strings.Split(header, ",")
		}
		companies, err := mcptoken.CompanySelection(p.Mode, c.Request().Header.Get("X-BC-Company-Code"), many)
		if err != nil {
			return c.JSON(403, map[string]any{"success": false, "message": err.Error()})
		}
		principals := make([]mcptoken.Principal, 0, len(companies))
		// Authorize the complete selection before reading any company's data.
		for _, code := range companies {
			resolved, err := resolve(ctx, p, code)
			if err != nil {
				return c.JSON(403, map[string]any{"success": false, "message": "Company denied; specify allowed company codes"})
			}
			principals = append(principals, resolved)
		}
		results := make([]any, 0, len(principals))
		allSucceeded := true
		for _, principal := range principals {
			c.Set("UserInfo", principal.User)
			request := &mcpGLContext{IContext: microservice.NewHTTPContext(h.ms, c), request: c.Request().WithContext(ctx), params: map[string]string{"resource": c.Param("resource"), "id": c.Param("id"), "report": c.Param("report")}, tokenID: p.ID, tokenKind: "api"}
			if err := run(request); err != nil {
				return c.JSON(http.StatusServiceUnavailable, map[string]any{"success": false, "message": "Unable to process request"})
			}
			if request.status == 0 {
				return c.JSON(http.StatusServiceUnavailable, map[string]any{"success": false, "message": "Empty API response"})
			}
			if len(principals) == 1 {
				return c.JSON(request.status, request.payload)
			}
			results = append(results, map[string]any{"companyCode": principal.User.BusinessCode, "response": request.payload})
			allSucceeded = allSucceeded && request.status < 400
		}
		return c.JSON(200, map[string]any{"success": allSucceeded, "data": map[string]any{"companies": results}})
	}
}
