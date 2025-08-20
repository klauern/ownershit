package ownershit

import (
	"strings"
	"testing"
)

func TestGetRequiredTokenPermissions(t *testing.T) {
	permissions := GetRequiredTokenPermissions()

	// Test that all expected categories are present
	expectedCategories := []string{
		"classic_token_scopes",
		"fine_grained_permissions",
		"operations_requiring_permissions",
	}

	for _, category := range expectedCategories {
		if _, exists := permissions[category]; !exists {
			t.Errorf("GetRequiredTokenPermissions() missing category: %s", category)
		}
	}

	// Test classic token scopes
	classicScopes := permissions["classic_token_scopes"]
	expectedScopes := []string{"repo", "admin:org", "read:org", "user"}

	for _, expectedScope := range expectedScopes {
		found := false
		for _, scope := range classicScopes {
			if scope == expectedScope {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetRequiredTokenPermissions() missing classic scope: %s", expectedScope)
		}
	}

	// Test that fine-grained permissions contain key terms
	fineGrainedPerms := permissions["fine_grained_permissions"]
	expectedTerms := []string{"Repository permissions:", "Organization permissions:", "Administration", "Metadata"}

	permText := strings.Join(fineGrainedPerms, " ")
	for _, term := range expectedTerms {
		if !strings.Contains(permText, term) {
			t.Errorf("GetRequiredTokenPermissions() fine-grained permissions missing term: %s", term)
		}
	}

	// Test operations
	operations := permissions["operations_requiring_permissions"]
	if len(operations) == 0 {
		t.Error("GetRequiredTokenPermissions() operations list is empty")
	}
}
