package textguard

import "testing"

type nulName struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type nulRequest struct {
	Username string            `json:"username"`
	Avatar   *string           `json:"avatar,omitempty"`
	Names    []nulName         `json:"names"`
	Extra    map[string]string `json:"extra"`
	Hidden   string            `json:"-"`
	internal string
}

func TestNULField(t *testing.T) {
	bad := "ฝ่าย\x00บัญชี"
	for want, req := range map[string]nulRequest{
		"":              {Username: "somchai01", Names: []nulName{{"th", "บัญชี"}}, Hidden: bad, internal: bad},
		"username":      {Username: bad},
		"avatar":        {Avatar: &bad},
		"names[1].name": {Names: []nulName{{"th", "บัญชี"}, {"en", bad}}},
		"extra.th":      {Extra: map[string]string{"th": bad}},
		"extra":         {Extra: map[string]string{bad: "x"}},
	} {
		if got := NULField(&req); got != want {
			t.Errorf("NULField = %q, want %q", got, want)
		}
	}
	if got := LastSegment("names[1].name"); got != "name" {
		t.Fatalf("LastSegment = %q", got)
	}
	if got := LastSegment("permissionsets[2]"); got != "permissionsets" {
		t.Fatalf("LastSegment = %q", got)
	}
}

func TestMessageNamesTheField(t *testing.T) {
	text := func(key string) string {
		return map[string]string{"text_contains_nul_field": "ช่อง “{0}” มี NUL", "text_contains_nul": "มี NUL", "user_position": "ตำแหน่ง"}[key]
	}
	labels := map[string]string{"position": "user_position"}
	if got := Message("position", labels, text); got != "ช่อง “ตำแหน่ง” มี NUL" {
		t.Fatalf("labelled = %q", got)
	}
	if got := Message("lineuserid", labels, text); got != "มี NUL" {
		t.Fatalf("unlabelled = %q", got)
	}
}
