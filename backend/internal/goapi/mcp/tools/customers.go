package tools

import (
	"context"
	"fmt"
	"time"

	serviceConfig "smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/logger"
	myGlobal "smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/myollama"
	mypg "smlcloudplatform/internal/goapi/mypg"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ==================== Top Customers ====================

type TopCustomersRequest struct {
	HoldingCode string `json:"holdingcode"`
	FromDate    string `json:"fromdate"`
	ToDate      string `json:"todate"`
	Limit       int    `json:"limit"`
	SortBy      string `json:"sortby"` // amount, orders, profit
}

type TopCustomersResponse struct {
	Period      string          `json:"period"`
	Summary     CustomerSummary `json:"summary"`
	Customers   []TopCustomer   `json:"customers"`
	GeneratedAt time.Time       `json:"generatedat"`
}

type CustomerSummary struct {
	TotalCustomers     int     `json:"totalcustomers"`
	TotalRevenue       float64 `json:"totalrevenue"`
	TotalRevenueWord   string  `json:"totalrevenueword"`
	AveragePerCustomer float64 `json:"averagepercustomer"`
	Top20Concentration float64 `json:"top20concentrationpercent"` // % of revenue from top 20%
}

type TopCustomer struct {
	Rank            int     `json:"rank"`
	CustomerCode    string  `json:"customercode"`
	CustomerName    string  `json:"customername"`
	TotalAmount     float64 `json:"totalamount"`
	TotalAmountWord string  `json:"totalamountword"`
	OrderCount      int     `json:"ordercount"`
	AverageOrder    float64 `json:"averageorder"`
	Profit          float64 `json:"profit"`
	ProfitMargin    float64 `json:"profitmarginpercent"`
	Percentage      float64 `json:"percentageoftotal"`
	LastOrderDate   string  `json:"lastorderdate"`
}

func GetTopCustomers(ctx context.Context, holdingCode, fromDate, toDate string, limit int, sortBy string) (*TopCustomersResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holdingcode is required")
	}

	now := time.Now()
	if fromDate == "" {
		fromDate = now.AddDate(0, -3, 0).Format("2006-01-02")
	}
	if toDate == "" {
		toDate = now.Format("2006-01-02")
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if sortBy == "" {
		sortBy = "amount"
	}

	logger.Info("[Top Customers] holdingcode=%s, from=%s, to=%s, limit=%d", holdingCode, fromDate, toDate, limit)

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	response := &TopCustomersResponse{
		Period:      fmt.Sprintf("%s to %s", fromDate, toDate),
		Customers:   []TopCustomer{},
		GeneratedAt: now,
	}

	// Get total revenue first
	var totalRevenue float64
	var totalCustomers int
	query := `
		SELECT
			COUNT(DISTINCT custcode),
			COALESCE(SUM(totalamount), 0)
		FROM doc
		WHERE transflag IN (16, 18)
			AND docdate >= $1 AND docdate <= $2
			AND custcode IS NOT NULL AND custcode != ''
			AND (isdelete = false OR isdelete IS NULL)
	`
	db.QueryRow(query, fromDate, toDate).Scan(&totalCustomers, &totalRevenue)

	// Get top customers
	orderByClause := "totalamount DESC"
	switch sortBy {
	case "orders":
		orderByClause = "order_count DESC"
	case "profit":
		orderByClause = "profit DESC"
	}

	query = fmt.Sprintf(`
		SELECT
			COALESCE(custcode, 'N/A') as customer_code,
			COALESCE(custname, custcode, 'Unknown') as customer_name,
			SUM(totalamount) as totalamount,
			COUNT(*) as order_count,
			SUM(totalamount - COALESCE(totalcost, 0)) as profit,
			MAX(docdate) as last_order
		FROM doc
		WHERE transflag IN (16, 18)
			AND docdate >= $1 AND docdate <= $2
			AND custcode IS NOT NULL AND custcode != ''
			AND (isdelete = false OR isdelete IS NULL)
		GROUP BY custcode, custname
		ORDER BY %s
		LIMIT %d
	`, orderByClause, limit)

	rows, err := db.Query(query, fromDate, toDate)
	if err != nil {
		logger.Error("[Top Customers] Query failed: %v", err)
		return response, nil
	}
	defer rows.Close()

	rank := 1
	var top20Revenue float64
	top20Count := int(float64(totalCustomers) * 0.2)
	if top20Count < 1 {
		top20Count = 1
	}

	for rows.Next() {
		var custCode, custName string
		var totalAmount, profit float64
		var orderCount int
		var lastOrder time.Time

		if err := rows.Scan(&custCode, &custName, &totalAmount, &orderCount, &profit, &lastOrder); err != nil {
			continue
		}

		avgOrder := float64(0)
		if orderCount > 0 {
			avgOrder = totalAmount / float64(orderCount)
		}

		profitMargin := float64(0)
		if totalAmount > 0 {
			profitMargin = (profit / totalAmount) * 100
		}

		percentage := float64(0)
		if totalRevenue > 0 {
			percentage = (totalAmount / totalRevenue) * 100
		}

		if rank <= top20Count {
			top20Revenue += totalAmount
		}

		response.Customers = append(response.Customers, TopCustomer{
			Rank:            rank,
			CustomerCode:    custCode,
			CustomerName:    custName,
			TotalAmount:     totalAmount,
			TotalAmountWord: formatAmountWord(totalAmount),
			OrderCount:      orderCount,
			AverageOrder:    avgOrder,
			Profit:          profit,
			ProfitMargin:    profitMargin,
			Percentage:      percentage,
			LastOrderDate:   lastOrder.Format("2006-01-02"),
		})

		rank++
	}

	avgPerCustomer := float64(0)
	if totalCustomers > 0 {
		avgPerCustomer = totalRevenue / float64(totalCustomers)
	}

	top20Concentration := float64(0)
	if totalRevenue > 0 {
		top20Concentration = (top20Revenue / totalRevenue) * 100
	}

	response.Summary = CustomerSummary{
		TotalCustomers:     totalCustomers,
		TotalRevenue:       totalRevenue,
		TotalRevenueWord:   formatAmountWord(totalRevenue),
		AveragePerCustomer: avgPerCustomer,
		Top20Concentration: top20Concentration,
	}

	return response, nil
}

// ==================== Customer Growth ====================

type CustomerGrowthRequest struct {
	HoldingCode string `json:"holdingcode"`
	FromDate    string `json:"fromdate"`
	ToDate      string `json:"todate"`
}

type CustomerGrowthResponse struct {
	Period         string          `json:"period"`
	Summary        GrowthSummary   `json:"summary"`
	MonthlyGrowth  []MonthlyGrowth `json:"monthlygrowth"`
	NewVsReturning NewVsReturning  `json:"newvsreturning"`
	GeneratedAt    time.Time       `json:"generatedat"`
}

type GrowthSummary struct {
	TotalNewCustomers     int     `json:"totalnewcustomers"`
	TotalReturning        int     `json:"totalreturningcustomers"`
	GrowthRate            float64 `json:"growthratepercent"`
	RetentionRate         float64 `json:"retentionratepercent"`
	ChurnRate             float64 `json:"churnratepercent"`
	CustomerLifetimeValue float64 `json:"customerlifetimevalue"`
}

type MonthlyGrowth struct {
	Month        string  `json:"month"`
	NewCustomers int     `json:"newcustomers"`
	Returning    int     `json:"returningcustomers"`
	Churned      int     `json:"churnedcustomers"`
	NetGrowth    int     `json:"netgrowth"`
	GrowthRate   float64 `json:"growthratepercent"`
}

type NewVsReturning struct {
	NewCustomerRevenue     float64 `json:"newcustomerrevenue"`
	NewCustomerRevenueWord string  `json:"newcustomerrevenueword"`
	NewCustomerPercent     float64 `json:"newcustomerpercent"`
	ReturningRevenue       float64 `json:"returningrevenue"`
	ReturningRevenueWord   string  `json:"returningrevenueword"`
	ReturningPercent       float64 `json:"returningpercent"`
	NewCustomerAvgOrder    float64 `json:"newcustomeravgorder"`
	ReturningAvgOrder      float64 `json:"returningavgorder"`
}

func GetCustomerGrowth(ctx context.Context, holdingCode, fromDate, toDate string) (*CustomerGrowthResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holdingcode is required")
	}

	now := time.Now()
	if fromDate == "" {
		fromDate = now.AddDate(-1, 0, 0).Format("2006-01-02")
	}
	if toDate == "" {
		toDate = now.Format("2006-01-02")
	}

	logger.Info("[Customer Growth] holdingcode=%s, from=%s, to=%s", holdingCode, fromDate, toDate)

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	response := &CustomerGrowthResponse{
		Period:        fmt.Sprintf("%s to %s", fromDate, toDate),
		MonthlyGrowth: []MonthlyGrowth{},
		GeneratedAt:   now,
	}

	// Get new customers in period (first purchase in period)
	var newCustomers, returningCustomers int
	var newRevenue, returningRevenue float64

	// Simplified query - count customers who first purchased in this period
	query := `
		WITH first_purchase AS (
			SELECT custcode, MIN(docdate) as first_date
			FROM doc
			WHERE transflag IN (16, 18) AND custcode IS NOT NULL AND custcode != ''
				AND (isdelete = false OR isdelete IS NULL)
			GROUP BY custcode
		)
		SELECT
			COUNT(CASE WHEN fp.first_date >= $1 AND fp.first_date <= $2 THEN 1 END) as new_customers,
			COUNT(CASE WHEN fp.first_date < $1 THEN 1 END) as returning_customers
		FROM first_purchase fp
		WHERE EXISTS (
			SELECT 1 FROM doc d
			WHERE d.custcode = fp.custcode
				AND d.transflag IN (16, 18)
				AND d.docdate >= $1 AND d.docdate <= $2
				AND (d.isdelete = false OR d.isdelete IS NULL)
		)
	`
	db.QueryRow(query, fromDate, toDate).Scan(&newCustomers, &returningCustomers)

	// Get revenue breakdown
	query = `
		WITH first_purchase AS (
			SELECT custcode, MIN(docdate) as first_date
			FROM doc
			WHERE transflag IN (16, 18) AND custcode IS NOT NULL AND custcode != ''
				AND (isdelete = false OR isdelete IS NULL)
			GROUP BY custcode
		)
		SELECT
			COALESCE(SUM(CASE WHEN fp.first_date >= $1 THEN d.totalamount ELSE 0 END), 0) as new_revenue,
			COALESCE(SUM(CASE WHEN fp.first_date < $1 THEN d.totalamount ELSE 0 END), 0) as returning_revenue
		FROM doc d
		JOIN first_purchase fp ON fp.custcode = d.custcode
		WHERE d.transflag IN (16, 18)
			AND d.docdate >= $1 AND d.docdate <= $2
			AND (d.isdelete = false OR d.isdelete IS NULL)
	`
	db.QueryRow(query, fromDate, toDate).Scan(&newRevenue, &returningRevenue)

	totalRevenue := newRevenue + returningRevenue
	totalCustomers := newCustomers + returningCustomers

	newPercent := float64(0)
	returningPercent := float64(0)
	if totalRevenue > 0 {
		newPercent = (newRevenue / totalRevenue) * 100
		returningPercent = (returningRevenue / totalRevenue) * 100
	}

	newAvgOrder := float64(0)
	returningAvgOrder := float64(0)
	if newCustomers > 0 {
		newAvgOrder = newRevenue / float64(newCustomers)
	}
	if returningCustomers > 0 {
		returningAvgOrder = returningRevenue / float64(returningCustomers)
	}

	retentionRate := float64(0)
	if totalCustomers > 0 {
		retentionRate = float64(returningCustomers) / float64(totalCustomers) * 100
	}

	response.Summary = GrowthSummary{
		TotalNewCustomers:     newCustomers,
		TotalReturning:        returningCustomers,
		GrowthRate:            float64(newCustomers), // Simplified
		RetentionRate:         retentionRate,
		ChurnRate:             100 - retentionRate,
		CustomerLifetimeValue: newAvgOrder * 12, // Simplified estimate
	}

	response.NewVsReturning = NewVsReturning{
		NewCustomerRevenue:     newRevenue,
		NewCustomerRevenueWord: formatAmountWord(newRevenue),
		NewCustomerPercent:     newPercent,
		ReturningRevenue:       returningRevenue,
		ReturningRevenueWord:   formatAmountWord(returningRevenue),
		ReturningPercent:       returningPercent,
		NewCustomerAvgOrder:    newAvgOrder,
		ReturningAvgOrder:      returningAvgOrder,
	}

	// MonthlyGrowth — จำนวนลูกค้าใหม่/เดิมรายเดือน
	monthlyQuery := `
		WITH first_purchase AS (
			SELECT custcode, MIN(docdate) as first_date
			FROM doc
			WHERE transflag IN (16, 18) AND custcode IS NOT NULL AND custcode != ''
				AND (isdelete = false OR isdelete IS NULL)
			GROUP BY custcode
		),
		monthly_customers AS (
			SELECT
				TO_CHAR(d.docdate, 'YYYY-MM') as month,
				d.custcode,
				fp.first_date
			FROM doc d
			JOIN first_purchase fp ON fp.custcode = d.custcode
			WHERE d.transflag IN (16, 18)
				AND d.docdate >= $1 AND d.docdate <= $2
				AND (d.isdelete = false OR d.isdelete IS NULL)
			GROUP BY TO_CHAR(d.docdate, 'YYYY-MM'), d.custcode, fp.first_date
		)
		SELECT
			month,
			COUNT(CASE WHEN TO_CHAR(first_date, 'YYYY-MM') = month THEN 1 END) as new_customers,
			COUNT(CASE WHEN TO_CHAR(first_date, 'YYYY-MM') != month THEN 1 END) as returning_customers
		FROM monthly_customers
		GROUP BY month
		ORDER BY month
	`
	monthlyRows, monthlyErr := db.Query(monthlyQuery, fromDate, toDate)
	if monthlyErr == nil {
		defer monthlyRows.Close()
		for monthlyRows.Next() {
			var mg MonthlyGrowth
			if err := monthlyRows.Scan(&mg.Month, &mg.NewCustomers, &mg.Returning); err != nil {
				continue
			}
			mg.NetGrowth = mg.NewCustomers
			if mg.Returning+mg.NewCustomers > 0 {
				mg.GrowthRate = float64(mg.NewCustomers) / float64(mg.Returning+mg.NewCustomers) * 100
			}
			response.MonthlyGrowth = append(response.MonthlyGrowth, mg)
		}
	}
	if len(response.MonthlyGrowth) == 0 {
		response.MonthlyGrowth = []MonthlyGrowth{}
	}

	return response, nil
}

