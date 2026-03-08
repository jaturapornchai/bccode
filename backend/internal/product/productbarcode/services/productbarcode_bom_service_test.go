package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/product/productbarcode/services"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildBOMView(t *testing.T) {

	data01, _ := findByBarcode(context.Background(), "shopID", "01")
	productBarcodeDict := map[string]models.ProductBarcodeDoc{}
	bomViewDict := map[string]*models.ProductBarcodeBOMView{}

	bomView := models.ProductBarcodeBOMView{}
	bomView.FromProductBarcode(data01.ProductBarcodeData)

	if _, ok := bomViewDict[data01.Barcode]; !ok {
		bomViewDict[data01.Barcode] = &bomView
	}

	services.BuildBOMView(context.Background(), findByBarcode, 1, &productBarcodeDict, &bomViewDict, "shopID", data01.BOM, &bomView.BOM)

	bomView2 := models.ProductBarcodeBOMView{}
	bomView2.FromProductBarcode(data01.ProductBarcodeData)

	if _, ok := bomViewDict[data01.Barcode]; !ok {
		bomViewDict[data01.Barcode] = &bomView
	}

	services.BuildBOMView(context.Background(), findByBarcode, 1, &productBarcodeDict, &bomViewDict, "shopID", data01.BOM, &bomView2.BOM)

	assert.Equal(t, "guid1", bomViewDict["01"].BarcodeGuidFixed)
}

func TestBuildBOMView2(t *testing.T) {

	data01, _ := findByBarcode(context.Background(), "shopID", "01")
	productBarcodeDict := map[string]models.ProductBarcodeDoc{}
	bomViewDict := map[string]*models.ProductBarcodeBOMView{}
	bomView := models.ProductBarcodeBOMView{}

	services.BuildBOMViewCache(context.Background(), findByBarcode, 0, &productBarcodeDict, &bomViewDict, "shopID", data01.Barcode, []models.BOMProductBarcode{}, &bomView)

	jsonData, err := json.Marshal(bomView)

	assert.Nil(t, err)

	fmt.Println(string(jsonData))
	fmt.Println(bomView.Qty)
}

var data = map[string]models.ProductBarcodeDoc{}

func findByBarcode(ctx context.Context, shopID string, barcode string) (models.ProductBarcodeDoc, error) {

	if len(data) == 0 {
		fmt.Println("init data")
		data01 := models.ProductBarcodeDoc{}
		data01.GuidFixed = "guid1"
		data01.Barcode = "01"
		data01.Names = &[]common.NameX{
			*common.NewNameXWithCodeName("en", "name01"),
		}
		data01.ItemUnitCode = "unit01"
		data01.ItemUnitNames = &[]common.NameX{
			*common.NewNameXWithCodeName("en", "unit01"),
		}

		data02 := models.ProductBarcodeDoc{}
		data02.GuidFixed = "guid2"
		data02.Barcode = "02"
		data02.Names = &[]common.NameX{
			*common.NewNameXWithCodeName("en", "name02"),
		}
		data02.ItemUnitCode = "unit02"
		data02.ItemUnitNames = &[]common.NameX{
			*common.NewNameXWithCodeName("en", "unit02"),
		}

		data03 := models.ProductBarcodeDoc{}
		data03.GuidFixed = "guid3"
		data03.Barcode = "03"
		data03.Names = &[]common.NameX{
			*common.NewNameXWithCodeName("en", "name03"),
		}
		data03.ItemUnitCode = "unit03"
		data03.ItemUnitNames = &[]common.NameX{
			*common.NewNameXWithCodeName("en", "unit03"),
		}

		data04 := models.ProductBarcodeDoc{}
		data04.GuidFixed = "guid4"
		data04.Barcode = "04"
		data04.Names = &[]common.NameX{
			*common.NewNameXWithCodeName("en", "name04"),
		}
		data04.ItemUnitCode = "unit04"
		data04.ItemUnitNames = &[]common.NameX{
			*common.NewNameXWithCodeName("en", "unit04"),
		}

		data05 := models.ProductBarcodeDoc{}
		data05.GuidFixed = "guid5"
		data05.Barcode = "05"
		data05.Names = &[]common.NameX{
			*common.NewNameXWithCodeName("en", "name05"),
		}

		data03.BOM = &[]models.BOMProductBarcode{
			data04.ToBOM(),
			data05.ToBOM(),
		}

		data02.BOM = &[]models.BOMProductBarcode{
			data03.ToBOM(),
		}

		data01.BOM = &[]models.BOMProductBarcode{
			data02.ToBOM(),
		}

		data["01"] = data01
		data["02"] = data02
		data["03"] = data03
		data["04"] = data04
		data["05"] = data05

		if _, ok := data[barcode]; !ok {
			return models.ProductBarcodeDoc{}, errors.New("not found")
		}
	} else {
		fmt.Println("data exist")
	}

	tempBarcode01 := data[barcode]

	return tempBarcode01, nil
}
