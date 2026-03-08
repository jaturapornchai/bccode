package microservice_test

import (
	"context"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/pkg/microservice"
	"testing"
)

// type PersisterClickHouseConfig struct{}

// func NewPersisterClickHouseConfig() *PersisterClickHouseConfig {
// 	return &PersisterClickHouseConfig{}
// }

// func (p PersisterClickHouseConfig) ServerAddress() []string {
// 	addr := "103.212.36.91:19000"
// 	return []string{addr}
// }

// func (p PersisterClickHouseConfig) DatabaseName() string {
// 	return "clouddev"
// }

// func (p PersisterClickHouseConfig) Username() string {
// 	return getEnv("CH_USERNAME", "")
// }

// func (p PersisterClickHouseConfig) Password() string {
// 	return getEnv("CH_PASSWORD", "")
// }

type TestClickHouseModel struct {
	Name string `ch:"name"`
	Age  uint8  `ch:"age"`
}

func (m *TestClickHouseModel) TableName() string {
	return "test1"
}

type StockBalanceImport struct {
	ShopID    string  `json:"shopid" ch:"shopid"`
	TaskID    string  `json:"taskid" ch:"taskid"`
	RowNumber uint32  `json:"rownumber" ch:"rownumber"`
	Barcode   string  `json:"barcode" ch:"barcode"`
	Name      string  `json:"name" ch:"name"`
	UnitCode  string  `json:"unitcode" ch:"unitcode"`
	Qty       float64 `json:"qty" ch:"qty"`
	Price     float64 `json:"price" ch:"price"`
	SumAmount float64 `json:"sumamount" ch:"sumamount"`
}

func (StockBalanceImport) TableName() string {
	return "stockbalanceimport"
}

func TestClickHouseCount(t *testing.T) {

	cfg := config.NewConfig()
	clickhouseCfg := cfg.ClickHouseConfig()

	pst := microservice.NewPersisterClickHouse(clickhouseCfg)

	count, err := pst.Count(context.Background(), &StockBalanceImport{}, "guidfixed = ?", "2XKwDWexFMoiFPz5G3IMEp9bZoY")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(count)
}

func TestClickHouseProductInsert(t *testing.T) {

	cfg := config.NewConfig()
	clickhouseCfg := cfg.ClickHouseConfig()

	pst := microservice.NewPersisterClickHouse(clickhouseCfg)

	data := &StockBalanceImport{
		ShopID:    "123",
		TaskID:    "123",
		RowNumber: 1,
		Barcode:   "123",
		Name:      "123",
		UnitCode:  "123",
		Qty:       1.0,
		Price:     1.0,
		SumAmount: 1.0,
	}

	err := pst.Create(context.Background(), data)
	if err != nil {
		t.Fatal(err)
	}

}

func TestClickHouseProductBulkInsert(t *testing.T) {

	cfg := config.NewConfig()
	clickhouseCfg := cfg.ClickHouseConfig()

	pst := microservice.NewPersisterClickHouse(clickhouseCfg)

	data := []StockBalanceImport{
		{
			ShopID:    "123",
			TaskID:    "123",
			RowNumber: 1,
			Barcode:   "123",
			Name:      "123",
			UnitCode:  "123",
			Qty:       1.0,
			Price:     2.0,
			SumAmount: 3.0,
		},
		{
			ShopID:    "123",
			TaskID:    "123",
			RowNumber: 1,
			Barcode:   "123",
			Name:      "123",
			UnitCode:  "123",
			Qty:       4.0,
			Price:     5.0,
			SumAmount: 6.0,
		},
	}

	temp := make([]interface{}, len(data))
	for i, v := range data {
		temp[i] = &v
	}

	err := pst.CreateInBatch(context.Background(), temp)
	if err != nil {
		t.Fatal(err)
	}

}

func TestClickHouseProductUpdate(t *testing.T) {

	cfg := config.NewConfig()
	clickhouseCfg := cfg.ClickHouseConfig()

	pst := microservice.NewPersisterClickHouse(clickhouseCfg)

	err := pst.Exec(context.Background(), "ALTER TABLE clouddev.stockbalanceimport UPDATE name = ? WHERE barcode = ?", "modify", "123")
	if err != nil {
		t.Fatal(err)
	}

}
