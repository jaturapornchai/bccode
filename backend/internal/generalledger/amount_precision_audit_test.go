package generalledger

import (
	"encoding/json"
	"testing"
)

func TestAmountSmallestSupportedJSONRoundTrip(t *testing.T) {
	for _, value := range []string{"0.00000001", "0.0000001", "0.000001", "-0.00000001"} {
		t.Run(value, func(t *testing.T) {
			amount, err := ParseAmount(value)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(struct {
				Amount Amount `json:"amount"`
			}{amount})
			if err != nil {
				t.Fatal(err)
			}
			var restored struct {
				Amount Amount `json:"amount"`
			}
			if err := json.Unmarshal(raw, &restored); err != nil || restored.Amount != amount {
				t.Fatalf("JSON %s round-trip: got %s; %v", raw, restored.Amount, err)
			}
		})
	}
}
