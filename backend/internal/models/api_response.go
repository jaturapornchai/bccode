package models

import "time"

type ApiResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message,omitempty"`
	DocNo      string      `json:"docno,omitempty"`
	ID         interface{} `json:"id,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Pagination interface{} `json:"pagination,omitempty"`
	Total      interface{} `json:"total,omitempty"`
}

type AuthResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
}

type AuthResponseFailed struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ResponseSuccessWithID struct {
	Success bool        `json:"success"`
	ID      interface{} `json:"id,omitempty"`
}

type PaginationDataResponse struct {
	Total     int64 `json:"total"`
	Page      int64 `json:"page"`
	PerPage   int64 `json:"perPage"`
	Prev      int64 `json:"prev"`
	Next      int64 `json:"next"`
	TotalPage int64 `json:"totalPage"`
}

type ResponseSuccess struct {
	Success bool `json:"success"`
}

type BulkInsertResponse struct {
	Success    bool     `json:"success"`
	Created    []string `json:"created"`
	Updated    []string `json:"updated"`
	Failed     []string `json:"updateFailed"`
	Duplicated []string `json:"payloadDuplicate"`
}

type RestErrorResponse struct {
	ErrStatus  int         `json:"status,omitempty"`
	ErrError   string      `json:"error,omitempty"`
	ErrMessage interface{} `json:"message,omitempty"`
	Timestamp  time.Time   `json:"timestamp,omitempty"`
}
