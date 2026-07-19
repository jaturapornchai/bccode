package models

import (
	"encoding/json"
	"testing"

	common "smlcloudplatform/internal/models"

	"github.com/go-playground/validator/v10"
)

func TestCodeXSortJSONIsProductOnly(t *testing.T) {
	payload, err := json.Marshal(CodeXSort{Code: "P001", XOrder: 2})
	if err != nil {
		t.Fatal(err)
	}

	var fields map[string]any
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"barcode", "unitcode", "unitnames", "manufacturerguid"} {
		if _, exists := fields[forbidden]; exists {
			t.Fatalf("CodeXSort must not expose %q", forbidden)
		}
	}
}

func TestProductCategoryCodeListValidation(t *testing.T) {
	nameCode, name := "th", "หมวดทดสอบ"
	names := []common.NameX{{Code: &nameCode, Name: &name}}
	xsorts := []common.XSort{}

	tests := []struct {
		name     string
		codeList []CodeXSort
		wantErr  bool
	}{
		{name: "distinct product codes", codeList: []CodeXSort{{Code: "P001"}, {Code: "P002"}}},
		{name: "duplicate product code", codeList: []CodeXSort{{Code: "P001"}, {Code: "P001"}}, wantErr: true},
		{name: "empty product code", codeList: []CodeXSort{{Code: ""}}, wantErr: true},
	}

	validate := validator.New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := ProductCategory{Names: &names, XSorts: &xsorts, CodeList: &tt.codeList}
			err := validate.Struct(doc)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
