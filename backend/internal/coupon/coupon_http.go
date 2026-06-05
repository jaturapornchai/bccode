package coupon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/coupon/models"
	"smlcloudplatform/internal/coupon/repositories"
	"smlcloudplatform/internal/coupon/services"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/pkg/microservice"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type ICouponHttp interface{}

type CouponHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.ICouponHttpService
}

func NewCouponHttp(ms *microservice.Microservice, cfg config.IConfig) CouponHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewCouponRepository(pst)
	reservationRepo := repositories.NewCouponReservationRepository(pst)
	usageHistoryRepo := repositories.NewCouponUsageHistoryRepository(pst)

	// เพิ่ม dependencies สำหรับการตรวจสอบสินค้า
	productBarcodeRepo := productbarcode_repositories.NewProductBarcodeRepository(pst, cache)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)

	svc := services.NewCouponHttpService(repo, reservationRepo, usageHistoryRepo, productBarcodeRepo, masterSyncCacheRepo)

	return CouponHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h CouponHttp) RegisterHttp() {
	h.ms.POST("/coupon/bulk", h.SaveBulk)
	h.ms.POST("/coupon/bulk/preview", h.PreviewBulkImport)
	h.ms.GET("/coupon/import/template", h.DownloadImportTemplate)
	h.ms.GET("/coupon", h.SearchCoupon)
	h.ms.POST("/coupon", h.CreateCoupon)
	h.ms.GET("/coupon/:id", h.InfoCoupon)
	h.ms.GET("/coupon/code/:code", h.GetCouponByCode)

	h.ms.PUT("/coupon/:id", h.UpdateCoupon)
	h.ms.DELETE("/coupon/:id", h.DeleteCoupon)
	h.ms.GET("/coupon-type", h.InfoCouponType)
	h.ms.GET("/coupon-status", h.InfoCouponStatus)

	// Coupon calculation endpoint
	h.ms.POST("/coupon/calculate", h.CalculateCoupons)

	// Coupon reservation endpoints
	h.ms.POST("/coupon/:id/reserve", h.ReserveCoupon)
	h.ms.DELETE("/coupon/:id/reserve", h.CancelReserveCoupon)
	h.ms.POST("/coupon/:id/use", h.UseCoupon)
	h.ms.GET("/coupon/:id/availability", h.CheckCouponAvailability)
	h.ms.POST("/coupon/:id/availability/check", h.CheckCouponAvailabilityAdvanced)

	// Coupon reservation status checking endpoints
	h.ms.GET("/coupon/reservation/status", h.CheckReservationStatus)
	h.ms.POST("/coupon/reservation/lookup", h.LookupReservation)

	// Usage History endpoints
	h.ms.POST("/coupon/usage-history", h.CreateUsageHistory)
	h.ms.GET("/coupon/usage-history/search", h.GetUsageHistory)
	h.ms.GET("/coupon/:coupon_id/usage-history", h.GetUsageHistoryByCoupon)
	h.ms.GET("/coupon/customer/usage-history", h.GetUsageHistoryByCustomer)
	h.ms.GET("/coupon/sale-invoice/usage-history", h.GetUsageHistoryBySaleInvoice)
	h.ms.GET("/coupon/transaction/usage-history", h.GetUsageHistoryByTransactionID)

	// Excel Import endpoint
	h.ms.POST("/coupon/upload-excel", h.UploadExcelImport)
	h.ms.GET("/coupon/import-status/:batch_id", h.GetImportStatus)

}

// Create Coupon godoc
// @Summary		สร้าง coupon
// @Description สร้าง coupon
// @Tags		Coupon
// @Param		Coupon  body      models.Coupon  true  "coupon"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /coupon [post]
func (h CouponHttp) CreateCoupon(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.Coupon{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	// Validation for new model structure
	if docReq.CouponType < models.CouponTypeValueDiscount || docReq.CouponType > models.CouponTypeCashVoucher {
		ctx.ResponseError(400, "invalid coupon type")
		return nil
	}

	// Validate usage count logic
	if docReq.IsOneTimeUse {
		// หากเป็นการใช้ครั้งเดียว ไม่ต้องใส่ MaxUsageCount หรือ MaxUsageCountPerCustomer
		docReq.MaxUsageCount = 1
		docReq.MaxUsageCountPerCustomer = 1
	} else {
		// ตรวจสอบ Global vs Per-Customer mode
		if len(docReq.CustomerCodes) == 0 {
			// Global mode: ต้องมี MaxUsageCount
			if docReq.MaxUsageCount <= 0 {
				ctx.ResponseError(400, "MaxUsageCount is required when CustomerCodes is empty (global mode)")
				return nil
			}
		} else {
			// Per-Customer mode: ต้องมี MaxUsageCountPerCustomer
			if docReq.MaxUsageCountPerCustomer <= 0 {
				ctx.ResponseError(400, "MaxUsageCountPerCustomer is required when CustomerCodes is specified")
				return nil
			}
		}
	}

	idx, err := h.svc.CreateCoupon(holdingCode, authUsername, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      idx,
	})
	return nil
}

// Update Coupon godoc
// @Summary		แก้ไข coupon
// @Description แก้ไข coupon
// @Tags		Coupon
// @Param		id  path      string  true  "Coupon ID"
// @Param		Coupon  body      models.Coupon  true  "coupon"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /coupon/{id} [put]
func (h CouponHttp) UpdateCoupon(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.Coupon{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	// Validation for new model structure
	if docReq.CouponType < models.CouponTypeValueDiscount || docReq.CouponType > models.CouponTypeCashVoucher {
		ctx.ResponseError(400, "invalid coupon type")
		return nil
	}

	// Validate usage count logic
	if docReq.IsOneTimeUse {
		// หากเป็นการใช้ครั้งเดียว ไม่ต้องใส่ MaxUsageCount หรือ MaxUsageCountPerCustomer
		docReq.MaxUsageCount = 1
		docReq.MaxUsageCountPerCustomer = 1
	} else {
		// ตรวจสอบ Global vs Per-Customer mode
		if len(docReq.CustomerCodes) == 0 {
			// Global mode: ต้องมี MaxUsageCount
			if docReq.MaxUsageCount <= 0 {
				ctx.ResponseError(400, "MaxUsageCount is required when CustomerCodes is empty (global mode)")
				return nil
			}
		} else {
			// Per-Customer mode: ต้องมี MaxUsageCountPerCustomer
			if docReq.MaxUsageCountPerCustomer <= 0 {
				ctx.ResponseError(400, "MaxUsageCountPerCustomer is required when CustomerCodes is specified")
				return nil
			}
		}
	}

	err = h.svc.UpdateCoupon(id, holdingCode, authUsername, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      id,
	})

	return nil
}

