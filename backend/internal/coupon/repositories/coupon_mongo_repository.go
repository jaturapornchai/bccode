package repositories

import (
	"context"
	"smlcloudplatform/internal/coupon/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ICouponRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(category models.CouponDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.CouponDoc) error
	Update(ctx context.Context, holdingCode string, guid string, category models.CouponDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Find(holdingCode string, searchInFields []string, q string) ([]models.CouponInfo, error)
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.CouponInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.CouponDoc, error)

	// Coupon reservation methods
	UpdateCouponUsage(ctx context.Context, holdingCode string, couponID string, remainingValue float64, isOnceOnly bool) error
}

type ICouponReservationRepository interface {
	CreateReservation(ctx context.Context, holdingCode string, reservation models.CouponReservationDoc) (string, error)
	FindReservationByID(ctx context.Context, holdingCode string, reservationID string) (models.CouponReservationDoc, error)
	FindReservationByTransactionID(ctx context.Context, holdingCode string, transactionID string) (models.CouponReservationDoc, error)
	FindActiveReservationsByCoupon(ctx context.Context, holdingCode string, couponID string) ([]models.CouponReservationDoc, error)
	FindActiveReservationsByCustomerAndCoupon(ctx context.Context, holdingCode string, customerID string, couponID string) ([]models.CouponReservationDoc, error)
	CountActiveReservations(ctx context.Context, holdingCode string, couponID string) (int, error)
	CountActiveReservationsByCustomer(ctx context.Context, holdingCode string, couponID string, customerID string) (int, error)
	UpdateReservationStatus(ctx context.Context, holdingCode string, reservationID string, status models.ReservationStatus) error
	CancelReservation(ctx context.Context, holdingCode string, reservationID string, customerID string) error
	FindExpiredReservations(ctx context.Context, holdingCode string) ([]models.CouponReservationDoc, error)
	CleanupExpiredReservations(ctx context.Context, holdingCode string) error
}

type ICouponUsageHistoryRepository interface {
	CreateUsageHistory(ctx context.Context, holdingCode string, history models.CouponUsageHistoryDoc) (string, error)
	FindUsageHistoryByID(ctx context.Context, holdingCode string, historyID string) (models.CouponUsageHistoryDoc, error)
	FindUsageHistoryByCoupon(ctx context.Context, holdingCode string, couponID string, page, pageSize int) ([]models.CouponUsageHistoryDoc, int, error)
	FindUsageHistoryByCustomer(ctx context.Context, holdingCode string, customerID string, page, pageSize int) ([]models.CouponUsageHistoryDoc, int, error)
	FindUsageHistoryBySaleInvoice(ctx context.Context, holdingCode string, saleInvoiceID string) ([]models.CouponUsageHistoryDoc, error)
	FindUsageHistoryByTransactionID(ctx context.Context, holdingCode string, transactionID string) ([]models.CouponUsageHistoryDoc, error)
	SearchUsageHistory(ctx context.Context, holdingCode string, req models.CouponUsageHistoryRequest) ([]models.CouponUsageHistoryDoc, int, error)
	GetUsageHistorySummary(ctx context.Context, holdingCode string, req models.CouponUsageHistoryRequest) (models.CouponUsageHistoryResponse, error)
}

type CouponRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.CouponDoc]
	repositories.SearchRepository[models.CouponInfo]
	repositories.GuidRepository[models.CouponItemGuid]
}

type CouponReservationRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.CouponReservationDoc]
}

type CouponUsageHistoryRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.CouponUsageHistoryDoc]
}

func NewCouponRepository(pst microservice.IPersisterMongo) CouponRepository {

	insRepo := CouponRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.CouponDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.CouponInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.CouponItemGuid](pst)

	return insRepo
}

func NewCouponReservationRepository(pst microservice.IPersisterMongo) CouponReservationRepository {
	insRepo := CouponReservationRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.CouponReservationDoc](pst)
	return insRepo
}

func NewCouponUsageHistoryRepository(pst microservice.IPersisterMongo) CouponUsageHistoryRepository {
	insRepo := CouponUsageHistoryRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.CouponUsageHistoryDoc](pst)
	return insRepo
}

