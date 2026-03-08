package models

type Pagination struct {
	Total     int `json:"total"`
	Page      int `json:"page"`
	PerPage   int `json:"perPage"`
	TotalPage int `json:"totalPage"`
}
