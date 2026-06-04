package usecases_test

import (
	"encoding/json"
	commonModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/product/productbarcode/usecases"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeserializeJsonProductBarcode(t *testing.T) {

	var jsonStr = `{"id":"000000000000000000000000","holding_code":"2QoilMQkX9i6vtAE88ilEubnrhz","guid_fixed":"2QotSoH9vZZ1LW7loBAZbuCt1pW","itemcode":"","barcode":"KSC-001","group_code":"","group_names":[],"names":[{"code":"th","name":"SIZE S ปีกกลาง 4 ชิ้น","isauto":false,"isdelete":false},{"code":"en","name":"CHICKEN WINGS SIZE S (4 PCS.)","isauto":false,"isdelete":false},{"code":"ko","name":"미들 윙 4 개","isauto":false,"isdelete":false}],"xsorts":[],"item_unit_code":"PLATE","itemunitnames":[{"code":"th","name":"จาน","isauto":false,"isdelete":false}],"prices":[{"key_number":1,"price":99},{"key_number":2,"price":0}],"imageuri":"","options":[{"guid":"a6270c74-5a86-488a-951b-c63d651e52bb","choicetype":0,"maxselect":2,"minselect":1,"names":[{"code":"th","name":"โซลมายด์ ชิกเก้นท์ *เลือกได้ 2 ซอส","isauto":false,"isdelete":false},{"code":"en","name":"SEOULMIND CHICKEN ","isauto":false,"isdelete":false},{"code":"ko","name":"서울 마인드 치킨(2가지 소스 선택할 수 있음)","isauto":false,"isdelete":false}],"choices":[{"guid":"0e3e9cd9-fa5c-496e-b391-f95287b54393","refbarcode":"","refproductcode":"","refunitcode":"","names":[{"code":"th","name":"ซอสเกาหลี","isauto":false,"isdelete":false},{"code":"en","name":"KOREA SAUCE","isauto":false,"isdelete":false},{"code":"ko","name":"한국 소스","isauto":false,"isdelete":false}],"price":"","qty":0,"isstock":false,"isdefault":false},{"guid":"8580a534-591e-46f6-801e-1504502d3e42","refbarcode":"","refproductcode":"","refunitcode":"","names":[{"code":"th","name":"ซอสกระเทียม ","isauto":false,"isdelete":false},{"code":"en","name":"GARLIC SAUCE","isauto":false,"isdelete":false},{"code":"ko","name":"마늘 소스","isauto":false,"isdelete":false}],"price":"","qty":0,"isstock":false,"isdefault":false},{"guid":"8074cfa4-a82b-4dbd-bf32-cb392cbad345","refbarcode":"","refproductcode":"","refunitcode":"","names":[{"code":"th","name":"ซอสหมาล่า ","isauto":false,"isdelete":false},{"code":"en","name":"MALA SAUCE ","isauto":false,"isdelete":false},{"code":"ko","name":"마라 소스","isauto":false,"isdelete":false}],"price":"","qty":0,"isstock":false,"isdefault":false}]}],"images":null,"useimageorcolor":true,"colorselect":"","colorselecthex":"","condition":false,"dividevalue":1,"standvalue":1,"isusesubbarcodes":false,"item_type":0,"tax_type":0,"vat_type":0,"issumpoint":false,"maxdiscount":"","isdividend":false,"refunitnames":null,"stockbarcode":"","qty":0,"refdividevalue":0,"refstandvalue":0,"vatcal":0,"refbarcodes":[],"bom":[{"guid_fixed":"some-guid-fixed-value","names":[{"code":"name-code-1","name":"Name 1","isauto":true,"isdelete":false},{"code":"name-code-2","name":"Name 2","isauto":false,"isdelete":true}],"item_unit_code":"item-unit-code","itemunitnames":[{"code":"item-unit-code-1","name":"Item Unit Name 1","isauto":false,"isdelete":false}],"barcode":"1234567890","condition":true,"dividevalue":2.5,"standvalue":10,"qty":100}]}`

	var giveProductBarcodeDoc models.ProductBarcodeDoc

	err := json.Unmarshal([]byte(jsonStr), &giveProductBarcodeDoc)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, "2QoilMQkX9i6vtAE88ilEubnrhz", giveProductBarcodeDoc.HoldingCode)

	phaser := usecases.ProductBarcodePhaser{}
	got, err := phaser.PhaseProductBarcodeDoc(&giveProductBarcodeDoc)

	assert.NoError(t, err)
	assert.Equal(t, "2QoilMQkX9i6vtAE88ilEubnrhz", got.HoldingCode)
	assert.Equal(t, "PLATE", got.UnitCode)
	assert.Equal(t, "th", *got.UnitNames[0].Code)
	assert.Equal(t, "จาน", *got.UnitNames[0].Name)
	assert.Equal(t, 1, len(got.BOM))

	assert.Equal(t, "some-guid-fixed-value", (*got).BOM[0].BarcodeGuidFixed)
	assert.Equal(t, "1234567890", (*got).BOM[0].Barcode)

}

