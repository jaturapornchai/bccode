package repositories_test

import (
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/stockprocess/models"
	"smlcloudplatform/internal/stockprocess/repositories"
	"smlcloudplatform/pkg/microservice"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPushMovementToRedis(t *testing.T) {

	cfg := config.NewConfig()
	cachePersister := microservice.NewCacher(cfg.CacherConfig())
	err := cachePersister.Healthcheck()
	assert.Nil(t, err)
	stockProcessRedisRepo := repositories.NewStockProcessRedisRepository(cachePersister)

	stockProcessDocs := []models.StockData{
		{
			ShopID:    "SHOP1",
			DocNo:     "DOC1",
			Barcode:   "BAR1",
			CalcQty:   10,
			CalcFlag:  1,
			Price:     100,
			SumAmount: 1000,
			TransFlag: 12,
		},
		{
			ShopID:    "shop1",
			DocNo:     "DOC1",
			Barcode:   "BAR1",
			CalcQty:   10,
			CalcFlag:  1,
			Price:     100,
			SumAmount: 1000,
			TransFlag: 12,
		},
	}

	err = stockProcessRedisRepo.BulkAddStockData("SHOP1", "BAR1", stockProcessDocs)
	assert.Nil(t, err)
}

func TestCountMovementInRedis(t *testing.T) {
	cfg := config.NewConfig()
	cachePersister := microservice.NewCacher(cfg.CacherConfig())
	err := cachePersister.Healthcheck()
	assert.Nil(t, err)
	stockProcessRedisRepo := repositories.NewStockProcessRedisRepository(cachePersister)

	type args struct {
		shopID  string
		barcode string
	}

	cases := []struct {
		name     string
		args     args
		wantErr  bool
		wantData int64
	}{
		{
			name: "Test Count Movement In Redis Want Data 2",
			args: args{
				shopID:  "SHOP1",
				barcode: "BAR1",
			},
			wantErr:  false,
			wantData: 2,
		},
		{
			name: "Test Count Movement In Redis Want Error",
			args: args{
				shopID:  "SHOP1",
				barcode: "BAR2",
			},
			wantErr:  false,
			wantData: 0,
		},
	}
	for _, tt := range cases {

		size, err := stockProcessRedisRepo.GetStockDataLength(tt.args.shopID, tt.args.barcode)
		if tt.wantErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, tt.wantData, size)
		}
	}

}

func TestFindElementInRedisArray(t *testing.T) {
	cfg := config.NewConfig()
	cachePersister := microservice.NewCacher(cfg.CacherConfig())
	err := cachePersister.Healthcheck()
	assert.Nil(t, err)
	stockProcessRedisRepo := repositories.NewStockProcessRedisRepository(cachePersister)

	type args struct {
		shopID    string
		barcode   string
		stockData models.StockData
	}

	cases := []struct {
		name     string
		args     args
		wantErr  bool
		wantData int64
	}{
		{
			name: "Test Count Movement In Redis Want Data 2",
			args: args{
				shopID:  "SHOP1",
				barcode: "BAR1",
				stockData: models.StockData{

					ShopID:    "SHOP1",
					DocNo:     "DOC1",
					Barcode:   "BAR1",
					CalcQty:   10,
					CalcFlag:  1,
					Price:     100,
					SumAmount: 1000,
					TransFlag: 12,
				},
			},
			wantErr:  false,
			wantData: 0,
		},
		{
			name: "Test Count Movement In Redis Want Error",
			args: args{
				shopID:  "SHOP1",
				barcode: "BAR2",
				stockData: models.StockData{
					ShopID:    "SHOP999",
					DocNo:     "DOC999",
					Barcode:   "BAR999",
					CalcQty:   10,
					CalcFlag:  1,
					Price:     100,
					SumAmount: 1000,
					TransFlag: 12,
				},
			},
			wantErr:  false,
			wantData: -1,
		},
	}
	for _, tt := range cases {

		size, err := stockProcessRedisRepo.FindStockMovement(tt.args.shopID, tt.args.barcode, tt.args.stockData)
		if tt.wantErr {
			assert.NotNil(t, err, tt.name)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, tt.wantData, size, tt.name)
		}
	}
}
