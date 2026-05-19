package language

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func resetLanguageCacheForTest(t *testing.T) {
	t.Helper()

	cacheLock.Lock()
	previousCache := cache
	previousLoaded := loaded
	previousPath := loadedPath
	previousModTime := loadedModTime
	previousSize := loadedSize
	cache = nil
	loaded = false
	loadedPath = ""
	loadedModTime = time.Time{}
	loadedSize = 0
	cacheLock.Unlock()

	t.Cleanup(func() {
		cacheLock.Lock()
		cache = previousCache
		loaded = previousLoaded
		loadedPath = previousPath
		loadedModTime = previousModTime
		loadedSize = previousSize
		cacheLock.Unlock()
	})
}

func TestNormalizeLanguageAliases(t *testing.T) {
	cases := map[string]string{
		"ai:ko": "ko",
		"de-DE": "th",
		"jp":    "ja",
		"kr":    "ko",
		"tl-PH": "fil",
		"zh-CN": "cn",
	}

	for input, expected := range cases {
		if got := Normalize(input); got != expected {
			t.Fatalf("Normalize(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestDictionaryLoadsSharedLanguageFile(t *testing.T) {
	dict, err := Dictionary("ko")
	if err != nil {
		t.Fatalf("Dictionary returned error: %v", err)
	}

	if got := dict["select_company"]; got != "회사 선택" {
		t.Fatalf("select_company ko = %q", got)
	}
	if got := dict["report_product_stock_movement"]; got != "상품 재고 이동 보고서" {
		t.Fatalf("report_product_stock_movement ko = %q", got)
	}
}

func TestDictionaryReloadsMountedLanguageFileWhenEnabled(t *testing.T) {
	resetLanguageCacheForTest(t)

	path := filepath.Join(t.TempDir(), "languages.tsv")
	if err := os.WriteFile(path, []byte("key\tth\ten\nhello\tเก่า\tOld\n"), 0o600); err != nil {
		t.Fatalf("write initial language file: %v", err)
	}
	t.Setenv("LANGUAGE_FILE_PATH", path)
	t.Setenv("LANGUAGE_RELOAD_ENABLED", "true")

	dict, err := Dictionary("en")
	if err != nil {
		t.Fatalf("Dictionary initial load returned error: %v", err)
	}
	if got := dict["hello"]; got != "Old" {
		t.Fatalf("initial hello = %q, want Old", got)
	}

	if err := os.WriteFile(path, []byte("key\tth\ten\nhello\tใหม่กว่า\tNew Value\n"), 0o600); err != nil {
		t.Fatalf("write updated language file: %v", err)
	}
	future := time.Now().Add(time.Second)
	if err := os.Chtimes(path, future, future); err != nil {
		t.Fatalf("touch updated language file: %v", err)
	}

	dict, err = Dictionary("en")
	if err != nil {
		t.Fatalf("Dictionary reload returned error: %v", err)
	}
	if got := dict["hello"]; got != "New Value" {
		t.Fatalf("reloaded hello = %q, want New Value", got)
	}
}

func TestDictionaryFallsBackToKeyForMissingSelectedLanguage(t *testing.T) {
	resetLanguageCacheForTest(t)

	cacheLock.Lock()
	cache = map[string]map[string]string{
		"only_english": {"en": "Only English", "th": "มีเฉพาะไทย"},
	}
	loaded = true
	cacheLock.Unlock()

	dict, err := Dictionary("lo")
	if err != nil {
		t.Fatalf("Dictionary returned error: %v", err)
	}
	if got := dict["only_english"]; got != "only_english" {
		t.Fatalf("missing lo cell = %q, want key fallback", got)
	}
}

func TestTextFallsBackToKeyForMissingKey(t *testing.T) {
	if got := Text("missing_key", "ko"); got != "missing_key" {
		t.Fatalf("missing key = %q", got)
	}
}
