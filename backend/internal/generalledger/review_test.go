package generalledger

import "testing"

func TestReviewValidation(t *testing.T) {
	zero := int64(0)
	for _, r := range []*ReviewInput{nil, {Status: 2, ExpectedEventNo: &zero}, {Status: 3}, {Status: 4, ExpectedEventNo: &zero}} {
		if r.validate() == nil {
			t.Fatalf("accepted invalid review: %+v", r)
		}
	}
	for _, status := range []int{1, 2, 3} {
		if err := (&ReviewInput{Status: status, Note: "checked", ExpectedEventNo: &zero}).validate(); err != nil {
			t.Fatal(err)
		}
	}
}
