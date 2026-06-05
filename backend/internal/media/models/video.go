package models

type Media struct {
	Uri       string `json:"uri"`
	Size      int64  `json:"size"`
	Extension string `json:"extension"`
}