// Delete Coupon godoc
// @Summary		ลบ coupon
// @Description ลบ coupon
// @Tags		Coupon
// @Param		id  path      string  true  "Journal ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /coupon/{id} [delete]
func (h CouponHttp) DeleteCoupon(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteCoupon(id, holdingCode, authUsername)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})

	return nil
}

// Get Coupon godoc
// @Summary		แสดงรายละเอียด coupon
// @Description แสดงรายละเอียด coupon
// @Tags		Coupon
// @Param		id  path      string  true  "Journal Id"
// @Accept 		json
// @Success		200	{object}	models.CouponInfoResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /coupon/{id} [get]
func (h CouponHttp) InfoCoupon(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get Coupon %v", id)
	doc, err := h.svc.InfoCoupon(id, holdingCode)

	if err != nil {
		h.ms.Logger.Errorf("Error getting document %v: %v", id, err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// GetCouponByCode godoc
// @Summary		ดึงข้อมูลคูปองตาม coupon code
// @Description ดึงข้อมูลคูปองตาม coupon code
// @Tags		Coupon
// @Param		code	path	string	true	"Coupon Code"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse{data=models.CouponInfo}
// @Failure		400 {object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /coupon/code/{code} [get]
func (h CouponHttp) GetCouponByCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	code := ctx.Param("code")

	h.ms.Logger.Debugf("Get Coupon by code %v", code)
	doc, err := h.svc.InfoCouponByCode(code, holdingCode)

	if err != nil {
		h.ms.Logger.Errorf("Error getting coupon by code %v: %v", code, err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// List Coupon godoc
// @Summary		แสดงรายการ coupon
// @Description แสดงรายการ coupon
// @Tags		Coupon
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Size"
// @Accept 		json
// @Success		200	{object}	models.CouponPageResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /coupon [get]
func (h CouponHttp) SearchCoupon(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	q := ctx.QueryParam("q")
	docList, err := h.svc.SearchCoupon(holdingCode, q)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    docList,
	})
	return nil
}

// Create Coupon godoc
// @Summary		นำเข้าข้อมูล coupon
// @Description นำเข้าข้อมูล coupon แบบ bulk import
// @Tags		Coupon
// @Param		Coupons  body      []models.Coupon  true  "coupons"
// @Accept 		json
// @Success		201	{object}	common.BulkInsertResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /coupon/bulk [post]
func (h CouponHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	dataReq := []models.Coupon{}
	err := json.Unmarshal([]byte(input), &dataReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	// Validate each coupon in bulk import
	for i, coupon := range dataReq {
		// Validation for new model structure
		if coupon.CouponType < models.CouponTypeValueDiscount || coupon.CouponType > models.CouponTypeCashVoucher {
			ctx.ResponseError(400, "invalid coupon type at index "+strconv.Itoa(i))
			return nil
		}

		// Validate usage count logic
		if coupon.IsOneTimeUse {
			// หากเป็นการใช้ครั้งเดียว ไม่ต้องใส่ MaxUsageCount หรือ MaxUsageCountPerCustomer
			dataReq[i].MaxUsageCount = 1
			dataReq[i].MaxUsageCountPerCustomer = 1
		} else {
			// ตรวจสอบ Global vs Per-Customer mode
			if len(coupon.CustomerCodes) == 0 {
				// Global mode: ต้องมี MaxUsageCount
				if coupon.MaxUsageCount <= 0 {
					ctx.ResponseError(400, "MaxUsageCount is required when CustomerCodes is empty (global mode) at index "+strconv.Itoa(i))
					return nil
				}
			} else {
				// Per-Customer mode: ต้องมี MaxUsageCountPerCustomer
				if coupon.MaxUsageCountPerCustomer <= 0 {
					ctx.ResponseError(400, "MaxUsageCountPerCustomer is required when CustomerCodes is specified at index "+strconv.Itoa(i))
					return nil
				}
			}
		}
	}

	bulkResponse, err := h.svc.SaveInBatch(holdingCode, authUsername, dataReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	ctx.Response(
		http.StatusCreated,
		common.BulkResponse{
			Success:    true,
			BulkImport: bulkResponse,
		},
	)

	return nil
}

// Preview Bulk Import godoc
// @Summary		ตรวจสอบข้อมูลก่อนนำเข้าคูปองแบบกลุ่ม
// @Description ตรวจสอบความถูกต้องและข้อมูลซ้ำก่อนนำเข้าจริง
// @Tags		Coupon
// @Accept 		json
// @Produce		json
// @Param 		data body []models.Coupon true "ข้อมูลคูปองที่ต้องการตรวจสอบ"
// @Success		200	{object}	common.BulkPreviewResponse
// @Failure		400 {object}	common.AuthResponseFailed
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /coupon/bulk/preview [post]
func (h CouponHttp) PreviewBulkImport(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()
	dataReq := []models.Coupon{}
	err := json.Unmarshal([]byte(input), &dataReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	// Validate each coupon
	validationErrors := []string{}
	for i, coupon := range dataReq {
		// Validation for new model structure
		if coupon.CouponType < models.CouponTypeValueDiscount || coupon.CouponType > models.CouponTypeCashVoucher {
			validationErrors = append(validationErrors, "invalid coupon type at index "+strconv.Itoa(i))
		}

		// Validate usage count logic
		if coupon.IsOneTimeUse {
			dataReq[i].MaxUsageCount = 1
			dataReq[i].MaxUsageCountPerCustomer = 1
		} else {
			if len(coupon.CustomerCodes) == 0 {
				if coupon.MaxUsageCount <= 0 {
					validationErrors = append(validationErrors, "MaxUsageCount is required when CustomerCodes is empty (global mode) at index "+strconv.Itoa(i))
				}
			} else {
				if coupon.MaxUsageCountPerCustomer <= 0 {
					validationErrors = append(validationErrors, "MaxUsageCountPerCustomer is required when CustomerCodes is specified at index "+strconv.Itoa(i))
				}
			}
		}
	}

	// Check for duplicates in the database
	previewResponse, err := h.svc.PreviewBulkImport(holdingCode, dataReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	// Add validation errors to response
	previewResponse.ValidationErrors = validationErrors

	ctx.Response(http.StatusOK, common.BulkPreviewResponse{
		Success:      true,
		PreviewData:  previewResponse,
		CanProceed:   len(validationErrors) == 0,
		ErrorCount:   len(validationErrors),
		TotalRecords: len(dataReq),
	})

	return nil
}

// Download Import Template godoc
// @Summary		ดาวน์โหลดเทมเพลตสำหรับการนำเข้าคูปอง
// @Description ดาวน์โหลดไฟล์ Excel template สำหรับการนำเข้าคูปอง
// @Tags		Coupon
// @Produce		application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Success		200	{file}	file
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /coupon/import/template [get]
func (h CouponHttp) DownloadImportTemplate(ctx microservice.IContext) error {
	// สร้าง Excel file ใหม่
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			h.ms.Logger.Errorf("Error closing Excel file: %v", err)
		}
	}()

	// ตั้งชื่อ sheet
	sheetName := "Coupon Import Template"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		ctx.ResponseError(500, "Error creating Excel sheet: "+err.Error())
		return err
	}
	f.SetActiveSheet(index)
	_ = f.DeleteSheet("Sheet1") // ลบ sheet เริ่มต้น (ignore error)

	// กำหนดหัวตาราง (ใช้ชื่อฟิลด์จาก model)
	headers := []string{
		"coupon_code",              // รหัสคูปอง
		"names",                    // ชื่อคูปอง (string เดียว)
		"couponvalue",              // มูลค่าคูปอง
		"issued_date",              // วันที่ออกคูปอง (YYYY-MM-DD)
		"expiry_date",              // วันหมดอายุ (YYYY-MM-DD)
		"coupon_type",              // ประเภทคูปอง (0,1,2)
		"customer_codes",           // รหัสลูกค้า (ถ้ามีหลายตัวใช้ ;)
		"remark",                   // หมายเหตุ
		"status",                   // สถานะ (0,1)
		"isonetimeuse",             // ใช้ครั้งเดียว (true/false)
		"maxusagecount",            // จำนวนครั้งสูงสุดรวม
		"maxusagecountpercustomer", // จำนวนครั้งสูงสุดต่อลูกค้า
		"product_codes",            // รายการ Code ของสินค้า (ถ้ามีหลายตัวใช้ ;)
		"groupcodes",               // รายการ GroupCode (ถ้ามีหลายตัวใช้ ;)
		"group_subone_codes",       // รายการ GroupsuboneCode (ถ้ามีหลายตัวใช้ ;)
		"group_subtwo_codes",       // รายการ GroupsubtwoCode (ถ้ามีหลายตัวใช้ ;)
		"brand_codes",              // รายการ BrandCode (ถ้ามีหลายตัวใช้ ;)
		"design_codes",             // รายการ DesignCode (ถ้ามีหลายตัวใช้ ;)
		"model_codes",              // รายการ ModelCode (ถ้ามีหลายตัวใช้ ;)
		"pattern_codes",            // รายการ PatternCode (ถ้ามีหลายตัวใช้ ;)
		"grade_codes",              // รายการ GradeCode (ถ้ามีหลายตัวใช้ ;)
		"category_codes",           // รายการ CategoryCode (ถ้ามีหลายตัวใช้ ;)
		"class_codes",              // รายการ ClassCode (ถ้ามีหลายตัวใช้ ;)
		"minimum_amount",           // มูลค่าขั้นต่ำของสินค้าที่เข้าเงื่อนไข (0 = ไม่จำกัด)
		"ignore_branchcode",        // รายการสาขาที่ไม่ต้องการใช้คูปอง (ถ้ามีหลายตัวใช้ ;)
	}

	// เขียนหัวตาราง
	for i, header := range headers {
		cell := string(rune('A'+i)) + "1"
		_ = f.SetCellValue(sheetName, cell, header)
	}

	// สร้างข้อมูลตัวอย่าง
	sampleData := [][]interface{}{
		{
			"EXAMPLE001",
			"ตัวอย่างคูปองลด 50 บาท", // string เดียว
			50.00,
			time.Now().UTC().Format("2006-01-02"),
			time.Now().UTC().AddDate(0, 6, 0).Format("2006-01-02"),
			0,
			"", // empty = ทุกคนใช้ได้
			"ตัวอย่างคูปองส่วนลด",
			0,
			false,
			100,
			1,
			"PROD001;PROD002", // product_codes
			"",                // groupcodes
			"",                // group_subone_codes
			"",                // group_subtwo_codes
			"BRAND001",        // brand_codes
			"",                // design_codes
			"",                // model_codes
			"",                // pattern_codes
			"",                // grade_codes
			"CAT001;CAT002",   // category_codes
			"",                // class_codes
			2000.00,           // minimum_amount (ขั้นต่ำ 2000 บาท)
			"BR001;BR002",     // ignore_branchcode
		},
		{
			"EXAMPLE002",
			"ตัวอย่างคูปองลด 10%", // string เดียว
			10.00,
			time.Now().UTC().Format("2006-01-02"),
			time.Now().UTC().AddDate(0, 3, 0).Format("2006-01-02"),
			1,
			"CUST001;CUST002", // รหัสลูกค้าเฉพาะ (หลายตัวใช้ ;)
			"ตัวอย่างคูปองส่วนลดเปอร์เซ็นต์",
			0,
			true,
			1,
			1,
			"",              // product_codes
			"GRP001;GRP002", // groupcodes
			"SUB001",        // group_subone_codes
			"SUB201;SUB202", // group_subtwo_codes
			"",              // brand_codes
			"DES001",        // design_codes
			"",              // model_codes
			"",              // pattern_codes
			"GRD001",        // grade_codes
			"",              // category_codes
			"CLS001",        // class_codes
			500.00,          // minimum_amount (ขั้นต่ำ 500 บาท)
			"",              // ignore_branchcode
		},
		{
			"EXAMPLE003",
			"ตัวอย่างคูปองแทนเงินสด", // string เดียว
			500.00,
			time.Now().UTC().Format("2006-01-02"),
			time.Now().UTC().AddDate(0, 12, 0).Format("2006-01-02"),
			2,
			"", // empty = ทุกคนใช้ได้
			"ตัวอย่างคูปองแทนเงินสด",
			0,
			false,
			50,
			2,
			"", // product_codes - ไม่มีเงื่อนไขสินค้า
			"", // groupcodes
			"", // group_subone_codes
			"", // group_subtwo_codes
			"", // brand_codes
			"", // design_codes
			"", // model_codes
			"", // pattern_codes
			"", // grade_codes
			"", // category_codes
			"", // class_codes
			0,  // minimum_amount (0 = ไม่จำกัด)
			"", // ignore_branchcode
		},
	}

	// เขียนข้อมูลตัวอย่าง
	for rowIndex, row := range sampleData {
		for colIndex, value := range row {
			cell := string(rune('A'+colIndex)) + strconv.Itoa(rowIndex+2)
			_ = f.SetCellValue(sheetName, cell, value)
		}
	}

	// สร้าง sheet สำหรับคำแนะนำ
	instructionSheet := "Instructions"
	_, err = f.NewSheet(instructionSheet)
	if err == nil {
		instructions := []string{
			"คำแนะนำการใช้งาน Template การนำเข้าคูปอง",
			"",
			"1. ฟิลด์บังคับ:",
			"   - couponcode: รหัสคูปองที่ไม่ซ้ำกัน",
			"   - names: ชื่อคูปอง (string เดียว)",
			"   - couponvalue: มูลค่าคูปอง (ตัวเลข)",
			"   - issueddate: วันที่ออกคูปอง (YYYY-MM-DD)",
			"   - expirydate: วันหมดอายุ (YYYY-MM-DD)",
			"   - coupontype: ประเภทคูปอง",
			"",
			"2. ประเภทคูปอง (coupontype):",
			"   - 0: ลดตามมูลค่า",
			"   - 1: ลดตามเปอร์เซ็นต์",
			"   - 2: คูปองแทนเงินสด",
			"",
			"3. สถานะ (status):",
			"   - 0: ปกติ (ใช้งานได้)",
			"   - 1: ยกเลิก",
			"",
			"4. รูปแบบข้อมูล:",
			"   - names: ชื่อคูปอง (string เดียว) เช่น 'ส่วนลด 50 บาท'",
			"   - customercodes: รหัสลูกค้า ถ้ามีหลายตัวใช้ ; เช่น 'CUST001;CUST002' หรือ ว่างไว้สำหรับทุกคน",
			"   - isonetimeuse: true/false",
			"   - dates: รูปแบบ YYYY-MM-DD",
			"",
			"5. เงื่อนไขสินค้า (Product Condition):",
			"   - product_codes: รายการรหัสสินค้า เช่น 'PROD001;PROD002'",
			"   - groupcodes: รายการรหัสกลุ่ม เช่น 'GRP001;GRP002'",
			"   - group_subone_codes: รายการรหัสกลุ่มย่อย 1 เช่น 'SUB001;SUB002'",
			"   - group_subtwo_codes: รายการรหัสกลุ่มย่อย 2 เช่น 'SUB201;SUB202'",
			"   - brand_codes: รายการรหัสแบรนด์ เช่น 'BRAND001;BRAND002'",
			"   - design_codes: รายการรหัสดีไซน์ เช่น 'DES001;DES002'",
			"   - model_codes: รายการรหัสโมเดล เช่น 'MOD001;MOD002'",
			"   - pattern_codes: รายการรหัสลวดลาย เช่น 'PAT001;PAT002'",
			"   - grade_codes: รายการรหัสเกรด เช่น 'GRD001;GRD002'",
			"   - category_codes: รายการรหัสหมวดหมู่ เช่น 'CAT001;CAT002'",
			"   - class_codes: รายการรหัสคลาส เช่น 'CLS001;CLS002'",
			"",
			"6. การยกเว้นสาขา (Branch Exclusion):",
			"   - ignore_branchcode: รายการรหัสสาขาที่ไม่ต้องการใช้คูปอง เช่น 'BR001;BR002'",
			"   - ว่าง = ใช้ได้ทุกสาขา",
			"",
			"7. ตัวอย่างการใช้งาน:",
			"   - คูปองทั่วไป: ทุกคอลัมน์เงื่อนไขสินค้า = ว่าง",
			"   - คูปองเฉพาะสินค้า: product_codes='PROD001;PROD002', อื่นๆ = ว่าง",
			"   - คูปองเฉพาะหมวดหมู่: category_codes='CAT001;CAT002', อื่นๆ = ว่าง",
			"   - คูปองแบบผสม: สามารถกรอกหลายคอลัมน์พร้อมกันได้",
			"   - ยกเว้นสาขา: ignore_branchcode='BR001;BR003' = ไม่ใช้ที่สาขา BR001 และ BR003",
			"",
			"8. Logic การใช้งาน:",
			"   - หาก isonetimeuse = true: maxusagecount และ maxusagecountpercustomer จะถูกตั้งเป็น 1 อัตโนมัติ",
			"   - หาก customercodes = ว่าง: ใช้ maxusagecount (Global mode)",
			"   - หาก customercodes มีข้อมูล: ใช้ maxusagecountpercustomer (Per-Customer mode)",
		}

		for i, instruction := range instructions {
			_ = f.SetCellValue(instructionSheet, "A"+strconv.Itoa(i+1), instruction)
		}
	}

	// ปรับความกว้างคอลัมน์
	_ = f.SetColWidth(sheetName, "A", "A", 15) // couponcode
	_ = f.SetColWidth(sheetName, "B", "B", 35) // names (string เดียว)
	_ = f.SetColWidth(sheetName, "C", "C", 15) // couponvalue
	_ = f.SetColWidth(sheetName, "D", "E", 15) // dates
	_ = f.SetColWidth(sheetName, "F", "F", 12) // coupontype
	_ = f.SetColWidth(sheetName, "G", "G", 25) // customercodes
	_ = f.SetColWidth(sheetName, "H", "H", 25) // remark
	_ = f.SetColWidth(sheetName, "I", "I", 10) // status
	_ = f.SetColWidth(sheetName, "J", "J", 15) // isonetimeuse
	_ = f.SetColWidth(sheetName, "K", "L", 20) // usage counts
	_ = f.SetColWidth(sheetName, "M", "M", 20) // product_codes
	_ = f.SetColWidth(sheetName, "N", "N", 18) // groupcodes
	_ = f.SetColWidth(sheetName, "O", "O", 22) // group_subone_codes
	_ = f.SetColWidth(sheetName, "P", "P", 22) // group_subtwo_codes
	_ = f.SetColWidth(sheetName, "Q", "Q", 18) // brand_codes
	_ = f.SetColWidth(sheetName, "R", "R", 18) // design_codes
	_ = f.SetColWidth(sheetName, "S", "S", 18) // model_codes
	_ = f.SetColWidth(sheetName, "T", "T", 20) // pattern_codes
	_ = f.SetColWidth(sheetName, "U", "U", 18) // grade_codes
	_ = f.SetColWidth(sheetName, "V", "V", 20) // category_codes
	_ = f.SetColWidth(sheetName, "W", "W", 18) // class_codes
	_ = f.SetColWidth(sheetName, "X", "X", 25) // ignore_branchcode

	// จัดรูปแบบหัวตาราง
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E2EFDA"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	if err == nil {
		_ = f.SetCellStyle(sheetName, "A1", string(rune('A'+len(headers)-1))+"1", headerStyle)
	}

	// บันทึกไฟล์ลง buffer
	buffer, err := f.WriteToBuffer()
	if err != nil {
		ctx.ResponseError(500, "Error generating Excel file: "+err.Error())
		return err
	}

	// ตั้งค่า headers สำหรับดาวน์โหลดไฟล์
	filename := "coupon-import-template.xlsx"
	w := ctx.ResponseWriter()
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Length", strconv.Itoa(buffer.Len()))

	// ส่งไฟล์ออกไป
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buffer.Bytes())
	return nil
}

// Upload Excel Import godoc
// @Summary		อัปโหลดไฟล์ Excel สำหรับนำเข้าคูปอง
// @Description อัปโหลดไฟล์ Excel เพื่อนำเข้าคูปองแบบ bulk import พร้อมตรวจสอบข้อมูลซ้ำ และประมวลผลเบื้องหลัง
// @Tags		Coupon
// @Accept 		multipart/form-data
// @Param		file formData file true "Excel file (.xlsx)"
// @Param		validate_only formData bool false "เฉพาะตรวจสอบข้อมูล ไม่บันทึก (default: false)"
// @Success		200	{object}	models.CouponExcelImportResponse
// @Failure		400 {object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /coupon/upload-excel [post]
func (h CouponHttp) UploadExcelImport(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	// อ่านไฟล์ Excel จาก form data
	header, err := ctx.FormFile("file")
	if err != nil {
		ctx.ResponseError(400, "ไม่พบไฟล์ Excel: "+err.Error())
		return err
	}

	// เปิดไฟล์
	file, err := header.Open()
	if err != nil {
		ctx.ResponseError(500, "ไม่สามารถเปิดไฟล์ได้: "+err.Error())
		return err
	}
	defer file.Close()

	// ตรวจสอบ file extension
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".xlsx") {
		ctx.ResponseError(400, "รองรับเฉพาะไฟล์ .xlsx เท่านั้น")
		return nil
	}

	// อ่านข้อมูลไฟล์
	fileData, err := io.ReadAll(file)
	if err != nil {
		ctx.ResponseError(500, "ไม่สามารถอ่านไฟล์ได้: "+err.Error())
		return err
	}

	// อ่าน form parameters
	validateOnly := ctx.FormValue("validate_only") == "true"
	overwriteExisting := ctx.FormValue("overwrite_existing") == "true"
	sheetName := ctx.FormValue("sheet_name")
	startRowStr := ctx.FormValue("start_row")

	startRow := 2 // default
	if startRowStr != "" {
		if parsed, err := strconv.Atoi(startRowStr); err == nil && parsed > 0 {
			startRow = parsed
		}
	}

	// สร้าง import request
	importReq := &models.CouponExcelImportRequest{
		FileName:          header.Filename,
		FileData:          fileData,
		SheetName:         sheetName,
		StartRow:          startRow,
		ValidateOnly:      validateOnly,
		OverwriteExisting: overwriteExisting,
	}

	// เรียก service ประมวลผล
	response, err := h.processExcelImportSync(holdingCode, authUsername, *importReq)
	if err != nil {
		ctx.ResponseError(500, "เกิดข้อผิดพลาดในการประมวลผล: "+err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    response,
	})
	return nil
}

// Get Import Status godoc
// @Summary		ดูสถานะการนำเข้าคูปอง
// @Description ดูสถานะการนำเข้าคูปองจากการประมวลผลเบื้องหลัง
// @Tags		Coupon
// @Param		batch_id path string true "Batch ID"
// @Success		200	{object}	models.CouponBatchImportResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /coupon/import-status/{batch_id} [get]
func (h CouponHttp) GetImportStatus(ctx microservice.IContext) error {
	batchID := ctx.Param("batch_id")
	if batchID == "" {
		ctx.ResponseError(400, "batch_id is required")
		return nil
	}

	// ใช้ batch tracking จริง
	status := h.getBatchStatus(batchID)
	if status == nil {
		ctx.ResponseError(404, "ไม่พบ batch_id นี้")
		return nil
	}

	// สร้าง response โดยการคัดลอกข้อมูลจาก status
	response := &models.CouponBatchImportResponse{
		BatchID:       status.BatchID,
		Status:        status.Status,
		Progress:      status.Progress,
		TotalRows:     status.TotalRows,
		ProcessedRows: status.ProcessedRows,
		SuccessCount:  status.SuccessCount,
		ErrorCount:    status.ErrorCount,
		Message:       status.Message,
		UpdatedAt:     status.UpdatedAt,
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    response,
	})
	return nil
}

