package httpapi

import (
	"context"
	"smlcloudplatform/pkg/microservice"
)

func (h *Http) getJournalReview(ctx context.Context, request microservice.IContext, scope requestScope) error {
	j, err := h.store.JournalForAuthorization(ctx, scope.Scope, request.Param("id"))
	if err != nil {
		return failure(request, err)
	}
	if !allowed(scope.Permissions, journalScreen(j.BookCode, j.Kind, ""), "") {
		return fail(request, 403, "ไม่มีสิทธิ์อ่านรายการบัญชีนี้")
	}
	version, err := h.startRead(ctx, scope.Scope, request)
	if err != nil {
		return failure(request, err)
	}
	result, err := h.pg.JournalReview(ctx, scope.Scope, j.ID)
	if err != nil {
		return failure(request, err)
	}
	if err = h.endRead(ctx, scope.Scope, version); err != nil {
		return failure(request, err)
	}
	return response(request, result)
}