// UpdateCouponUsage อัพเดทการใช้งานคูปอง
func (repo CouponRepository) UpdateCouponUsage(ctx context.Context, holdingCode string, couponID string, remainingValue float64, isOnceOnly bool) error {
	filter := bson.M{
		"holdingcode": holdingCode,
		"guidfixed":   couponID,
	}

	updateFields := bson.M{
		"remainingvalue": remainingValue,
		"updatedat":      time.Now().UTC(), // Use UTC time consistently
	}

	// ถ้าเป็นคูปองใช้ครั้งเดียวและมูลค่าเหลือ 0 ให้ยกเลิกคูปอง
	if isOnceOnly && remainingValue <= 0 {
		updateFields["status"] = models.CouponStatusCanceled
	}

	update := bson.M{
		"$set": updateFields,
	}

	return repo.pst.Update(ctx, &models.CouponDoc{}, filter, update)
}

// CreateReservation สร้างการจองคูปอง
func (repo CouponReservationRepository) CreateReservation(ctx context.Context, holdingCode string, reservation models.CouponReservationDoc) (string, error) {
	reservation.ID = primitive.NewObjectID()
	reservation.HoldingCode = holdingCode
	reservation.CreatedAt = time.Now().UTC() // Use UTC time consistently
	reservation.UpdatedAt = time.Now().UTC() // Use UTC time consistently

	idx, err := repo.pst.Create(ctx, &models.CouponReservationDoc{}, reservation)
	if err != nil {
		return "", err
	}

	return idx.Hex(), nil
}

// FindReservationByID ค้นหาการจองด้วย ID
func (repo CouponReservationRepository) FindReservationByID(ctx context.Context, holdingCode string, reservationID string) (models.CouponReservationDoc, error) {
	var reservation models.CouponReservationDoc

	objID, err := primitive.ObjectIDFromHex(reservationID)
	if err != nil {
		return reservation, err
	}

	filter := bson.M{
		"_id":         objID,
		"holdingcode": holdingCode,
	}

	err = repo.pst.FindOne(ctx, &models.CouponReservationDoc{}, filter, &reservation)
	return reservation, err
}

// FindReservationByTransactionID ค้นหาการจองด้วย Transaction ID
func (repo CouponReservationRepository) FindReservationByTransactionID(ctx context.Context, holdingCode string, transactionID string) (models.CouponReservationDoc, error) {
	var reservation models.CouponReservationDoc

	filter := bson.M{
		"transactionid": transactionID,
		"holdingcode":   holdingCode,
	}

	err := repo.pst.FindOne(ctx, &models.CouponReservationDoc{}, filter, &reservation)
	return reservation, err
}

// FindActiveReservationsByCoupon ค้นหาการจองที่ยังใช้งานได้ของคูปอง
func (repo CouponReservationRepository) FindActiveReservationsByCoupon(ctx context.Context, holdingCode string, couponID string) ([]models.CouponReservationDoc, error) {
	var reservations []models.CouponReservationDoc

	filter := bson.M{
		"holdingcode": holdingCode,
		"couponid":    couponID,
		"status":      models.ReservationStatusActive,
		"expiresat":   bson.M{"$gt": time.Now().UTC()}, // Use UTC time for MongoDB queries
	}

	err := repo.pst.Find(ctx, &models.CouponReservationDoc{}, filter, &reservations)
	return reservations, err
}

// FindActiveReservationsByCustomerAndCoupon ค้นหาการจองที่ยังใช้งานได้ของลูกค้าและคูปองเฉพาะ
func (repo CouponReservationRepository) FindActiveReservationsByCustomerAndCoupon(ctx context.Context, holdingCode string, customerID string, couponID string) ([]models.CouponReservationDoc, error) {
	var reservations []models.CouponReservationDoc

	filter := bson.M{
		"holdingcode": holdingCode,
		"customerid":  customerID,
		"couponid":    couponID,
		"status":      models.ReservationStatusActive,
		"expiresat":   bson.M{"$gt": time.Now().UTC()}, // Use UTC time for MongoDB queries
	}

	err := repo.pst.Find(ctx, &models.CouponReservationDoc{}, filter, &reservations)
	return reservations, err
}

// CountActiveReservations นับจำนวนการจองที่ยังใช้งานได้ของคูปอง
func (repo CouponReservationRepository) CountActiveReservations(ctx context.Context, holdingCode string, couponID string) (int, error) {
	filter := bson.M{
		"holdingcode": holdingCode,
		"couponid":    couponID,
		"status":      models.ReservationStatusActive,
		"expiresat":   bson.M{"$gt": time.Now().UTC()}, // Use UTC time for MongoDB queries
	}

	return repo.pst.Count(ctx, &models.CouponReservationDoc{}, filter)
}

