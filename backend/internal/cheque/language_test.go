package cheque

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ทุก key ข้อความผิดพลาดที่โค้ดของโมดูลเช็ค/ธนาคารใช้ ต้องมีแถวเต็ม 13 คอลัมน์ใน languages.tsv และไม่มีช่องว่าง
// (ไม่มีแถว = ผู้ใช้เห็น key ดิบแทนข้อความ — docs/kms/17-dev-gotchas.md)
var errorKeyPattern = regexp.MustCompile(`"((?:cheque_err|bank_code_err|bank_contact_err)_[a-z0-9_]+)"`)

func loadLanguageRows(t *testing.T) map[string][]string {
	t.Helper()
	file, err := os.Open(filepath.Join("..", "..", "assets", "language", "languages.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	rows := map[string][]string{}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		cells := strings.Split(scanner.Text(), "\t")
		rows[cells[0]] = cells
	}
	if err = scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return rows
}

func TestChequeModuleMessagesHaveLanguageRows(t *testing.T) {
	rows := loadLanguageRows(t)
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	more, _ := filepath.Glob(filepath.Join("httpapi", "*.go"))
	used := map[string]bool{}
	for _, name := range append(files, more...) {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		source, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range errorKeyPattern.FindAllStringSubmatch(string(source), -1) {
			used[m[1]] = true
		}
	}
	if len(used) < 20 {
		t.Fatalf("found only %d message keys in the module sources — pattern is stale", len(used))
	}
	for key := range used {
		cells, ok := rows[key]
		if !ok {
			t.Errorf("%s has no row in languages.tsv", key)
			continue
		}
		if len(cells) != 13 {
			t.Errorf("%s has %d columns, want 13", key, len(cells))
			continue
		}
		for i, cell := range cells {
			if strings.TrimSpace(cell) == "" {
				t.Errorf("%s column %d is empty", key, i)
			}
		}
	}
}