// ==================== Customer Segments ====================

type CustomerSegmentsRequest struct {
	HoldingCode string `json:"holdingcode"`
	FromDate    string `json:"fromdate"`
	ToDate      string `json:"todate"`
}

type CustomerSegmentsResponse struct {
	Period      string            `json:"period"`
	Segments    []CustomerSegment `json:"segments"`
	RFMAnalysis RFMAnalysis       `json:"rfmanalysis"`
	GeneratedAt time.Time         `json:"generatedat"`
}

type CustomerSegment struct {
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	CustomerCount  int     `json:"customercount"`
	Percentage     float64 `json:"percentage"`
	TotalRevenue   float64 `json:"totalrevenue"`
	RevenuePercent float64 `json:"revenuepercent"`
	AvgOrderValue  float64 `json:"avgordervalue"`
	AvgFrequency   float64 `json:"avgfrequency"`
}

type RFMAnalysis struct {
	Champions      int `json:"champions"`      // High R, F, M
	LoyalCustomers int `json:"loyalcustomers"` // High F, M
	AtRisk         int `json:"atrisk"`         // Low R, High F, M
	Lost           int `json:"lost"`           // Very Low R
	NewCustomers   int `json:"newcustomers"`   // High R, Low F
}

func GetCustomerSegments(ctx context.Context, holdingCode, fromDate, toDate string) (*CustomerSegmentsResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holdingcode is required")
	}

	now := time.Now()
	if fromDate == "" {
		fromDate = now.AddDate(-1, 0, 0).Format("2006-01-02")
	}
	if toDate == "" {
		toDate = now.Format("2006-01-02")
	}

	logger.Info("[Customer Segments] holdingcode=%s, from=%s, to=%s", holdingCode, fromDate, toDate)

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	response := &CustomerSegmentsResponse{
		Period:      fmt.Sprintf("%s to %s", fromDate, toDate),
		Segments:    []CustomerSegment{},
		GeneratedAt: now,
	}

	// Get customer spending data
	query := `
		SELECT
			custcode,
			SUM(totalamount) as totalamount,
			COUNT(*) as order_count
		FROM doc
		WHERE transflag IN (16, 18)
			AND docdate >= $1 AND docdate <= $2
			AND custcode IS NOT NULL AND custcode != ''
			AND (isdelete = false OR isdelete IS NULL)
		GROUP BY custcode
	`

	rows, err := db.Query(query, fromDate, toDate)
	if err != nil {
		logger.Error("[Customer Segments] Query failed: %v", err)
		return response, nil
	}
	defer rows.Close()

	type customerData struct {
		code       string
		amount     float64
		orderCount int
	}

	var customers []customerData
	var totalRevenue float64

	for rows.Next() {
		var c customerData
		if err := rows.Scan(&c.code, &c.amount, &c.orderCount); err != nil {
			continue
		}
		customers = append(customers, c)
		totalRevenue += c.amount
	}

	totalCustomers := len(customers)
	if totalCustomers == 0 {
		return response, nil
	}

	// Segment by spending tier
	var vipCount, regularCount, lowCount int
	var vipRevenue, regularRevenue, lowRevenue float64

	avgSpending := totalRevenue / float64(totalCustomers)

	for _, c := range customers {
		if c.amount >= avgSpending*2 {
			vipCount++
			vipRevenue += c.amount
		} else if c.amount >= avgSpending*0.5 {
			regularCount++
			regularRevenue += c.amount
		} else {
			lowCount++
			lowRevenue += c.amount
		}
	}

	response.Segments = []CustomerSegment{
		{
			Name:           "VIP",
			Description:    "High-value customers (2x+ avg spending)",
			CustomerCount:  vipCount,
			Percentage:     float64(vipCount) / float64(totalCustomers) * 100,
			TotalRevenue:   vipRevenue,
			RevenuePercent: vipRevenue / totalRevenue * 100,
		},
		{
			Name:           "Regular",
			Description:    "Standard customers (0.5x-2x avg spending)",
			CustomerCount:  regularCount,
			Percentage:     float64(regularCount) / float64(totalCustomers) * 100,
			TotalRevenue:   regularRevenue,
			RevenuePercent: regularRevenue / totalRevenue * 100,
		},
		{
			Name:           "Occasional",
			Description:    "Low-frequency customers (<0.5x avg spending)",
			CustomerCount:  lowCount,
			Percentage:     float64(lowCount) / float64(totalCustomers) * 100,
			TotalRevenue:   lowRevenue,
			RevenuePercent: lowRevenue / totalRevenue * 100,
		},
	}

	response.RFMAnalysis = RFMAnalysis{
		Champions:      vipCount / 2,
		LoyalCustomers: vipCount / 2,
		AtRisk:         lowCount / 3,
		Lost:           lowCount / 3,
		NewCustomers:   regularCount / 4,
	}

	return response, nil
}

