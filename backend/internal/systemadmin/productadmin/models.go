package productadmin

type RequestReSyncProductBarcode struct {
	HoldingCode string `json:"holdingcode"`
	Barcode     string `json:"barcode,omitempty"`
}