func TestProductBarcodePhaser(t *testing.T) {

	var jsonStr = `{
	"id": "000000000000000000000000",
	"holding_code": "30LbRx3l0SLaK84gLpcF0W4x9Z0",
	"guid_fixed": "335YJBgJXUok8crFh7u1t8E4mME",
	"itemcode": "ITEM99",
	"barcode": "955004",
	"groupguid": "",
	"group_code": "",
	"group_names": [],
	"groupsuboneguid": "",
	"groupsubonecode": "",
	"groupsubonenames": [],
	"groupsubtwoguid": "",
	"groupsubtwocode": "",
	"groupsubtwonames": [],
	"brandguid": "",
	"brand_code": "BRAND_CODE",
	"brandnames": [
		{
			"code": "th",
			"name": "แบรนด์",
			"isauto": false,
			"isdelete": false
		}
	],
	"designguid": "",
	"designcode": "",
	"designnames": [],
	"modelguid": "",
	"modelcode": "",
	"modelnames": [],
	"patternguid": "",
	"patterncode": "",
	"patternnames": [],
	"gradeguid": "",
	"gradecode": "",
	"gradenames": [],
	"category_guid": "",
	"categorycode": "",
	"category_names": [],
	"classguid": "",
	"classcode": "",
	"classnames": [],
	"orderpoint": 0,
	"minpoint": 0,
	"maxpoint": 0,
	"refguidfixed": "",
	"names": [
		{
			"code": "th",
			"name": "ref3",
			"isauto": false,
			"isdelete": false
		}
	],
	"xsorts": [],
	"item_guid": "",
	"itemunitguid": "30LbbTTQ5662KOBwQHh5DPnuI4r",
	"item_unit_code": "BKT",
	"itemunitnames": [
		{
			"code": "th",
			"name": "ปี๊บ",
			"isauto": false,
			"isdelete": false
		},
		{
			"code": "en",
			"name": "BUCKET",
			"isauto": false,
			"isdelete": false
		}
	],
	"itemunitsize": 0,
	"prices": [
		{
			"key_number": 1,
			"price": 2
		},
		{
			"key_number": 2,
			"price": 0
		},
		{
			"key_number": 9,
			"price": 0
		},
		{
			"key_number": 10,
			"price": 0
		},
		{
			"key_number": 11,
			"price": 0
		},
		{
			"key_number": 12,
			"price": 0
		},
		{
			"key_number": 13,
			"price": 0
		},
		{
			"key_number": 14,
			"price": 0
		},
		{
			"key_number": 15,
			"price": 0
		},
		{
			"key_number": 16,
			"price": 0
		},
		{
			"key_number": 17,
			"price": 0
		}
	],
	"imageuri": "",
	"options": [],
	"images": null,
	"useimageorcolor": true,
	"colorselect": "",
	"colorselecthex": "",
	"condition": false,
	"dividevalue": 1,
	"standvalue": 1,
	"isusesubbarcodes": false,
	"is_main_barcode": true,
	"item_type": 0,
	"materialtype": 0,
	"tax_type": 0,
	"vat_type": 0,
	"issumpoint": true,
	"maxdiscount": "",
	"isdividend": false,
	"fixedcost": [],
	"refunitnames": null,
	"stockbarcode": "",
	"qty": 0,
	"refdividevalue": 0,
	"refstandvalue": 0,
	"vatcal": 0,
	"isalacarte": true,
	"ordertypes": [],
	"producttype": {
		"guid_fixed": "",
		"code": "",
		"names": []
	},
	"issplitunitprint": true,
	"isonlystaff": false,
	"foodtype": 0,
	"discount": "",
	"isstockforrestaurant": false,
	"manufacturerguid": "",
	"manufacturercode": "",
	"manufacturernames": [],
	"dimensions": [],
	"isdiscountpointofpurchase": true,
	"restaurant": {
		"isforrestaurant": true,
		"isfortakeaway": true,
		"isfordelivery": true,
		"isforcustomer": true,
		"isforcustomerpreorder": true
	},
	"isalert": false,
	"alertdescription": "",
	"description": "",
	"timeforsales": [],
	"refbarcodes": [],
	"bom": [],
	"businesstypes": [
		{
			"guid_fixed": "30LbS14GyApu6OX70r23D2suc35",
			"code": "003",
			"names": [
				{
					"code": "th",
					"name": "ธุรกิจก่อสร้าง วัสดุก่อสร้าง และพัฒนาอสังหาริมทรัพย์",
					"isauto": false,
					"isdelete": false
				},
				{
					"code": "en",
					"name": "construction business construction materials and real estate development",
					"isauto": false,
					"isdelete": false
				}
			],
			"isignore": false
		}
	],
	"ignorebranches": []
}`

	var giveProductBarcodeDoc models.ProductBarcodeDoc

	err := json.Unmarshal([]byte(jsonStr), &giveProductBarcodeDoc)
	if err != nil {
		panic(err)
	}

	// Verify input document
	assert.Equal(t, "30LbRx3l0SLaK84gLpcF0W4x9Z0", giveProductBarcodeDoc.HoldingCode)

	phaser := usecases.ProductBarcodePhaser{}
	got, err := phaser.PhaseProductBarcodeDoc(&giveProductBarcodeDoc)

	assert.NoError(t, err)

	// Test all ProductBarcodePg fields
	t.Run("Assert All ProductBarcodePg Fields", func(t *testing.T) {
		// Basic identifiers
		assert.Equal(t, "30LbRx3l0SLaK84gLpcF0W4x9Z0", got.HoldingCode, "HoldingCode should match")
		assert.Equal(t, "955004", got.Barcode, "Barcode should match")
		assert.Equal(t, "", got.ParID, "ParID should be empty (not set by phaser)")

		// Item details
		assert.Equal(t, "ITEM99", got.ItemCode, "ItemCode should match")
		assert.Equal(t, int8(0), got.ItemType, "ItemType should match")
		assert.Equal(t, int8(0), got.MaterialType, "MaterialType should match")
		assert.Equal(t, "BKT", got.UnitCode, "UnitCode should match from ItemUnitCode")

		// Names
		assert.NotNil(t, got.Names, "Names should not be nil")
		namesList := []commonModels.NameX(got.Names)
		assert.Equal(t, 1, len(namesList), "Should have 1 name entry")
		assert.Equal(t, "th", *namesList[0].Code, "Name code should be 'th'")
		assert.Equal(t, "ref3", *namesList[0].Name, "Name should be 'ref3'")
		assert.False(t, namesList[0].IsAuto, "IsAuto should be false")
		assert.False(t, namesList[0].IsDelete, "IsDelete should be false")

		// Unit names
		assert.NotNil(t, got.UnitNames, "UnitNames should not be nil")
		unitNamesList := []commonModels.NameX(got.UnitNames)
		assert.Equal(t, 2, len(unitNamesList), "Should have 2 unit name entries")

		// Check Thai unit name
		assert.Equal(t, "th", *unitNamesList[0].Code, "First unit name code should be 'th'")
		assert.Equal(t, "ปี๊บ", *unitNamesList[0].Name, "First unit name should be 'ปี๊บ'")
		assert.False(t, unitNamesList[0].IsAuto, "First unit name IsAuto should be false")
		assert.False(t, unitNamesList[0].IsDelete, "First unit name IsDelete should be false")

		// Check English unit name
		assert.Equal(t, "en", *unitNamesList[1].Code, "Second unit name code should be 'en'")
		assert.Equal(t, "BUCKET", *unitNamesList[1].Name, "Second unit name should be 'BUCKET'")
		assert.False(t, unitNamesList[1].IsAuto, "Second unit name IsAuto should be false")
		assert.False(t, unitNamesList[1].IsDelete, "Second unit name IsDelete should be false")

		// Quantities and values
		assert.Equal(t, float64(0), got.BalanceQty, "BalanceQty should be 0")
		assert.Equal(t, float64(0), got.BalanceAmount, "BalanceAmount should be 0")
		assert.Equal(t, float64(0), got.AverageCost, "AverageCost should be 0")
		assert.Equal(t, float64(1), got.StandValue, "StandValue should be 1 (default)")
		assert.Equal(t, float64(1), got.DivideValue, "DivideValue should be 1 (default)")

		// Barcode reference
		assert.Equal(t, "955004", got.MainBarcodeRef, "MainBarcodeRef should be same as Barcode")

		// Category fields - all should be empty strings based on JSON
		assert.Equal(t, "BRAND_CODE", got.BrandCode, "BrandCode should be BRAND_CODE")
		assert.Equal(t, "", got.CategoryCode, "CategoryCode should be empty")
		assert.Equal(t, "", got.ClassCode, "ClassCode should be empty")
		assert.Equal(t, "", got.DesignCode, "DesignCode should be empty")
		assert.Equal(t, "", got.GradeCode, "GradeCode should be empty")
		assert.Equal(t, "", got.GroupCode, "GroupCode should be empty")
		assert.Equal(t, "", got.GroupSubOneCode, "GroupSubOneCode should be empty")
		assert.Equal(t, "", got.GroupSubTwoCode, "GroupSubTwoCode should be empty")
		assert.Equal(t, "", got.ModelCode, "ModelCode should be empty")
		assert.Equal(t, "", got.PatternCode, "PatternCode should be empty")

		// Category JSONB names - should be empty based on JSON
		brandNames := []commonModels.NameX(got.BrandName)
		assert.Equal(t, 1, len(brandNames), "BrandNames should be 1")
		assert.Equal(t, "th", *brandNames[0].Code, "Brand name code should be 'th'")
		assert.Equal(t, "แบรนด์", *brandNames[0].Name, "Brand name should be 'แบรนด์'")

		categoryNames := []commonModels.NameX(got.CategoryName)
		assert.Equal(t, 0, len(categoryNames), "CategoryNames should be empty")

		classNames := []commonModels.NameX(got.ClassNames)
		assert.Equal(t, 0, len(classNames), "ClassNames should be empty")

		designNames := []commonModels.NameX(got.DesignNames)
		assert.Equal(t, 0, len(designNames), "DesignNames should be empty")

		gradeNames := []commonModels.NameX(got.GradeNames)
		assert.Equal(t, 0, len(gradeNames), "GradeNames should be empty")

		groupNames := []commonModels.NameX(got.GroupNames)
		assert.Equal(t, 0, len(groupNames), "GroupNames should be empty")

		groupSubOneNames := []commonModels.NameX(got.GroupSubOneNames)
		assert.Equal(t, 0, len(groupSubOneNames), "GroupSubOneNames should be empty")

		groupSubTwoNames := []commonModels.NameX(got.GroupSubTwoNames)
		assert.Equal(t, 0, len(groupSubTwoNames), "GroupSubTwoNames should be empty")

		modelNames := []commonModels.NameX(got.ModelNames)
		assert.Equal(t, 0, len(modelNames), "ModelNames should be empty")

		patternNames := []commonModels.NameX(got.PatternNames)
		assert.Equal(t, 0, len(patternNames), "PatternNames should be empty")

		// BOM should be empty based on JSON
		bomList := []models.BOMProductBarcode(got.BOM)
		assert.Equal(t, 0, len(bomList), "BOM should be empty")
	})
}