// Get Coupon Type godoc
// @Summary		แสดงประเภทคูปอง
// @Description แสดงประเภทคูปอง
// @Tags		Coupon
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /coupon-type [get]
func (h CouponHttp) InfoCouponType(ctx microservice.IContext) error {
	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data: map[int8]string{
			0: "ลดตามมูลค่า",
			1: "ลดตามเปอร์เซ็นต์",
			2: "คูปองแทนเงินสด",
		},
	})
	return nil
}

// Get Coupon Status godoc
// @Summary		แสดงสถานะคูปอง
// @Description แสดงสถานะคูปอง
// @Tags		Coupon
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /coupon-status [get]
func (h CouponHttp) InfoCouponStatus(ctx microservice.IContext) error {
	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data: map[int8]string{
			0: "ปกติ",
			1: "ยกเลิก",
		},
	})
	return nil
}

// Check Coupon Availability godoc
// @Summary		ตรวจสอบสถานะและความพร้อมใช้งานของคูปอง
// @Description ตรวจสอบสถานะและความพร้อมใช้งานของคูปอง
// @Tags		Coupon
// @Param		id  path      string  true  "Coupon Code"
// @Param		customer_id  query  string  false  "Customer ID"
// @Accept 		json
// @Success		200	{object}	models.CouponAvailabilityResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /coupon/{id}/availability [get]
func (h CouponHttp) CheckCouponAvailability(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	customerID := ctx.QueryParam("customer_id")

	id := ctx.Param("id")

	availability, err := h.svc.CheckCouponAvailability(id, holdingCode, customerID)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    availability,
	})
	return nil
}

