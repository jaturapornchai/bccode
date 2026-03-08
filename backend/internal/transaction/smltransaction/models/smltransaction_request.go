package models

type SMLTransactionRequest struct {
	Collection string                 `json:"collection" validate:"required,min=1,max=100"`
	KeyID      string                 `json:"keyid" validate:"required,min=1,max=100"`
	Body       map[string]interface{} `json:"body" validate:"required"`
}

type SMLTransactionBulkRequest struct {
	Collection string                   `json:"collection" validate:"required,min=1,max=100"`
	KeyID      string                   `json:"keyid" validate:"required,min=1,max=100"`
	Body       []map[string]interface{} `json:"body" validate:"required"`
}

type SMLTransactionKeyRequest struct {
	Collection string   `json:"collection" validate:"required,min=1,max=100"`
	KeyID      string   `json:"keyid" validate:"required,min=1,max=100"`
	DeleteKeys []string `json:"deletekeys" validate:"required"`
}
