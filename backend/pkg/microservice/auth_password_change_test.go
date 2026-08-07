package microservice

import (
	"net/http"
	"testing"
)

func TestPasswordChangeAllowed(t *testing.T) {
	tests := []struct {
		method  string
		path    string
		allowed bool
	}{
		{http.MethodPut, "/profile/password", true},
		{http.MethodPost, "/logout", true},
		{http.MethodGet, "/profile", true},
		{http.MethodPut, "/profile", false},
		{http.MethodGet, "/list-holding", false},
		{http.MethodPost, "/goapi/document", false},
	}

	for _, test := range tests {
		if got := passwordChangeAllowed(test.method, test.path); got != test.allowed {
			t.Fatalf("passwordChangeAllowed(%q, %q) = %v, want %v", test.method, test.path, got, test.allowed)
		}
	}
}
