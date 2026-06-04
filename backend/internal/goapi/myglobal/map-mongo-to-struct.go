package myglobal

import (
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"

	"github.com/mitchellh/mapstructure"
)

func MapBarcodeFromMongoToStruct(productBarcode models.ProcessMongoBarcodeModel) (models.BarcodeModel, []models.BarcodeRefModel) {
	var barcodeModel = models.NewBarcodeModel()
	var barcodeRefModel = []models.BarcodeRefModel{}

	// กำหนดค่า itemCode
	barcodeModel.ItemCode = productBarcode.ItemCode
	if barcodeModel.ItemCode == "" {
		barcodeModel.ItemCode = productBarcode.Barcode
	}

	// กำหนดค่าพื้นฐาน
	barcodeModel.HoldingCode = productBarcode.HoldingCode
	barcodeModel.Barcode = productBarcode.Barcode
	barcodeModel.UnitCode = productBarcode.ItemUnitCode
	barcodeModel.ItemType = productBarcode.ItemType
	barcodeModel.MaterialType = productBarcode.MaterialType

	// ดึงชื่อสินค้าหลายภาษา
	if len(productBarcode.Names) > 0 {
		barcodeModel.Name0 = productBarcode.Names[0].Name
	}
	if len(productBarcode.Names) > 1 {
		barcodeModel.Name1 = productBarcode.Names[1].Name
	}
	if len(productBarcode.Names) > 2 {
		barcodeModel.Name2 = productBarcode.Names[2].Name
	}
	if len(productBarcode.Names) > 3 {
		barcodeModel.Name3 = productBarcode.Names[3].Name
	}
	if len(productBarcode.Names) > 4 {
		barcodeModel.Name4 = productBarcode.Names[4].Name
	}
	if len(productBarcode.Names) > 5 {
		barcodeModel.Name5 = productBarcode.Names[5].Name
	}

	// ดึงชื่อหน่วย
	if len(productBarcode.ItemUnitNames) > 0 {
		barcodeModel.UnitName = productBarcode.ItemUnitNames[0].Name
	}

	// ดึง ImageUri
	barcodeModel.ImageUri = productBarcode.ImageUri

	// ดึงราคาจาก Prices array
	// keynumber = 1 → ราคาขายปลีก (PriceRetail)
	// keynumber = 2 → ราคาขายส่ง (Price1) ถ้าไม่มี ใช้ keynumber=1 แทน
	for _, priceItem := range productBarcode.Prices {
		if priceItem.KeyNumber == 1 {
			barcodeModel.PriceRetail = priceItem.Price
			// ถ้ายังไม่มี Price1 ให้ใช้ราคาขายปลีกก่อน
			if barcodeModel.Price1 == 0 {
				barcodeModel.Price1 = priceItem.Price
			}
		}
		if priceItem.KeyNumber == 2 {
			barcodeModel.Price1 = priceItem.Price
		}
	}

	// ดึงข้อมูล unit conversion และ barcodeRef จาก RefBarCodes
	barcodeModel.UnitStand = 1.0
	barcodeModel.UnitDivide = 1.0
	barcodeModel.BarcodeRefUnitStand = 1.0
	barcodeModel.BarcodeRefUnitDivide = 1.0
	if len(productBarcode.RefBarCodes) > 0 {
		barcodeModel.IsStock = 0
		// กำหนด BarcodeRef จาก RefBarCodes ตัวแรก
		barcodeModel.BarcodeRef = productBarcode.RefBarCodes[0].Barcode
		barcodeModel.BarcodeRefUnitStand = productBarcode.RefBarCodes[0].UnitStand
		barcodeModel.BarcodeRefUnitDivide = productBarcode.RefBarCodes[0].UnitDivide

		// สร้าง BarcodeRefModel สำหรับทุก RefBarCodes
		for _, ref := range productBarcode.RefBarCodes {
			barcodeRefModel = append(barcodeRefModel, models.BarcodeRefModel{
				HoldingCode: productBarcode.HoldingCode,
				Barcode:     productBarcode.Barcode,
				BarcodeRef:  ref.Barcode,
				ItemCode:    barcodeModel.ItemCode,
				UnitCode:    ref.ItemUnitCode,
				StandValue:  ref.UnitStand,
				DivideValue: ref.UnitDivide,
			})
		}
	} else {
		barcodeModel.IsStock = 1
	}

	// คำนวณ checksum
	barcodeModel.Checksum = CalculateMD5(fmt.Sprintf("%v", productBarcode))

	return barcodeModel, barcodeRefModel
}

func ProcessProductBarcodeDecode(jsonData string) models.ProcessMongoBarcodeModel {
	jsonDecode := map[string]any{}
	err := json.Unmarshal([]byte(jsonData), &jsonDecode)
	if err != nil {
		logger.Error("Error unmarshal: %v", err)
		return models.ProcessMongoBarcodeModel{}
	}

	fixedData := FixMongoData(jsonDecode)

	var productBarcode models.ProcessMongoBarcodeModel
	decoderConfig := mapstructure.DecoderConfig{
		DecodeHook: CustomDecodeHookFunc,
		Result:     &productBarcode,
		TagName:    "json",
	}
	decoder, err := mapstructure.NewDecoder(&decoderConfig)
	if err != nil {
		logger.Fatal("%v", err)
	}

	err = decoder.Decode(fixedData)
	if err != nil {
		logger.Error("Error decoding map to struct: %v", err)
	}
	return productBarcode
}
