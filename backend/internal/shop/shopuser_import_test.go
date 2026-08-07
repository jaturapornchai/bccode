package shop

import "testing"

func TestParseImportUsersAllowsUserCodeWithoutEmail(t *testing.T) {
	rows, err := parseImportUsers("users.csv", []byte("usercode,username,email,role\nsomchai01,สมชาย,,user\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || !rows[0].Valid || rows[0].Username != "somchai01" || rows[0].Email != "" {
		t.Fatalf("unexpected rows: %#v", rows)
	}
}

func TestParseImportUsersRejectsDuplicateEmail(t *testing.T) {
	rows, err := parseImportUsers("users.csv", []byte("usercode,username,email\nu01,หนึ่ง,same@example.com\nu02,สอง,same@example.com\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[1].Valid || rows[1].Message != "อีเมลซ้ำในไฟล์" {
		t.Fatalf("unexpected rows: %#v", rows)
	}
}
