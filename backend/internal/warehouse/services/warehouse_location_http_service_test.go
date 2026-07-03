package services

import "testing"

func TestValidateLocationCompanyScope(t *testing.T) {
	cases := []struct {
		name                 string
		locationCompanyGuids []string
		warehouseGuids       []string
		wantErr              bool
	}{
		{
			name:                 "warehouse unrestricted allows any location scope",
			locationCompanyGuids: []string{"c1", "c2"},
			warehouseGuids:       nil,
			wantErr:              false,
		},
		{
			name:                 "location with no scope always valid (inherits all)",
			locationCompanyGuids: nil,
			warehouseGuids:       []string{"c1"},
			wantErr:              false,
		},
		{
			name:                 "location subset of restricted warehouse is valid",
			locationCompanyGuids: []string{"c1"},
			warehouseGuids:       []string{"c1", "c2"},
			wantErr:              false,
		},
		{
			name:                 "location with guid outside warehouse scope is rejected",
			locationCompanyGuids: []string{"c1", "c3"},
			warehouseGuids:       []string{"c1", "c2"},
			wantErr:              true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateLocationCompanyScope(tc.locationCompanyGuids, tc.warehouseGuids)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
