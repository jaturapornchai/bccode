package productadmin

type RequestReSyncProductBarcode struct {
	HoldingCode string `json:"holding_code"`
	Barcode     string `json:"barcode,omitempty"`
}