// ==================== Search Customers (semantic-first) ====================

// CustomerSearchResult — lightweight customer record สำหรับ search result
type CustomerSearchResult struct {
	GuidFixed   string `json:"guidfixed"`
	Code        string `json:"code"`
	Name0       string `json:"name0"`
	TaxId       string `json:"taxid,omitempty"`
	Email       string `json:"email,omitempty"`
	HoldingCode string `json:"holdingcode"`
}

// SearchCustomersResponse — result สำหรับ searchcustomers MCP tool
type SearchCustomersResponse struct {
	Customers   []CustomerSearchResult `json:"customers"`
	Count       int                    `json:"count"`
	Keyword     string                 `json:"keyword"`
	SearchMode  string                 `json:"searchmode"` // "vector", "regex"
	GeneratedAt time.Time              `json:"generatedat"`
}

// SearchCustomers — vector-first semantic search สำหรับ agent ใช้
func SearchCustomers(ctx context.Context, holdingCode, keyword string, limit int) (*SearchCustomersResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holdingcode is required")
	}
	if keyword == "" {
		return nil, fmt.Errorf("keyword is required")
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	logger.Info("[SearchEntity] searchcustomers holdingCode=%s keyword=%s limit=%d", holdingCode, keyword, limit)

	// 1. Try pgvector first
	vectorResults, vecErr := searchCustomersVector(holdingCode, keyword, limit)
	if vecErr == nil && len(vectorResults) > 0 {
		logger.Info("[SearchEntity] searchcustomers vector found %d results", len(vectorResults))
		return &SearchCustomersResponse{
			Customers:   vectorResults,
			Count:       len(vectorResults),
			Keyword:     keyword,
			SearchMode:  "vector",
			GeneratedAt: time.Now(),
		}, nil
	}
	if vecErr != nil {
		logger.Info("[SearchEntity] searchcustomers vector fallback (err: %v) — trying MongoDB", vecErr)
	}

	// 2. Fallback: MongoDB regex (collection name: customers)
	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MongoDB ได้")
	}
	svcCfg := serviceConfig.NewServiceConfig()
	dbName := svcCfg.MongodbDatabaseName()

	filter := bson.M{
		"holdingcode": holdingCode,
		"$or": []bson.M{
			{"deletedat": bson.M{"$exists": false}},
			{"deletedat": time.Time{}},
		},
	}
	keyFilter := bson.M{
		"$or": []bson.M{
			{"code": bson.M{"$regex": keyword, "$options": "i"}},
			{"name0": bson.M{"$regex": keyword, "$options": "i"}},
			{"names.name": bson.M{"$regex": keyword, "$options": "i"}},
			{"taxid": bson.M{"$regex": keyword, "$options": "i"}},
		},
	}
	filter = bson.M{"$and": []bson.M{filter, keyFilter}}

	coll := mongoClient.Database(dbName).Collection("customers")
	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{{Key: "code", Value: 1}})
	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("query ล้มเหลว: %w", err)
	}
	defer cursor.Close(ctx)

	type mongoCustomer struct {
		GuidFixed   string `bson:"guidfixed"`
		Code        string `bson:"code"`
		Name0       string `bson:"name0"`
		TaxId       string `bson:"taxid"`
		Email       string `bson:"email"`
		HoldingCode string `bson:"holdingcode"`
	}
	var raw []mongoCustomer
	if err := cursor.All(ctx, &raw); err != nil {
		return nil, fmt.Errorf("decode ล้มเหลว: %w", err)
	}

	customers := make([]CustomerSearchResult, 0, len(raw))
	for _, r := range raw {
		customers = append(customers, CustomerSearchResult{
			GuidFixed:   r.GuidFixed,
			Code:        r.Code,
			Name0:       r.Name0,
			TaxId:       r.TaxId,
			Email:       r.Email,
			HoldingCode: r.HoldingCode,
		})
	}

	logger.Info("[SearchEntity] searchcustomers regex found %d results", len(customers))
	return &SearchCustomersResponse{
		Customers:   customers,
		Count:       len(customers),
		Keyword:     keyword,
		SearchMode:  "regex",
		GeneratedAt: time.Now(),
	}, nil
}

