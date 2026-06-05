package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"smlcloudplatform/internal/goapi/logger"
)

type languageNameModel struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CheckSumModel struct {
	RefCode     string `json:"refcode"`
	Checksum    string `json:"checksum"`
	MongoDbName string `json:"mongodbname"`
}

type productOrderModel struct {
	HoldingCode     string                     `json:"holdingcode"`
	Barcode         string                     `json:"barcode"`
	Itemcode        string                     `json:"itemcode"`
	Names           []languageNameModel        `json:"names"`
	Unituses        []productOrderUnitUseModel `json:"unituses"`
	Unitcode        string                     `json:"unitcode"`
	Units           []productOrderUnitModel    `json:"units"`
	Unitcost        string                     `json:"unitcost"`
	Unitstandard    string                     `json:"unitstandard"`
	Multiunit       bool                       `json:"multiunit"`
	Itemtype        int                        `json:"itemtype"`
	Itemvat         int                        `json:"itemvat"`
	Normalprice     float64                    `json:"normalprice"`
	Price           float64                    `json:"price"`
	Memberprice     float64                    `json:"memberprice"`
	Pricerangemin   float64                    `json:"pricerangemin"`
	Pricerangemax   float64                    `json:"pricerangemax"`
	Images          []productOrderImageModel   `json:"images"`
	Recommended     bool                       `json:"recommended"`
	Shoprecommended bool                       `json:"shoprecommended"`
	Havepoint       bool                       `json:"havepoint"`
	Starpersent     float64                    `json:"starpersent"`
	Ordercount      int                        `json:"ordercount"`
	Descriptions    []languageNameModel        `json:"descriptions"`
	Options         []productOrderOptionModel  `json:"options"`
	Orderminimum    float64                    `json:"orderminimum"`
}

type productOrderUnitModel struct {
	Unitcode  string              `json:"unitcode"`
	Unitnames []languageNameModel `json:"unitnames"`
}

type productOrderOptionModel struct {
	Guidcode      string                          `json:"guidcode"`
	Names         []languageNameModel             `json:"names"`
	Isstock       bool                            `json:"isstock"`
	Optiondetails []productOrderOptionDetailModel `json:"optiondetails"`
}

type productOrderOptionDetailIncludeModel struct {
	Optionguid string                                 `json:"optionguid"`
	Details    []productOrderOptionDetailIncludeModel `json:"details"`
}

type productOrderOptionDetailModel struct {
	Guidcode       string                                 `json:"guidcode"`
	Names          []languageNameModel                    `json:"names"`
	Image          string                                 `json:"image"`
	Includeoptions []productOrderOptionDetailIncludeModel `json:"includeoptions"`
	Selected       bool                                   `json:"-"`
	Isenable       bool                                   `json:"-"`
}

type productOrderUnitUseModel struct {
	Unitcode    string  `json:"unitcode"`
	Itemunitstd float64 `json:"itemunitstd"`
	Itemunitdiv float64 `json:"itemunitdiv"`
	Isunitcost  bool    `json:"isunitcost"`
}

type productOrderImageModel struct {
	Uri string `json:"uri"`
}

type productOrderBalanceModel struct {
	Itemcode string                           `json:"itemcode"`
	Qty      float64                          `json:"qty"`
	Units    []productOrderBalanceUnitModel   `json:"units"`
	Options  []productOrderBalanceOptionModel `json:"options"`
}

type productOrderBalanceUnitModel struct {
	Unitcode string  `json:"unitcode"`
	Qty      float64 `json:"qty"`
}

type productOrderBalanceOptionModel struct {
	Optionguid string                                 `json:"optionguid"`
	Details    []productOrderBalanceOptionDetailModel `json:"details"`
}

type productOrderBalanceOptionDetailModel struct {
	Optionguid string  `json:"optionguid"`
	Qty        float64 `json:"qty"`
}

func DataProductForTest() productOrderModel {
	return productOrderModel{
		HoldingCode: "holdingcode",
		Barcode:     "barcode",
		Itemcode:    "A001X",
		Names: []languageNameModel{
			{
				Code: "th",
				Name: "เสื้อยิดคอกลม ADIDAS รุ่นหล่อแน่นอน ...",
			},
		},
		Unitcost:        "U01",
		Unitstandard:    "U01",
		Multiunit:       true,
		Itemtype:        1,
		Itemvat:         1,
		Normalprice:     150,
		Pricerangemin:   20,
		Pricerangemax:   50,
		Price:           111,
		Memberprice:     110,
		Orderminimum:    1,
		Shoprecommended: true,
		Descriptions: []languageNameModel{
			{
				Code: "th",
				Name: "<html>adlhjkalkadklaklsd adjkladl<br/> ...",
			},
		},
		Images: []productOrderImageModel{
			{Uri: "https://ideakidshop.com/sites/..."},
			{Uri: "https://hm-media-prod.s3.amazonaws.com/..."},
		},
		Starpersent: 70,
		Ordercount:  232322,
		Recommended: true,
		Havepoint:   true,
		Options: []productOrderOptionModel{
			{
				Guidcode: "X1",
				Names: []languageNameModel{
					{Code: "th", Name: "สี"},
				},
				Isstock: true,
				Optiondetails: []productOrderOptionDetailModel{
					{
						Guidcode: "G12",
						Names: []languageNameModel{
							{Code: "th", Name: "ขาว"},
						},
						Image: "",
						Includeoptions: []productOrderOptionDetailIncludeModel{
							{
								Optionguid: "X2",
								Details: []productOrderOptionDetailIncludeModel{
									{
										Optionguid: "S11",
										Details:    []productOrderOptionDetailIncludeModel{{Optionguid: "S111"}},
									},
									{
										Optionguid: "S12",
										Details:    []productOrderOptionDetailIncludeModel{{Optionguid: "S112"}},
									},
								},
							},
						},
					},
				},
			},
		},
		Unitcode: "U01",
		Units: []productOrderUnitModel{
			{
				Unitcode: "U01",
				Unitnames: []languageNameModel{
					{Code: "th", Name: "ตัว"},
					{Code: "en", Name: "Piece"},
					// ... [other LanguageNameModel instances]
				},
			},
		},
		Unituses: []productOrderUnitUseModel{
			{
				Unitcode:    "U01",
				Itemunitstd: 1.0,
				Itemunitdiv: 1.0,
				Isunitcost:  true,
			},
			{
				Unitcode:    "U02",
				Itemunitstd: 12.0,
				Itemunitdiv: 1.0,
				Isunitcost:  false,
			},
		},
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	var product = DataProductForTest()

	switch r.Method {
	case "GET":
		j, _ := json.Marshal(product)
		w.Write(j)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "I can't do that.")
	}
}
func main() {
	http.HandleFunc("/product", handler)

	logger.Info("Go!")
	http.ListenAndServe(":8086", nil)
}

// Warehouse models for Kafka messages
type MongoWarehouseModel struct {
	HoldingCode string                        `json:"holdingcode"`
	Code        string                        `json:"code"`
	Names       []languageNameModel           `json:"names"`
	Location    []MongoWarehouseLocationModel `json:"location"`
}

type MongoWarehouseLocationModel struct {
	Code  string              `json:"code"`
	Names []languageNameModel `json:"names"`
}
