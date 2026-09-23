package generalledger

import (
	"errors"
	"net/http"

	"smlcloudplatform/pkg/apperr"
)

// Stable machine-readable error codes for POST /gl/v2/command.
// The UI matches on these codes; the Thai text always travels in the JSON
// field `message` so the screen can render it without translating.
const (
	CodeDuplicateCode    = "duplicate_code"
	CodeDuplicateRequest = "duplicate_request"
	CodeValidationFailed = "validation_failed"
	CodeStaleVersion     = "stale_version"
	CodeParentNotFound   = "parent_not_found"
	CodeParentInvalid    = "parent_invalid"
	CodeTreeInvalid      = "account_tree_invalid"
	CodeLevelRange       = "level_out_of_range"
	CodeLevelShallow     = "level_not_deeper_than_parent"
	CodeHasChildren      = "account_has_children"
	CodeReferenced       = "account_referenced"
	CodeReferencedMaster = "account_referenced_master"
	CodePostedLocked     = "account_posted_immutable"
	CodeImmutableCode    = "code_immutable"
	CodeGroupNotFound    = "account_group_not_found"
	CodePayloadMissing   = "account_payload_required"
	CodeUnsupported      = "unsupported_command"
)

// UserError is a command failure caused by user input or by the current state of
// the chart of accounts (a guard), i.e. something the UI must explain in Thai.
// Error() returns the Thai message unchanged, so existing logs and tests that
// match on the Thai text keep working.
type UserError struct {
	Code    string
	Message string // Thai, user facing
	Status  int    // HTTP status; 0 = 409 Conflict
}

func (e *UserError) Error() string { return e.Message }

func (e *UserError) HTTPStatus() int {
	if e.Status == 0 {
		return http.StatusConflict
	}
	return e.Status
}

func (e *UserError) ToAppError() *apperr.AppError {
	return &apperr.AppError{
		Code:       e.Code,
		Message:    e.Message,
		ThaiMsg:    e.Message,
		HTTPStatus: e.HTTPStatus(),
	}
}

// AsUserError finds a user-caused failure inside err (also through wrapped errors).
func AsUserError(err error) (*UserError, bool) {
	var user *UserError
	if errors.As(err, &user) {
		return user, true
	}
	return nil, false
}

// userError builds a command failure with its machine code; message keeps the
// exact Thai text the guard already produced.
func userError(code string, message string) *UserError {
	return &UserError{Code: code, Message: message}
}

// validationFailed keeps the Thai text of a model validation error and tags it
// with the stable validation_failed code.
func validationFailed(err error) error {
	if err == nil {
		return nil
	}
	if user, ok := AsUserError(err); ok {
		return user
	}
	return userError(CodeValidationFailed, err.Error())
}