// CountActiveReservationsByCustomer นับจำนวนการจองที่ยังใช้งานได้ของลูกค้าและคูปองเฉพาะ
func (repo CouponReservationRepository) CountActiveReservationsByCustomer(ctx context.Context, holdingCode string, couponID string, customerID string) (int, error) {
	filter := bson.M{
		"holdingcode": holdingCode,
		"customerid":  customerID,
		"couponid":    couponID,
		"status":      models.ReservationStatusActive,
		"expiresat":   bson.M{"$gt": time.Now().UTC()}, // Use UTC time for MongoDB queries
	}

	return repo.pst.Count(ctx, &models.CouponReservationDoc{}, filter)
}

// UpdateReservationStatus อัพเดทสถานะการจอง
func (repo CouponReservationRepository) UpdateReservationStatus(ctx context.Context, holdingCode string, reservationID string, status models.ReservationStatus) error {
	objID, err := primitive.ObjectIDFromHex(reservationID)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id":         objID,
		"holdingcode": holdingCode,
	}

	updateFields := bson.M{
		"status":    status,
		"updatedat": time.Now().UTC(), // Use UTC time consistently
	}

	// เพิ่มเวลาที่ใช้งานหรือยกเลิก
	if status == models.ReservationStatusUsed {
		updateFields["usedat"] = time.Now().UTC() // Use UTC time consistently
	} else if status == models.ReservationStatusCanceled {
		updateFields["canceledat"] = time.Now().UTC() // Use UTC time consistently
	}

	update := bson.M{
		"$set": updateFields,
	}

	return repo.pst.Update(ctx, &models.CouponReservationDoc{}, filter, update)
}

// CancelReservation ยกเลิกการจอง
func (repo CouponReservationRepository) CancelReservation(ctx context.Context, holdingCode string, reservationID string, customerID string) error {
	objID, err := primitive.ObjectIDFromHex(reservationID)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id":         objID,
		"holdingcode": holdingCode,
		"customerid":  customerID,
		"status":      models.ReservationStatusActive,
	}

	updateFields := bson.M{
		"status":     models.ReservationStatusCanceled,
		"canceledat": time.Now().UTC(), // Use UTC time consistently
		"updatedat":  time.Now().UTC(), // Use UTC time consistently
	}

	update := bson.M{
		"$set": updateFields,
	}

	return repo.pst.Update(ctx, &models.CouponReservationDoc{}, filter, update)
}

// FindExpiredReservations ค้นหาการจองที่หมดอายุ
func (repo CouponReservationRepository) FindExpiredReservations(ctx context.Context, holdingCode string) ([]models.CouponReservationDoc, error) {
	var reservations []models.CouponReservationDoc

	filter := bson.M{
		"status":    models.ReservationStatusActive,
		"expiresat": bson.M{"$lt": time.Now().UTC()}, // Use UTC time for MongoDB queries
	}

	// ถ้า holdingCode ไม่ใช่ empty string ให้เพิ่มเงื่อนไข holdingcode
	if holdingCode != "" {
		filter["holdingcode"] = holdingCode
	}

	err := repo.pst.Find(ctx, &models.CouponReservationDoc{}, filter, &reservations)
	return reservations, err
}

// CleanupExpiredReservations ล้างการจองที่หมดอายุ
func (repo CouponReservationRepository) CleanupExpiredReservations(ctx context.Context, holdingCode string) error {
	filter := bson.M{
		"status":    models.ReservationStatusActive,
		"expiresat": bson.M{"$lt": time.Now().UTC()}, // Use UTC time for MongoDB queries
	}

	// ถ้า holdingCode ไม่ใช่ empty string ให้เพิ่มเงื่อนไข holdingcode
	if holdingCode != "" {
		filter["holdingcode"] = holdingCode
	}

	updateFields := bson.M{
		"status":    models.ReservationStatusExpired,
		"updatedat": time.Now().UTC(), // Use UTC time consistently
	}

	update := bson.M{
		"$set": updateFields,
	}

	err := repo.pst.Update(ctx, &models.CouponReservationDoc{}, filter, update)
	return err
}