// Check Coupon Availability Advanced godoc
// @Summary		ตรวจสอบความพร้อมใช้งานคูปองแบบขั้นสูง (พร้อมตรวจสอบสินค้าและสาขา)
// @Description ตรวจสอบความพร้อมใช้งานคูปองพร้อมกับการตรวจสอบเงื่อนไขสินค้า, สาขา และคำนวนส่วนลด
// @Tags		Coupon
// @Param		id  path      string  true  "Coupon Code"
// @Param		CheckRequest  body  models.CouponAvailabilityCheckRequest  true  "availability check data"
// @Accept 		json
// @Success		200	{object}	models.CouponAvailabilityAdvancedResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /coupon/{id}/availability/check [post]
func (h CouponHttp) CheckCouponAvailabilityAdvanced(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	id := ctx.Param("id")
	input := ctx.ReadInput()

	checkReq := &models.CouponAvailabilityCheckRequest{}
	err := json.Unmarshal([]byte(input), &checkReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	// Validate required fields
	if checkReq.BranchCode == "" {
		ctx.ResponseError(400, "branchcode is required")
		return nil
	}

	if len(checkReq.Items) == 0 {
		ctx.ResponseError(400, "items are required")
		return nil
	}

	// Validate each item
	for i, item := range checkReq.Items {
		if item.Barcode == "" {
			ctx.ResponseError(400, fmt.Sprintf("barcode is required for item at index %d", i))
			return nil
		}
		if item.Qty <= 0 {
			ctx.ResponseError(400, fmt.Sprintf("qty must be greater than 0 for item at index %d", i))
			return nil
		}
		if item.Price < 0 {
			ctx.ResponseError(400, fmt.Sprintf("price must be non-negative for item at index %d", i))
			return nil
		}
		if item.SumAmount < 0 {
			ctx.ResponseError(400, fmt.Sprintf("sumamount must be non-negative for item at index %d", i))
			return nil
		}
	}

	availability, err := h.svc.CheckCouponAvailabilityAdvanced(id, holdingCode, checkReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    availability,
	})
	return nil
}

