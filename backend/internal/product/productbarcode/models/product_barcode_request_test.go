package models

import "testing"

func TestProductBarcodeRequestCoreOnlyKeepsBarcodeMediaAndDropsProductStockFields(t *testing.T) {
	request := ProductBarcodeRequest{
		ProductBarcodeBase: ProductBarcodeBase{
			ItemCode:      "ITEM-1",
			Barcode:       "885000000001",
			ItemUnitCode:  "PCS",
			DivideValue:   1,
			StandValue:    1,
			GroupCode:     "FOOD",
			ImageURI:      "barcode-main.png",
			Images:        &[]ProductImage{{XOrder: 1, URI: "barcode-gallery.png"}},
			Videos:        &[]ProductVideo{{XOrder: 1, URI: "barcode-video.mp4", PosterURI: "barcode-video.jpg"}},
			OrderPoint:    10,
			MinPoint:      5,
			MaxPoint:      20,
			Qty:           99,
			Description:   "barcode packaging detail",
			PackageWeight: 2,
		},
		BOM: []BOMRequest{{Barcode: "CHILD"}},
	}

	core := request.CoreOnly()

	if core.ItemCode != request.ItemCode || core.Barcode != request.Barcode || core.ItemUnitCode != request.ItemUnitCode {
		t.Fatalf("barcode identity was not preserved: %#v", core.ProductBarcodeBase)
	}
	if core.ImageURI != request.ImageURI || core.Description != request.Description || core.Images == nil || len(*core.Images) != 1 || (*core.Images)[0].URI != "barcode-gallery.png" || core.Videos == nil || len(*core.Videos) != 1 || (*core.Videos)[0].URI != "barcode-video.mp4" || (*core.Videos)[0].PosterURI != "barcode-video.jpg" {
		t.Fatalf("barcode media or description was dropped: %#v", core.ProductBarcodeBase)
	}
	if core.GroupCode != "" || core.OrderPoint != 0 || core.MinPoint != 0 || core.MaxPoint != 0 || core.Qty != 0 || core.PackageWeight != 0 {
		t.Fatalf("product or stock fields leaked into barcode core: %#v", core.ProductBarcodeBase)
	}
	if len(core.BOM) != 0 || len(core.RefBarcodes) != 0 || len(core.BusinessTypes) != 0 || len(core.IgnoreBranches) != 0 {
		t.Fatalf("product relations leaked into barcode core: %#v", core)
	}
}
