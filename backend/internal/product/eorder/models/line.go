package models

type LinePayload struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}
