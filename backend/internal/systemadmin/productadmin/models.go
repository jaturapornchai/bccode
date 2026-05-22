package productadmin

type RequestReSyncProductBarcode struct {
	ShopID string `json:"shopid"`
	Barcode string `json:"barcode,omitempty"`
}