// Reserve Coupon godoc
// @Summary		จองคูปองเพื่อใช้งาน
// @Description จองคูปองเพื่อใช้งาน (มีระยะเวลาจำกัด 15 นาที) รองรับทั้งโหมด Global และ Per-Customer
// @Tags		Coupon
// @Param		id  path      string  true  "Coupon Code"
// @Param		ReserveCoupon  body  models.ReserveCouponRequest  true  "reserve data"
// @Accept 		json
// @Success		200	{object}	models.CouponReservationResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /coupon/{id}/reserve [post]
func (h CouponHttp) ReserveCoupon(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	couponCode := ctx.Param("id") // URL parameter `:id` คือ couponCode
	input := ctx.ReadInput()

	reserveReq := &models.ReserveCouponRequest{}
	err := json.Unmarshal([]byte(input), &reserveReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	// Validate required fields
	if reserveReq.CustomerID == "" {
		ctx.ResponseError(400, "customer_id is required")
		return nil
	}

	if reserveReq.TransactionID == "" {
		ctx.ResponseError(400, "transaction_id is required")
		return nil
	}

	// ตรวจสอบความพร้อมใช้งานของคูปองก่อนจอง
	availability, err := h.svc.CheckCouponAvailability(couponCode, holdingCode, reserveReq.CustomerID)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, "ไม่สามารถตรวจสอบความพร้อมใช้งานของคูปอง: "+err.Error())
		return err
	}

	if !availability.Available {
		ctx.ResponseError(http.StatusBadRequest, "ไม่สามารถจองคูปองได้: "+availability.Message)
		return nil
	}

	reservation, err := h.svc.ReserveCoupon(couponCode, holdingCode, *reserveReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    reservation,
	})
	return nil
}

