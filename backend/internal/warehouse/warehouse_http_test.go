package warehouse

import (
	"testing"

	warehouseModels "smlcloudplatform/internal/warehouse/models"
)

func TestValidateLocationCompanyScope(t *testing.T) {
	cases := []struct {
		name           string
		locations      []warehouseModels.Location
		warehouseGuids []string
		wantErr        bool
	}{
		{
			name:           "warehouse unrestricted allows any location scope",
			locations:      []warehouseModels.Location{{Code: "Z1", CompanyGuids: []string{"c1", "c2"}}},
			warehouseGuids: nil,
			wantErr:        false,
		},
		{
			name:           "location with no scope always valid (inherits all)",
			locations:      []warehouseModels.Location{{Code: "Z1"}},
			warehouseGuids: []string{"c1"},
			wantErr:        false,
		},
		{
			name:           "location subset of restricted warehouse is valid",
			locations:      []warehouseModels.Location{{Code: "Z1", CompanyGuids: []string{"c1"}}},
			warehouseGuids: []string{"c1", "c2"},
			wantErr:        false,
		},
		{
			name:           "location with guid outside warehouse scope is rejected",
			locations:      []warehouseModels.Location{{Code: "Z1", CompanyGuids: []string{"c1", "c3"}}},
			warehouseGuids: []string{"c1", "c2"},
			wantErr:        true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateLocationCompanyScope(tc.locations, tc.warehouseGuids)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
