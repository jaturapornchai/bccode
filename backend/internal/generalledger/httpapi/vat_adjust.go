package httpapi

import (
	"context"

	gl "smlcloudplatform/internal/generalledger"
	"smlcloudplatform/pkg/microservice"
)

// vatAdjustScreen is the menu id of "ปรับปรุงภาษีซื้อ" (Champ 5498): reading the list needs the screen, saving needs its update right.
const vatAdjustScreen = "purchase-vat-adjust"

// listVatAdjust - GET /gl/v2/vat-adjustments?docdatefrom=&docdateto=&invdatefrom=&invdateto=&partnerfrom=&partnerto=&page=&limit=
// Purchase VAT rows of posted journals (filter + paging in SQL, generalledger/vat_adjust.go); read-only.
func (h *Http) listVatAdjust(ctx context.Context, request microservice.IContext, scope requestScope) error {
	if !allowed(scope.Permissions, vatAdjustScreen, "") {
		return fail(request, 403, "ไม่มีสิทธิ์อ่านข้อมูลบัญชีนี้")
	}
	version, err := h.startRead(ctx, scope.Scope, request)
	if err != nil {
		return failure(request, err)
	}
	filter := gl.VatAdjustFilter{
		DocDateFrom: request.QueryParam("docdatefrom"), DocDateTo: request.QueryParam("docdateto"),
		InvDateFrom: request.QueryParam("invdatefrom"), InvDateTo: request.QueryParam("invdateto"),
		PartnerFrom: request.QueryParam("partnerfrom"), PartnerTo: request.QueryParam("partnerto"),
	}
	page, err := h.pg.VatAdjustList(ctx, scope.Scope, filter, pageNumber(request.QueryParam("page"), 1, 1000000), pageNumber(request.QueryParam("limit"), 100, 500))
	if err != nil {
		return failure(request, err)
	}
	if err = h.endRead(ctx, scope.Scope, version); err != nil {
		return failure(request, err)
	}
	page.Sequence = version
	return response(request, page)
}
