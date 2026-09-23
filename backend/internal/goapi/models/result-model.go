package models

type ResultModel struct {
	Success     bool   `json:"success"`
	Guid        string `json:"guid"`
	TotalRecord int    `json:"totalrecord"`
	TotalPage   int    `json:"totalpage"`
}
