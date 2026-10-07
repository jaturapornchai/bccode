package generalledger

import (
	"reflect"
	"testing"
)

func TestNormalizeAccountEntries(t *testing.T) {
	got, err := normalizeAccountEntries([]AccountEntry{{From: " 1110 ", To: "1130"}, {From: "1120", To: "1120"}, {From: "1120"}, {From: "1110", To: "1130"}, {From: "ค่าใช้จ่าย-1"}})
	if err != nil {
		t.Fatal(err)
	}
	want := []AccountEntry{{From: "1110", To: "1130"}, {From: "1120"}, {From: "ค่าใช้จ่าย-1"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("entries = %+v, want %+v", got, want)
	}
	if got, err = normalizeAccountEntries(nil); err != nil || got != nil {
		t.Fatalf("empty = %+v, %v", got, err)
	}
	for name, entries := range map[string][]AccountEntry{
		"reversed":   {{From: "1130", To: "1110"}},
		"no from":    {{From: " ", To: "1130"}},
		"bad to":     {{From: "1110", To: "11 30"}},
		"bad start":  {{From: "-1110"}},
		"second bad": {{From: "1110"}, {From: "11\u200b10"}},
	} {
		if _, err = normalizeAccountEntries(entries); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
	_, err = normalizeAccountEntries([]AccountEntry{{From: "1110"}, {From: "1130", To: "1110"}})
	user, ok := AsUserError(err)
	if !ok || user.Code != "account_group_entry_reversed" || user.Field != "accountentries[1].to" || !reflect.DeepEqual(user.Args, []string{"1130", "1110"}) {
		t.Fatalf("reversed error = %#v", err)
	}
}
