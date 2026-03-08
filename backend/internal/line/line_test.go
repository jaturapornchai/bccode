package line

import (
	"testing"
)

func TestValidateToken(t *testing.T) {
	adapter := NewLineAdapter("1657004770")

	token := "eyJraWQiOiI3ZjMxMTU5YTY1YWE0YmYxZGRmMzQyYjU3MTcwZGQ3NDY3ZTkyZDEyZTg0YzI0Y2EyMGUxNDQyNTUzYjNmMDhjIiwidHlwIjoiSldUIiwiYWxnIjoiRVMyNTYifQ.eyJpc3MiOiJodHRwczovL2FjY2Vzcy5saW5lLm1lIiwic3ViIjoiVTIxODhiN2E1YzNjMTM1MDBkYTczMmRlMGZkMTMyODM1IiwiYXVkIjoiMTY1NzAwNDc3MCIsImV4cCI6MTc2NzUxOTI0MCwiaWF0IjoxNzY3NTE1NjQwLCJhbXIiOlsibGluZXNzbyJdLCJuYW1lIjoic3V0ZWV0b2UiLCJwaWN0dXJlIjoiaHR0cHM6Ly9wcm9maWxlLmxpbmUtc2Nkbi5uZXQvMGh0a2tfM1ZxdUswUldKamdialI5VUUycGpKU2toQ0MwTUxrUm1KeUJ5Sm5ZcEh6bENZMFZpSUhZZ0ppUjRFMmxDUDBWdEppQXVKU0FyIn0.kYNsFSWRj_nlO65TnGXMklqBarxYXeIf8J8eTYfqySy0Mg1D3YskE2g_YaF7FTMhDhFi2xhdyZvhGEWkD-WbAA"

	user, err := adapter.ValidateToken(token)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedUserId := "U2188b7a5c3c13500da732de0fd132835"
	if user.UserId != expectedUserId {
		t.Errorf("Expected UserId %s, got %s", expectedUserId, user.UserId)
	}

	expectedDisplayName := "suteetoe"
	if user.DisplayName != expectedDisplayName {
		t.Errorf("Expected DisplayName %s, got %s", expectedDisplayName, user.DisplayName)
	}

	expectedPictureUrl := "https://profile.line-scdn.net/0htkk_3VquK0RWJjgbjR9UE2pjJSkhCC0MLkRmJyByJnYpHzlCY0ViIHYgJiR4E2lCP0VtJiAuJSAr"
	if user.PictureUrl != expectedPictureUrl {
		t.Errorf("Expected PictureUrl %s, got %s", expectedPictureUrl, user.PictureUrl)
	}
}