// Cancel Reserve Coupon godoc
// @Summary		ยกเลิกการจองคูปอง
// @Description ยกเลิกการจองคูปอง
// @Tags		Coupon
// @Param		id  path      string  true  "Coupon Code"
// @Param		CancelReserveCoupon  body  models.CancelReserveCouponRequest  true  "cancel reserve data"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /coupon/{id}/reserve [delete]
func (h CouponHttp) CancelReserveCoupon(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	couponCode := ctx.Param("id") // URL parameter `:id` คือ couponCode
	input := ctx.ReadInput()

	cancelReq := &models.CancelReserveCouponRequest{}
	err := json.Unmarshal([]byte(input), &cancelReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	// Validate required fields
	if cancelReq.CustomerID == "" {
		ctx.ResponseError(400, "customer_id is required")
		return nil
	}

	if cancelReq.ReservationID == "" {
		ctx.ResponseError(400, "reservation_id is required")
		return nil
	}

	err = h.svc.CancelReserveCoupon(couponCode, holdingCode, *cancelReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Message: "ยกเลิกการจองเรียบร้อยแล้ว",
	})
	return nil
}

// Use Coupon godoc
// @Summary		ใช้คูปอง
// @Description ใช้คูปอง (ตรวจสอบการจองและใช้งาน) รองรับทั้งโหมด Global และ Per-Customer
// @Tags		Coupon
// @Param		id  path      string  true  "Coupon Code"
// @Param		UseCoupon  body  models.UseCouponRequest  true  "use coupon data"
// @Accept 		json
// @Success		200	{object}	models.UseCouponResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /coupon/{id}/use [post]
func (h CouponHttp) UseCoupon(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	couponCode := ctx.Param("id") // URL parameter `:id` คือ couponCode
	input := ctx.ReadInput()

	useReq := &models.UseCouponRequest{}
	err := json.Unmarshal([]byte(input), &useReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	// Validate required fields
	if useReq.CustomerID == "" {
		ctx.ResponseError(400, "customer_id is required")
		return nil
	}

	if useReq.ReservationID == "" {
		ctx.ResponseError(400, "reservation_id is required")
		return nil
	}

	if useReq.TransactionID == "" {
		ctx.ResponseError(400, "transaction_id is required")
		return nil
	}

	if useReq.SaleInvoiceID == "" {
		ctx.ResponseError(400, "sale_invoice_id is required")
		return nil
	}

	if useReq.SaleInvoiceNumber == "" {
		ctx.ResponseError(400, "sale_invoice_number is required")
		return nil
	}

	if useReq.UseAmount <= 0 {
		ctx.ResponseError(400, "use_amount must be greater than 0")
		return nil
	}

	if useReq.OrderAmount <= 0 {
		ctx.ResponseError(400, "order_amount must be greater than 0")
		return nil
	}

	usageResult, err := h.svc.UseCoupon(couponCode, holdingCode, authUsername, *useReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    usageResult,
	})
	return nil
}

