package ocr

import "io"

type FileContent struct {
	FileName string
	Content  io.Reader
}

type OcrUpload struct {
	TrackingID string `json:"tracking_id"`
	FormIndex  uint   `json:"form_index"`
}

type OcrResault struct {
	TrackingID    string `json:"tracking_id"`
	Type          string `json:"type"`
	Url           uint   `json:"url"`
	RawHeader     uint   `json:"raw_header"`
	Confident     uint   `json:"confident"`
	SignatureCode uint   `json:"signature_code"`
	Startdate     string `json:"startdate"`
	Stopdate      string `json:"stopdate"`
}

type OcrRequest struct {
	ResourceKey  string   `json:"resourcekey"`
	UrlResources []string `json:"urlresources"`
}
