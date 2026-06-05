package ocr

import "io"

type FileContent struct {
	FileName string
	Content  io.Reader
}

type OcrUpload struct {
	TrackingID string `json:"trackingid"`
	FormIndex  uint   `json:"formindex"`
}

type OcrResault struct {
	TrackingID    string `json:"trackingid"`
	Type          string `json:"type"`
	Url           uint   `json:"url"`
	RawHeader     uint   `json:"rawheader"`
	Confident     uint   `json:"confident"`
	SignatureCode uint   `json:"signaturecode"`
	Startdate     string `json:"startdate"`
	Stopdate      string `json:"stopdate"`
}

type OcrRequest struct {
	ResourceKey  string   `json:"resourcekey"`
	UrlResources []string `json:"urlresources"`
}