// Calculate Coupons godoc
// @Summary		คำนวนส่วนลดจากคูปอง
// @Description คำนวนส่วนลดและมูลค่าแทนเงินสดจากรายการคูปองที่ส่งมา รองรับ Dynamic Usage Count Logic และตรวจสอบเงื่อนไขสินค้า/สาขา
// @Description use_amount: 0 = ใช้เต็มจำนวน, >0 = ใช้ตามจำนวนที่ระบุ
// @Description branchcode: รหัสสาขาสำหรับตรวจสอบ IgnoreBranchCode
// @Description items: รายการสินค้าสำหรับตรวจสอบเงื่อนไขสินค้าและคำนวนส่วนลด
// @Tags		Coupon
// @Param		CalculateCoupon  body  models.CalculateCouponRequest  true  "calculation data"
// @Accept 		json
// @Success		200	{object}	models.CalculateCouponResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /coupon/calculate [post]
func (h CouponHttp) CalculateCoupons(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	calcReq := &models.CalculateCouponRequest{}
	err := json.Unmarshal([]byte(input), &calcReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	// Validate request
	if calcReq.OrderAmount <= 0 {
		ctx.ResponseError(400, "order_amount must be greater than 0")
		return nil
	}

	if len(calcReq.Coupons) == 0 {
		ctx.ResponseError(400, "coupons array cannot be empty")
		return nil
	}

	// Validate new required fields
	if calcReq.BranchCode == "" {
		ctx.ResponseError(400, "branchcode is required")
		return nil
	}

	if len(calcReq.Items) == 0 {
		ctx.ResponseError(400, "items are required for product condition validation")
		return nil
	}

	// Validate each item
	for i, item := range calcReq.Items {
		if item.Barcode == "" {
			ctx.ResponseError(400, fmt.Sprintf("barcode is required for item at index %d", i))
			return nil
		}
		if item.Qty <= 0 {
			ctx.ResponseError(400, fmt.Sprintf("qty must be greater than 0 for item at index %d", i))
			return nil
		}
		if item.Price < 0 {
			ctx.ResponseError(400, fmt.Sprintf("price must be non-negative for item at index %d", i))
			return nil
		}
		if item.SumAmount < 0 {
			ctx.ResponseError(400, fmt.Sprintf("sumamount must be non-negative for item at index %d", i))
			return nil
		}
	}

	// Validate each coupon in the request
	for i, coupon := range calcReq.Coupons {
		if coupon.CouponCode == "" {
			ctx.ResponseError(400, "coupon_code is required at index "+strconv.Itoa(i))
			return nil
		}

		// use_amount >= 0:
		// - ถ้า use_amount = 0 หรือไม่ระบุ จะใช้เต็มจำนวน (CouponValue)
		// - ถ้า use_amount > 0 จะใช้ตามจำนวนที่ระบุ
		if coupon.UseAmount < 0 {
			ctx.ResponseError(400, "use_amount must be greater than or equal to 0 at index "+strconv.Itoa(i)+". Use 0 for full coupon value")
			return nil
		}
	}

	// Set default customer_id if not provided (for global mode coupons)
	if calcReq.CustomerID == "" {
		calcReq.CustomerID = "ANONYMOUS" // สำหรับคูปองแบบ global mode
	}

	// Calculate coupons
	result, err := h.svc.CalculateCoupons(holdingCode, *calcReq)
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

// CheckReservationStatus godoc
// @Summary		เช็คสถานะการจองคูปอง
// @Description เช็คสถานะการจองคูปองตาม transaction ID
// @Tags		Coupon
// @Param		transaction_id query string true "Transaction ID"
// @Accept 		json
// @Produce 	json
// @Success 	200 {object} common.ApiResponse
// @Failure 	401 {object} common.AuthResponseFailed
// @Security 	AccessToken
// @Router 		/coupon/reservation/status [get]
func (h CouponHttp) CheckReservationStatus(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	transactionID := ctx.QueryParam("transaction_id")

	if transactionID == "" {
		ctx.ResponseError(400, "transaction_id is required")
		return nil
	}

	// เช็คสถานะการจองด้วย transaction ID
	status, err := h.svc.CheckReservationByTransactionID(transactionID, holdingCode)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    status,
	})
	return nil
}

// LookupReservation godoc
// @Summary		ค้นหาข้อมูลการจองคูปอง
// @Description ค้นหาข้อมูลการจองคูปองอย่างละเอียด
// @Tags		Coupon
// @Param		LookupRequest body models.ReservationLookupRequest true "lookup request data"
// @Accept 		json
// @Produce 	json
// @Success 	200 {object} common.ApiResponse
// @Failure 	401 {object} common.AuthResponseFailed
// @Security 	AccessToken
// @Router 		/coupon/reservation/lookup [post]
func (h CouponHttp) LookupReservation(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	input := ctx.ReadInput()

	lookupReq := &models.ReservationLookupRequest{}
	err := json.Unmarshal([]byte(input), &lookupReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if lookupReq.TransactionID == "" {
		ctx.ResponseError(400, "transaction_id is required")
		return nil
	}

	// ค้นหาข้อมูลการจองคูปอง
	reservationInfo, err := h.svc.LookupReservationDetails(holdingCode, *lookupReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    reservationInfo,
	})

	return nil
}

// Usage History Handlers

// @Summary		สร้างประวัติการใช้คูปอง
// @Description บันทึกประวัติการใช้คูปองพร้อมรายละเอียด Sale Invoice และลูกค้า รองรับ Dynamic Usage Count Logic
// @Tags		Coupon
// @Param		UsageHistoryRequest body models.CreateUsageHistoryRequest true "usage history data"
// @Accept 		json
// @Produce 	json
// @Success 	200 {object} common.ApiResponse
// @Failure 	401 {object} common.AuthResponseFailed
// @Security 	AccessToken
// @Router 		/coupon/usage-history [post]
func (h CouponHttp) CreateUsageHistory(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	input := ctx.ReadInput()

	historyReq := &models.CreateUsageHistoryRequest{}
	err := json.Unmarshal([]byte(input), &historyReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	// Validate required fields
	if historyReq.CouponID == "" {
		ctx.ResponseError(400, "coupon_id is required")
		return nil
	}

	if historyReq.CouponCode == "" {
		ctx.ResponseError(400, "coupon_code is required")
		return nil
	}

	if historyReq.CustomerID == "" {
		ctx.ResponseError(400, "customer_id is required")
		return nil
	}

	if historyReq.SaleInvoiceID == "" {
		ctx.ResponseError(400, "sale_invoice_id is required")
		return nil
	}

	if historyReq.SaleInvoiceNumber == "" {
		ctx.ResponseError(400, "sale_invoice_number is required")
		return nil
	}

	if historyReq.TransactionID == "" {
		ctx.ResponseError(400, "transaction_id is required")
		return nil
	}

	if historyReq.UsedAmount <= 0 {
		ctx.ResponseError(400, "used_amount must be greater than 0")
		return nil
	}

	if historyReq.OrderAmount <= 0 {
		ctx.ResponseError(400, "order_amount must be greater than 0")
		return nil
	}

	if historyReq.CouponValue <= 0 {
		ctx.ResponseError(400, "coupon_value must be greater than 0")
		return nil
	}

	// Validate coupon type
	if !historyReq.CouponType.IsValid() {
		ctx.ResponseError(400, "invalid coupon_type")
		return nil
	}

	// Validate discount amounts
	if historyReq.DiscountAmount < 0 {
		ctx.ResponseError(400, "discount_amount must be greater than or equal to 0")
		return nil
	}

	if historyReq.CashVoucherAmount < 0 {
		ctx.ResponseError(400, "cash_voucher_amount must be greater than or equal to 0")
		return nil
	}

	// สร้างประวัติการใช้คูปอง
	historyID, err := h.svc.CreateUsageHistory(holdingCode, authUsername, *historyReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data: map[string]interface{}{
			"usage_history_id": historyID,
			"message":          "บันทึกประวัติการใช้คูปองสำเร็จ",
		},
	})

	return nil
}