// CouponUsageHistoryRepository methods
func (repo CouponUsageHistoryRepository) CreateUsageHistory(ctx context.Context, holdingCode string, history models.CouponUsageHistoryDoc) (string, error) {
	history.HoldingCode = holdingCode
	history.CreatedAt = time.Now().UTC()
	history.UpdatedAt = time.Now().UTC()

	insertedID, err := repo.pst.Create(ctx, &models.CouponUsageHistoryDoc{}, history)
	if err != nil {
		return "", err
	}

	return insertedID.Hex(), nil
}

func (repo CouponUsageHistoryRepository) FindUsageHistoryByID(ctx context.Context, holdingCode string, historyID string) (models.CouponUsageHistoryDoc, error) {
	var history models.CouponUsageHistoryDoc

	objID, err := primitive.ObjectIDFromHex(historyID)
	if err != nil {
		return history, err
	}

	filter := bson.M{
		"_id":         objID,
		"holdingcode": holdingCode,
	}

	err = repo.pst.FindOne(ctx, &models.CouponUsageHistoryDoc{}, filter, &history)
	return history, err
}

func (repo CouponUsageHistoryRepository) FindUsageHistoryByCoupon(ctx context.Context, holdingCode string, couponID string, page, pageSize int) ([]models.CouponUsageHistoryDoc, int, error) {
	var histories []models.CouponUsageHistoryDoc

	filter := bson.M{
		"couponid":    couponID,
		"holdingcode": holdingCode,
	}

	skip := (page - 1) * pageSize
	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(pageSize)).SetSort(bson.M{"usedat": -1})

	err := repo.pst.Find(ctx, &models.CouponUsageHistoryDoc{}, filter, &histories, opts)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	total, err := repo.pst.Count(ctx, &models.CouponUsageHistoryDoc{}, filter)
	if err != nil {
		return histories, 0, err
	}

	return histories, int(total), nil
}

func (repo CouponUsageHistoryRepository) FindUsageHistoryByCustomer(ctx context.Context, holdingCode string, customerID string, page, pageSize int) ([]models.CouponUsageHistoryDoc, int, error) {
	var histories []models.CouponUsageHistoryDoc

	filter := bson.M{
		"customerid":  customerID,
		"holdingcode": holdingCode,
	}

	skip := (page - 1) * pageSize
	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(pageSize)).SetSort(bson.M{"usedat": -1})

	err := repo.pst.Find(ctx, &models.CouponUsageHistoryDoc{}, filter, &histories, opts)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	total, err := repo.pst.Count(ctx, &models.CouponUsageHistoryDoc{}, filter)
	if err != nil {
		return histories, 0, err
	}

	return histories, int(total), nil
}

func (repo CouponUsageHistoryRepository) FindUsageHistoryBySaleInvoice(ctx context.Context, holdingCode string, saleInvoiceID string) ([]models.CouponUsageHistoryDoc, error) {
	var histories []models.CouponUsageHistoryDoc

	filter := bson.M{
		"sale_invoice_id": saleInvoiceID,
		"holdingcode":     holdingCode,
	}

	opts := options.Find().SetSort(bson.M{"usedat": -1})
	err := repo.pst.Find(ctx, &models.CouponUsageHistoryDoc{}, filter, &histories, opts)

	return histories, err
}

func (repo CouponUsageHistoryRepository) FindUsageHistoryByTransactionID(ctx context.Context, holdingCode string, transactionID string) ([]models.CouponUsageHistoryDoc, error) {
	var histories []models.CouponUsageHistoryDoc

	filter := bson.M{
		"transactionid": transactionID,
		"holdingcode":   holdingCode,
	}

	opts := options.Find().SetSort(bson.M{"usedat": -1})
	err := repo.pst.Find(ctx, &models.CouponUsageHistoryDoc{}, filter, &histories, opts)

	return histories, err
}

