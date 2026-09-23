package models

type Pagination struct {
	Total     int `json:"total"`
	Page      int `json:"page"`
	PerPage   int `json:"perpage"`
	TotalPage int `json:"totalpage"`
}

// PaginationData — ข้อมูลแบ่งหน้าที่ส่งกลับใน API (รูป JSON เดิม: total/page/perPage/prev/next/totalPage)
type PaginationData struct {
	Total     int64 `json:"total"`
	Page      int64 `json:"page"`
	PerPage   int64 `json:"perPage"`
	Prev      int64 `json:"prev"`
	Next      int64 `json:"next"`
	TotalPage int64 `json:"totalPage"`
}

// SinglePage — ผลลัพธ์ที่คืนครบทั้งชุดในหน้าเดียว
func SinglePage(total int) PaginationData {
	perPage := int64(total)
	if perPage < 1 {
		perPage = 1
	}
	return PaginationData{Total: int64(total), Page: 1, PerPage: perPage, TotalPage: 1}
}
