package organization

import (
	"encoding/json"
	"errors"
	"strings"
)

var ErrStatusReasonRequired = errors.New("status change reason is required")

func ResolveRequestedActiveStatus(input string, current bool) (active bool, changed bool, err error) {
	var request struct {
		IsActive *bool `json:"isactive"`
	}
	if err := json.Unmarshal([]byte(input), &request); err != nil {
		return false, false, err
	}
	if request.IsActive == nil {
		return current, false, nil
	}
	return *request.IsActive, *request.IsActive != current, nil
}

func StatusChangeReason(input string) (string, error) {
	var request struct {
		Reason string `json:"statusreason"`
	}
	if err := json.Unmarshal([]byte(input), &request); err != nil {
		return "", err
	}
	reason := strings.TrimSpace(request.Reason)
	if reason == "" {
		return "", ErrStatusReasonRequired
	}
	return reason, nil
}