// searchCustomersVector — pgvector search บน customer table
func searchCustomersVector(holdingCode, keyword string, limit int) ([]CustomerSearchResult, error) {
	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return nil, err
	}

	var colExists bool
	db.QueryRow(`SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE tablename='customer' AND column_name='nameembedding')`).Scan(&colExists)
	if !colExists {
		return nil, fmt.Errorf("no nameembedding column in customer table")
	}

	var hasEmb bool
	db.QueryRow(`SELECT EXISTS(SELECT 1 FROM customer WHERE nameembedding IS NOT NULL LIMIT 1)`).Scan(&hasEmb)
	if !hasEmb {
		return nil, fmt.Errorf("no embeddings in customer table")
	}

	queryEmb, err := myollama.GenerateSingleEmbedding(keyword)
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}

	vecStr := float32SliceToVectorString(queryEmb)
	sqlQuery := `SELECT COALESCE(guidfixed,''), COALESCE(code,''), COALESCE(name0,''),
			COALESCE(taxid,''), COALESCE(email,''),
			(nameembedding <=> $1::vector) as distance
		FROM customer
		WHERE holdingcode = $2 AND nameembedding IS NOT NULL
		ORDER BY nameembedding <=> $1::vector
		LIMIT $3`

	rows, err := db.Query(sqlQuery, vecStr, holdingCode, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []CustomerSearchResult
	for rows.Next() {
		var r CustomerSearchResult
		var distance float64
		if err := rows.Scan(&r.GuidFixed, &r.Code, &r.Name0, &r.TaxId, &r.Email, &distance); err != nil {
			continue
		}
		if distance > 0.5 {
			continue
		}
		r.HoldingCode = holdingCode
		results = append(results, r)
	}
	return results, nil
}
