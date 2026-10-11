package httpapi

import (
	"context"

	gl "smlcloudplatform/internal/generalledger"
	"smlcloudplatform/pkg/microservice"
)

func canReadJournalSupport(permissions map[string]bool) bool {
	for _, screen := range []string{journalsScreen, openingScreen, postingScreen} {
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
	return h.subledgerPage(ctx, request, scope, request.QueryParam("kind"))
}

// bookBankScreen คือสิทธิ์ของจอสมุดเงินฝาก (menu id "book-bank"): อ่านรายการสมุดผ่าน GET /gl/v2/bank-accounts และสั่งคำสั่ง resource "bank-accounts";
// ผู้ที่อ่านรายละเอียดประกอบในใบสำคัญได้ก็อ่านรายการสมุดได้เหมือนเดิม (ใช้เลือกสมุดในใบ)
const bookBankScreen = "book-bank"

func (h *Http) listBankAccounts(ctx context.Context, request microservice.IContext, scope requestScope) error {
	if !allowed(scope.Permissions, bookBankScreen, "") && !canReadJournalSupport(scope.Permissions) {
		return fail(request, 403, "ไม่มีสิทธิ์อ่านข้อมูลบัญชีนี้")
	}
	return h.subledgerPage(ctx, request, scope, gl.BankAccountKind)
}

func (h *Http) subledgerPage(ctx context.Context, request microservice.IContext, scope requestScope, kind string) error {
	version, err := h.startRead(ctx, scope.Scope, request)
	if err != nil {
		return failure(request, err)
	}
	page, err := h.pg.SubledgerList(ctx, scope.Scope, kind, request.QueryParam("q"), pageNumber(request.QueryParam("page"), 1, 1000000), pageNumber(request.QueryParam("limit"), 50, 1000), request.QueryParam("asof"))
	if err != nil {
		return failure(request, err)
	}
	if err = h.endRead(ctx, scope.Scope, version); err != nil {
		return failure(request, err)
	}
	page.Sequence = version
	return response(request, page)
}
