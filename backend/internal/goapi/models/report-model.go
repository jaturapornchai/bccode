package models

import (
	"time"
)

type ReportStyleModel struct {
	PaperType string
}

type ReportConditionProductBalanceModel struct {
	EndDateTime time.Time `json:"enddatetime" bson:"enddatetime"`
}

type ReportColumnModel struct {
	Name       string  `json:"name" bson:"name"`             // ชื่อคอลัมน์
	Width      float64 `json:"width" bson:"width"`           // ความกว้างของคอลัมน์
	WidthCalc  float64 `json:"widthcalc" bson:"widthcalc"`   // ความกว้างของคอลัมน์ที่คำนวณได้
	Align      int     `json:"align" bson:"align"`           // การจัดตำแหน่งข้อความ (0=left, 1=center, 2=right)
	ColumnType int     `json:"columntype" bson:"columntype"` // ประเภทของคอลัมน์
	Format     string  `json:"format" bson:"format"`         // รูปแบบการแสดงผล เช่น วันที่ ตัวเลข ทศนิยม
	Style      int     `json:"style" bson:"style"`           // 0=normal, 1=bold, 2=italic, 3=bold-italic
	TopLine    bool    `json:"topline" bson:"topline"`       // แสดงเส้นบน
	BottomLine bool    `json:"bottomline" bson:"bottomline"` // แสดงเส้นล่าง
}

type ReportRowModel struct {
	LeftMarginPercent float64             `json:"leftmarginpercent" bson:"leftmarginpercent"` // ระยะห่างด้านซ้ายของแถว
	FontSize          float64             `json:"fontsize" bson:"fontsize"`                   // ขนาดตัวอักษร
	TopLine           bool                `json:"topline" bson:"topline"`                     // แสดงเส้นบน
	BottomLine        bool                `json:"bottomline" bson:"bottomline"`               // แสดงเส้นล่าง
	Columns           []ReportColumnModel `json:"columns" bson:"columns"`                     // คอลัมน์ต่างๆ ในแถว
}

type ReportDataModel struct {
	RowIndex   int                     `json:"rowindex" bson:"rowindex"`       // ลำดับของแถว
	TopLine    bool                    `json:"topline" bson:"topline"`         // แสดงเส้นบน
	BottomLine bool                    `json:"bottomline" bson:"bottomline"`   // แสดงเส้นล่าง
	ColumnData []ReportDataColumnModel `json:"nextrowdata" bson:"nextrowdata"` // ข้อมูลแถวถัดไป
}

type ReportDataColumnModel struct {
	Value         any `json:"value" bson:"value"`                 // ข้อมูลคอลัมน์
	DateTimeStyle int `json:"datetimestyle" bson:"datetimestyle"` // 0=none, 1=short, 2=medium, 3=long
	Point         int `json:"point" bson:"point"`                 // จำนวนตำแหน่งทศนิยม
	Style         int `json:"style" bson:"style"`                 // 0=normal, 1=bold, 2=italic, 3=bold-italic
}

type ReportModel struct {
	Name       string            `json:"name" bson:"name"`             // ชื่อรายงาน
	Header     string            `json:"headers" bson:"headers"`       // ส่วนหัวรายงาน
	HeaderRows []ReportRowModel  `json:"headerrows" bson:"headerrows"` // แถวส่วนหัวรายงาน
	FooterRows []ReportRowModel  `json:"footerrows" bson:"footerrows"` // แถวส่วนท้ายรายงาน
	DataRows   []ReportDataModel `json:"datarows" bson:"datarows"`     // แถวข้อมูลรายงาน
}
