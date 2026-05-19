package handlers

import "testing"

func TestFormatAmountUsesThousandsSeparator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		amount float64
		want   string
	}{
		{name: "zero", amount: 0, want: "0.00"},
		{name: "positive", amount: 1234567.8, want: "1,234,567.80"},
		{name: "negative", amount: -9876.5, want: "-9,876.50"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := formatAmount(tt.amount); got != tt.want {
				t.Fatalf("formatAmount(%v) = %q, want %q", tt.amount, got, tt.want)
			}
		})
	}
}