func (repo CouponUsageHistoryRepository) SearchUsageHistory(ctx context.Context, holdingCode string, req models.CouponUsageHistoryRequest) ([]models.CouponUsageHistoryDoc, int, error) {
	var histories []models.CouponUsageHistoryDoc

	// Build filter
	filter := bson.M{"holdingcode": holdingCode}

	if req.CouponID != "" {
		filter["couponid"] = req.CouponID
	}
	if req.CouponCode != "" {
		filter["couponcode"] = bson.M{"$regex": req.CouponCode, "$options": "i"}
	}
	if req.CustomerID != "" {
		filter["customerid"] = req.CustomerID
	}
	if req.SaleInvoiceID != "" {
		filter["saleinvoiceid"] = req.SaleInvoiceID
	}
	if req.SaleInvoiceNumber != "" {
		filter["saleinvoicenumber"] = bson.M{"$regex": req.SaleInvoiceNumber, "$options": "i"}
	}
	if req.TransactionID != "" {
		filter["transactionid"] = req.TransactionID
	}

	// Date range filter
	if req.StartDate != nil || req.EndDate != nil {
		dateFilter := bson.M{}
		if req.StartDate != nil {
			dateFilter["$gte"] = req.StartDate.UTC()
		}
		if req.EndDate != nil {
			dateFilter["$lte"] = req.EndDate.UTC()
		}
		filter["usedat"] = dateFilter
	}

	// Pagination
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	skip := (page - 1) * pageSize
	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(pageSize)).SetSort(bson.M{"usedat": -1})

	err := repo.pst.Find(ctx, &models.CouponUsageHistoryDoc{}, filter, &histories, opts)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	total, err := repo.pst.Count(ctx, &models.CouponUsageHistoryDoc{}, filter)
	if err != nil {
		return histories, 0, err
	}

	return histories, int(total), nil
}

func (repo CouponUsageHistoryRepository) GetUsageHistorySummary(ctx context.Context, holdingCode string, req models.CouponUsageHistoryRequest) (models.CouponUsageHistoryResponse, error) {
	histories, total, err := repo.SearchUsageHistory(ctx, holdingCode, req)
	if err != nil {
		return models.CouponUsageHistoryResponse{}, err
	}

	// Convert to response items
	var items []models.CouponUsageHistoryItem
	var totalDiscount, totalCashVoucher, totalOrderAmount float64

	for _, history := range histories {
		item := models.CouponUsageHistoryItem{
			ID:                history.ID.Hex(),
			CouponID:          history.CouponUsageHistory.CouponID,
			CouponCode:        history.CouponUsageHistory.CouponCode,
			CouponType:        history.CouponUsageHistory.CouponType,
			CouponTypeName:    history.CouponUsageHistory.CouponType.String(),
			CouponValue:       history.CouponUsageHistory.CouponValue,
			CustomerID:        history.CouponUsageHistory.CustomerID,
			CustomerCode:      history.CouponUsageHistory.CustomerCode,
			CustomerName:      history.CouponUsageHistory.CustomerName,
			SaleInvoiceID:     history.CouponUsageHistory.SaleInvoiceID,
			SaleInvoiceNumber: history.CouponUsageHistory.SaleInvoiceNumber,
			TransactionID:     history.CouponUsageHistory.TransactionID,
			ReservationID:     history.CouponUsageHistory.ReservationID,
			UsedAmount:        history.CouponUsageHistory.UsedAmount,
			DiscountAmount:    history.CouponUsageHistory.DiscountAmount,
			CashVoucherAmount: history.CouponUsageHistory.CashVoucherAmount,
			OrderAmount:       history.CouponUsageHistory.OrderAmount,
			UsedAt:            history.CouponUsageHistory.UsedAt,
			UsedBy:            history.CouponUsageHistory.UsedBy,
			Remark:            history.CouponUsageHistory.Remark,
		}
		items = append(items, item)

		totalDiscount += history.CouponUsageHistory.DiscountAmount
		totalCashVoucher += history.CouponUsageHistory.CashVoucherAmount
		totalOrderAmount += history.CouponUsageHistory.OrderAmount
	}

	// Calculate pagination
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	totalPage := (total + pageSize - 1) / pageSize

	response := models.CouponUsageHistoryResponse{
		UsageHistory: items,
		Pagination: struct {
			Page      int `json:"page"`
			PageSize  int `json:"pagesize"`
			Total     int `json:"total"`
			TotalPage int `json:"totalpage"`
		}{
			Page:      page,
			PageSize:  pageSize,
			Total:     total,
			TotalPage: totalPage,
		},
		Summary: struct {
			TotalUsed        int     `json:"totalused"`
			TotalDiscount    float64 `json:"totaldiscount"`
			TotalCashVoucher float64 `json:"totalcashvoucher"`
			TotalOrderAmount float64 `json:"totalorderamount"`
		}{
			TotalUsed:        len(items),
			TotalDiscount:    totalDiscount,
			TotalCashVoucher: totalCashVoucher,
			TotalOrderAmount: totalOrderAmount,
		},
	}

	return response, nil
}
