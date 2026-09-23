package models

import (
	"encoding/json"
	"errors"
	"strings"
)

// RoleFromText maps the central holding_members.role column to UserRole. Any role other
// than OWNER/ADMIN is an ordinary member whose rights come from role_permissions.
func RoleFromText(role string) UserRole {
	switch strings.ToUpper(strings.TrimSpace(role)) {
	case "OWNER":
		return ROLE_OWNER
	case "ADMIN":
		return ROLE_ADMIN
	}
	return ROLE_USER
}

// RoleText is the value stored in holding_members.role.
func RoleText(role UserRole) string {
	switch role {
	case ROLE_OWNER:
		return "OWNER"
	case ROLE_ADMIN:
		return "ADMIN"
	}
	return "USER"
}

// ParseAccessScopes decodes holding_members.access_scopes. The central schema stores an
// empty grant as {} (or null); anything else must be a JSON array of scopes.
func ParseAccessScopes(raw []byte) ([]AccessScope, error) {
	text := strings.TrimSpace(string(raw))
	if text == "" || text == "{}" || text == "null" {
		return nil, nil
	}
	var scopes []AccessScope
	if err := json.Unmarshal(raw, &scopes); err != nil {
		return nil, errors.New("invalid access scopes")
	}
	return scopes, nil
}
