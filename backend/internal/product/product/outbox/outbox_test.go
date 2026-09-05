package outbox

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMessagePreservesExactJSON(t *testing.T) {
	amount := "9007199254740993.123456789"
	message, err := NewMessage("when-product-updated", "scope", map[string]string{"amount": amount})
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	var restored Message
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if string(restored.Payload) != string(message.Payload) || !strings.Contains(string(data), amount) {
		t.Fatal("exact value changed")
	}
	if _, err := NewMessage("", "scope", nil); err == nil {
		t.Fatal("blank topic accepted")
	}
	if _, err := NewMessage("topic", "", nil); err == nil {
		t.Fatal("blank key accepted")
	}
	if _, err := NewMessage("topic", "key", make(chan int)); err == nil {
		t.Fatal("unsupported payload accepted")
	}
}

func TestAggregateKeySeparatesCompanyAndIdentity(t *testing.T) {
	a := AggregateKey("holding", "company-a", "product")
	if a == AggregateKey("holding", "company-b", "product") {
		t.Fatal("company collision")
	}
	if AggregateKey("a/b", "c", "d") == AggregateKey("a", "b/c", "d") {
		t.Fatal("separator collision")
	}
	if a != AggregateKey("holding", "company-a", "product") {
		t.Fatal("unstable key")
	}
}