func TestProductBarcodePhaserKeepsProductSetClassification(t *testing.T) {
	doc := models.ProductBarcodeDoc{}
	doc.HoldingCode = "shop-1"
	doc.GuidFixed = "guid-1"
	doc.Barcode = "SET-001"
	doc.ItemCode = "SET-001"
	doc.ItemType = models.ItemTypeSet
	doc.MaterialType = models.MaterialTypeSet

	phaser := usecases.ProductBarcodePhaser{}
	got, err := phaser.PhaseProductBarcodeDoc(&doc)

	assert.NoError(t, err)
	assert.Equal(t, models.ItemTypeSet, got.ItemType)
	assert.Equal(t, models.MaterialTypeSet, got.MaterialType)
}

func TestProductBarcodePhaserRejectsInvalidProductSetClassification(t *testing.T) {
	doc := models.ProductBarcodeDoc{}
	doc.HoldingCode = "shop-1"
	doc.GuidFixed = "guid-1"
	doc.Barcode = "SET-INVALID"
	doc.ItemCode = "SET-INVALID"
	doc.ItemType = models.ItemTypeSet
	doc.MaterialType = models.MaterialTypeGeneral

	phaser := usecases.ProductBarcodePhaser{}
	got, err := phaser.PhaseProductBarcodeDoc(&doc)

	assert.Nil(t, got)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected materialtype 3")
}
