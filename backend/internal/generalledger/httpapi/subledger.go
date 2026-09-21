package httpapi

import (
	"context"
	"smlcloudplatform/pkg/microservice"
)

func canReadJournalSupport(permissions map[string]bool) bool {
	for _, screen := range []string{"jv-journal", "uv-journal", "sv-journal", "rv-journal", "pv-journal", "gl-opening-balance", "gl-post", "gl-unpost"} {
		if allowed(permissions, screen, "") {
			return true
		}
	}
	return false
}

func (h *Http) listJournalSupport(ctx context.Context, request microservice.IContext, scope requestScope) error {
	if !canReadJournalSupport(scope.Permissions) {
		return fail(request, 403, "ไม่มีสิทธิ์อ่านข้อมูลบัญชีนี้")
	}
	version, err := h.startRead(ctx, scope.Scope, request)
	if err != nil {
		return failure(request, err)
	}
	page, err := h.pg.SubledgerList(ctx, scope.Scope, request.QueryParam("kind"), request.QueryParam("q"), pageNumber(request.QueryParam("page"), 1, 1000000), pageNumber(request.QueryParam("limit"), 50, 1000), request.QueryParam("asof"))
	if err != nil {
		return failure(request, err)
	}
	if err = h.endRead(ctx, scope.Scope, version); err != nil {
		return failure(request, err)
	}
	page.Sequence = version
	return response(request, page)
}
