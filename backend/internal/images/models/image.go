package models

type Image struct {
	Uri string `json:"uri"`
}

type UploadImageResponse struct {
	Success bool  `json:"success"`
	Data    Image `json:"data,omitempty"`
}