// @Summary		ดูประวัติการใช้คูปอง
// @Description ดูประวัติการใช้คูปองตามเงื่อนไขต่างๆ รองรับการค้นหาแบบ Dynamic สำหรับทั้งโหมด Global และ Per-Customer
// @Tags		Coupon
// @Param		page query int false "หน้า (default: 1)"
// @Param		pageSize query int false "จำนวนต่อหน้า (default: 10)"
// @Param		couponCode query string false "รหัสคูปอง"
// @Param		couponID query string false "ID คูปอง"
// @Param		customerID query string false "รหัสลูกค้า"
// @Param		saleInvoiceID query string false "รหัสใบกำกับการขาย"
// @Param		saleInvoiceNumber query string false "หมายเลขใบกำกับการขาย"
// @Param		transactionID query string false "รหัสธุรกรรม"
// @Param		startDate query string false "วันที่เริ่มต้น (YYYY-MM-DD)"
// @Param		endDate query string false "วันที่สิ้นสุด (YYYY-MM-DD)"
// @Accept 		json
// @Produce 	json
// @Success 	200 {object} common.ApiResponse
// @Failure 	401 {object} common.AuthResponseFailed
// @Security 	AccessToken
// @Router 		/coupon/usage-history/search [get]
func (h CouponHttp) GetUsageHistory(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	// รับ parameters จาก query string
	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(ctx.QueryParam("pageSize"))
	if pageSize <= 0 {
		pageSize = 10
	}

	// ป้องกัน pageSize ที่ใหญ่เกินไป
	if pageSize > 100 {
		pageSize = 100
	}

	searchReq := models.CouponUsageHistoryRequest{
		Page:     page,
		PageSize: pageSize,
	}

	// เพิ่ม optional filters
	if couponCode := ctx.QueryParam("couponCode"); couponCode != "" {
		searchReq.CouponCode = couponCode
	}
	if couponID := ctx.QueryParam("couponID"); couponID != "" {
		searchReq.CouponID = couponID
	}
	if customerID := ctx.QueryParam("customerID"); customerID != "" {
		searchReq.CustomerID = customerID
	}
	if saleInvoiceID := ctx.QueryParam("saleInvoiceID"); saleInvoiceID != "" {
		searchReq.SaleInvoiceID = saleInvoiceID
	}
	if saleInvoiceNumber := ctx.QueryParam("saleInvoiceNumber"); saleInvoiceNumber != "" {
		searchReq.SaleInvoiceNumber = saleInvoiceNumber
	}
	if transactionID := ctx.QueryParam("transactionID"); transactionID != "" {
		searchReq.TransactionID = transactionID
	}

	// Date filtering
	if startDateStr := ctx.QueryParam("startDate"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			searchReq.StartDate = &startDate
		}
	}
	if endDateStr := ctx.QueryParam("endDate"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// เพิ่มเวลาไปถึงท้ายวัน
			endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			searchReq.EndDate = &endDate
		}
	}

	// ค้นหาประวัติการใช้คูปอง
	response, err := h.svc.GetUsageHistory(holdingCode, searchReq)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    response,
	})
	return nil
}

// @Summary		ดูประวัติการใช้คูปองตามรหัสคูปอง
// @Description ดูประวัติการใช้คูปองของคูปองใดคูปองหนึ่ง
// @Tags		Coupon
// @Param		coupon_id query string true "Coupon ID"
// @Param		page query int false "หน้า (default: 1)"
// @Param		page_size query int false "จำนวนต่อหน้า (default: 20)"
// @Accept 		json
// @Produce 	json
// @Success 	200 {object} common.ApiResponse
// @Failure 	401 {object} common.AuthResponseFailed
// @Security 	AccessToken
// @Router 		/coupon/usage-history/by-coupon [get]
func (h CouponHttp) GetUsageHistoryByCoupon(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	couponID := ctx.QueryParam("coupon_id")
	pageStr := ctx.QueryParam("page")
	pageSizeStr := ctx.QueryParam("page_size")

	if couponID == "" {
		ctx.ResponseError(400, "coupon_id is required")
		return nil
	}

	// Parse page and pageSize with defaults
	page := 1
	pageSize := 20

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	// ดึงประวัติการใช้
	response, err := h.svc.GetUsageHistoryByCoupon(holdingCode, couponID, page, pageSize)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    response,
	})

	return nil
}

// @Summary		ดูประวัติการใช้คูปองตามรหัสลูกค้า
// @Description ดูประวัติการใช้คูปองของลูกค้าคนใดคนหนึ่ง
// @Tags		Coupon
// @Param		customer_id query string true "Customer ID"
// @Param		page query int false "หน้า (default: 1)"
// @Param		page_size query int false "จำนวนต่อหน้า (default: 20)"
// @Accept 		json
// @Produce 	json
// @Success 	200 {object} common.ApiResponse
// @Failure 	401 {object} common.AuthResponseFailed
// @Security 	AccessToken
// @Router 		/coupon/customer/usage-history [get]
func (h CouponHttp) GetUsageHistoryByCustomer(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	customerID := ctx.QueryParam("customer_id")
	pageStr := ctx.QueryParam("page")
	pageSizeStr := ctx.QueryParam("page_size")

	if customerID == "" {
		ctx.ResponseError(400, "customer_id is required")
		return nil
	}

	// Parse page and pageSize with defaults
	page := 1
	pageSize := 20

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	// ดึงประวัติการใช้
	response, err := h.svc.GetUsageHistoryByCustomer(holdingCode, customerID, page, pageSize)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    response,
	})

	return nil
}

// @Summary		ดูประวัติการใช้คูปองตามรหัสใบกำกับสินค้า
// @Description ดูประวัติการใช้คูปองของใบกำกับสินค้าใดใบหนึ่ง
// @Tags		Coupon
// @Param		sale_invoice_id query string true "Sale Invoice ID"
// @Accept 		json
// @Produce 	json
// @Success 	200 {object} common.ApiResponse
// @Failure 	401 {object} common.AuthResponseFailed
// @Security 	AccessToken
// @Router 		/coupon/sale-invoice/usage-history [get]
func (h CouponHttp) GetUsageHistoryBySaleInvoice(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	saleInvoiceID := ctx.QueryParam("sale_invoice_id")

	if saleInvoiceID == "" {
		ctx.ResponseError(400, "sale_invoice_id is required")
		return nil
	}

	// ดึงประวัติการใช้
	items, err := h.svc.GetUsageHistoryBySaleInvoice(holdingCode, saleInvoiceID)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data: map[string]interface{}{
			"usage_history": items,
			"total":         len(items),
		},
	})

	return nil
}

// @Summary		ดูประวัติการใช้คูปองตาม Transaction ID
// @Description ดูประวัติการใช้คูปองของการทำรายการใดรายการหนึ่ง
// @Tags		Coupon
// @Param		transaction_id query string true "Transaction ID"
// @Accept 		json
// @Produce 	json
// @Success 	200 {object} common.ApiResponse
// @Failure 	401 {object} common.AuthResponseFailed
// @Security 	AccessToken
// @Router 		/coupon/transaction/usage-history [get]
func (h CouponHttp) GetUsageHistoryByTransactionID(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	transactionID := ctx.QueryParam("transaction_id")

	if transactionID == "" {
		ctx.ResponseError(400, "transaction_id is required")
		return nil
	}

	// ดึงประวัติการใช้
	items, err := h.svc.GetUsageHistoryByTransactionID(holdingCode, transactionID)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data: map[string]interface{}{
			"usage_history": items,
			"total":         len(items),
		},
	})

	return nil
}
