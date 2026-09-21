package mcptoken

import (
	"fmt"
	"strings"
)

// CompanySelection limits each operation, not the token's company allow-list.
func CompanySelection(mode, single string, many []string) ([]string, error) {
	if mode != "readonly" && mode != "readwrite" {
		return nil, ErrDenied
	}
	if single != "" && len(many) != 0 {
		return nil, fmt.Errorf("ระบุบริษัทแบบเดี่ยวหรือหลายบริษัทอย่างใดอย่างหนึ่ง")
	}
	if len(many) == 0 {
		return []string{strings.TrimSpace(single)}, nil
	}
	if mode == "readwrite" && len(many) != 1 {
		return nil, fmt.Errorf("readwrite เลือกได้ครั้งละหนึ่งบริษัทเท่านั้น")
	}
	if len(many) > 50 {
		return nil, fmt.Errorf("อ่านได้สูงสุด 50 บริษัทต่อคำขอ")
	}
	result := make([]string, 0, len(many))
	seen := map[string]bool{}
	for _, code := range many {
		code = strings.TrimSpace(code)
		if code == "" || seen[code] {
			return nil, fmt.Errorf("รายชื่อบริษัทต้องไม่ว่างหรือซ้ำกัน")
		}
		seen[code] = true
		result = append(result, code)
	}
	return result, nil
}
