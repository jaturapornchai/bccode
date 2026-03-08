package round_test

import (
	"smlcloudplatform/pkg/round"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoundUp(t *testing.T) {

	cases := []struct {
		num  float64
		want float64
	}{
		{
			num:  1.234,
			want: 1.23,
		},
		{
			num:  1.235,
			want: 1.24,
		},
	}

	for _, c := range cases {
		got := round.Round(c.num, 2)
		if got != c.want {
			t.Errorf("Round(%v) == %v, want %v", c.num, got, c.want)
		}
	}
}

func TestSubtract(t *testing.T) {

	amount := 1011.92

	amount -= 100.2

	got := round.Round(amount, 2)
	want := 911.72

	assert.Equal(t, want, got)
}

func TestRoundFloatIEEE754(t *testing.T) {

	balanceFirst := 2090.37
	sumOfCost := 6.39 * 60

	got := round.Round(sumOfCost, 2)
	assert.Equal(t, 383.4, got)

	balanceAmount := balanceFirst - got
	balanceAmount = round.Round(balanceAmount, 2)
	assert.Equal(t, 1706.97, balanceAmount)
}
