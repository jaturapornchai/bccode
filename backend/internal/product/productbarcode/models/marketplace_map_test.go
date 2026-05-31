package models

import (
	"encoding/json"
	"testing"
)

// TestMarketplaceProductMapJSONRoundTrip verifies the unified marketplace map
// marshals/unmarshals without losing fields and keeps the exact JSON keys the
// Next.js frontend (lib/product-barcode/types.ts) relies on.
func TestMarketplaceProductMapJSONRoundTrip(t *testing.T) {
	in := MarketplaceProductMap{
		Platform:      "shopee",
		AccountID:     "acc-1",
		ShopID:        "shop-1",
		MarketItemID:  "ITEM123",
		MarketModelID: "MODEL456",
		ItemURL:       "https://shopee.co.th/product/1/2",
		SellerSKU:     "SKU-001",
		ShopSKU:       "SHOP-SKU-001",
		GTIN:          "8850001234567",
		CategoryID:    "100012",
		CategoryName:  "Mobile & Gadgets",
		BrandID:       "BR-9",
		Currency:      "THB",
		CustomPrice:   199.5,
		PlatformPrice: 210,
		PlatformStock: 42,
		SyncStock:     true,
		SyncPrice:     true,
		Status:        "LIVE",
		RejectReason:  "",
		DaysToShip:    3,
		IsPreOrder:    false,
		SyncEnabled:   true,
		SyncStatus:    "ok",
		LastSyncAt:    "2026-05-29T10:00:00Z",
		LastSyncError: "",
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var out MarketplaceProductMap
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if out != in {
		t.Fatalf("round-trip mismatch:\n got  %+v\n want %+v", out, in)
	}

	// Contract guard: these JSON keys must stay in sync with the frontend type.
	wantKeys := []string{
		"platform", "account_id", "shop_id", "market_item_id", "market_model_id",
		"item_url", "seller_sku", "shop_sku", "gtin", "category_id", "category_name",
		"brand_id", "currency", "custom_price", "platform_price", "platform_stock",
		"sync_stock", "sync_price", "status", "reject_reason", "days_to_ship",
		"is_pre_order", "sync_enabled", "sync_status", "last_sync_at", "last_sync_error",
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode to map failed: %v", err)
	}
	for _, k := range wantKeys {
		if _, ok := decoded[k]; !ok {
			t.Errorf("missing JSON key %q in MarketplaceProductMap output", k)
		}
	}
}

// TestMarketplaceSKUMapJSONRoundTrip guards the sub-barcode (variation) map.
func TestMarketplaceSKUMapJSONRoundTrip(t *testing.T) {
	in := MarketplaceSKUMap{
		Platform:      "lazada",
		AccountID:     "acc-2",
		ShopID:        "shop-2",
		MarketItemID:  "L-ITEM-9",
		MarketModelID: "L-SKU-9",
		SellerSKU:     "SKU-VAR-01",
		ShopSKU:       "LZD-SHOP-SKU",
		GTIN:          "8850009876543",
		Currency:      "THB",
		SyncStock:     true,
		SyncPrice:     false,
		CustomPrice:   89,
		PlatformPrice: 95,
		PlatformStock: 7,
		Status:        "UNLIST",
		SyncEnabled:   true,
		LastSyncAt:    "2026-05-29T11:00:00Z",
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var out MarketplaceSKUMap
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if out != in {
		t.Fatalf("round-trip mismatch:\n got  %+v\n want %+v", out, in)
	}
}
