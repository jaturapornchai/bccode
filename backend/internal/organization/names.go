package organization

import (
	"encoding/json"
	"strings"

	common "smlcloudplatform/internal/models"
)

// PrimaryName is the value kept in the relational name column: Thai first, then the
// first non-blank localized name.
func PrimaryName(names common.JSONB) string {
	first := ""
	for _, name := range names {
		if name.Code == nil || name.Name == nil || strings.TrimSpace(*name.Name) == "" {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(*name.Code), "th") {
			return strings.TrimSpace(*name.Name)
		}
		if first == "" {
			first = strings.TrimSpace(*name.Name)
		}
	}
	return first
}

// DecodeNames reads a names JSONB column; rows without localized names expose the
// relational name as Thai.
func DecodeNames(raw []byte, name string) common.JSONB {
	var names common.JSONB
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &names)
	}
	if len(names) == 0 && strings.TrimSpace(name) != "" {
		code, value := "th", name
		names = common.JSONB{{Code: &code, Name: &value}}
	}
	if names == nil {
		names = common.JSONB{}
	}
	return names
}

// EncodeNames writes localized names for a JSONB column.
func EncodeNames(names common.JSONB) (string, error) {
	if names == nil {
		names = common.JSONB{}
	}
	raw, err := json.Marshal(names)
	return string(raw), err
}
