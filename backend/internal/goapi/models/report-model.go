package models

import (
	"time"
)

type ReportStyleModel struct {
	PaperType string
}

type ReportConditionProductBalanceModel struct {
	EndDateTime time.Time `json:"end_date_time" bson:"end_date_time"`
}

type ReportColumnModel struct {
	Name string  `json:"name" bson:"name"`               // ชื่อคอลัมน์
	Width float64 `json:"width" bson:"width"`             // ความกว้างของคอลัมน์
	WidthCalc float64 `json:"width_calc" bson:"width_calc"`   // ความกว้างของคอลัมน์ที่คำนวณได้
	Align int     `json:"align" bson:"align"`             // การจัดตำแหน่งข้อความ (0=left, 1=center, 2=right)
	ColumnType int     `json:"column_type" bson:"column_type"` // ประเภทของคอลัมน์
	Format string  `json:"format" bson:"format"`           // รูปแบบการแสดงผล เช่น วันที่ ตัวเลข ทศนิยม
	Style int     `json:"style" bson:"style"`             // 0=normal, 1=bold, 2=italic, 3=bold-italic
	TopLine bool    `json:"top_line" bson:"top_line"`       // แสดงเส้นบน
	BottomLine bool    `json:"bottom_line" bson:"bottom_line"` // แสดงเส้นล่าง
}

type ReportRowModel struct {
	LeftMarginPercent float64             `json:"left_margin_percent" bson:"left_margin_percent"` // ระยะห่างด้านซ้ายของแถว
	FontSize float64             `json:"font_size" bson:"font_size"`                     // ขนาดตัวอักษร
	TopLine bool                `json:"top_line" bson:"top_line"`                       // แสดงเส้นบน
	BottomLine bool                `json:"bottom_line" bson:"bottom_line"`                 // แสดงเส้นล่าง
	Columns []ReportColumnModel `json:"columns" bson:"columns"`                         // คอลัมน์ต่างๆ ในแถว
}

type ReportDataModel struct {
	RowIndex int                     `json:"row_index" bson:"row_index"`         // ลำดับของแถว
	TopLine bool                    `json:"top_line" bson:"top_line"`           // แสดงเส้นบน
	BottomLine bool                    `json:"bottom_line" bson:"bottom_line"`     // แสดงเส้นล่าง
	ColumnData []ReportDataColumnModel `json:"next_row_data" bson:"next_row_data"` // ข้อมูลแถวถัดไป
}

type ReportDataColumnModel struct {
	Value any `json:"value" bson:"value"`                     // ข้อมูลคอลัมน์
	DateTimeStyle int `json:"date_time_style" bson:"date_time_style"` // 0=none, 1=short, 2=medium, 3=long
	Point int `json:"point" bson:"point"`                     // จำนวนตำแหน่งทศนิยม
	Style int `json:"style" bson:"style"`                     // 0=normal, 1=bold, 2=italic, 3=bold-italic
}

type ReportModel struct {
	Name string            `json:"name" bson:"name"`               // ชื่อรายงาน
	Header string            `json:"headers" bson:"headers"`         // ส่วนหัวรายงาน
	HeaderRows []ReportRowModel  `json:"header_rows" bson:"header_rows"` // แถวส่วนหัวรายงาน
	FooterRows []ReportRowModel  `json:"footer_rows" bson:"footer_rows"` // แถวส่วนท้ายรายงาน
	DataRows []ReportDataModel `json:"data_rows" bson:"data_rows"`     // แถวข้อมูลรายงาน
}
