package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestPermissionsCommand(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the permissions command
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:   "permissions",
				Action: permissionsCommand,
			},
		},
	}

	err := app.Run([]string{"app", "permissions"})
	if err != nil {
		t.Fatalf("permissionsCommand failed: %v", err)
	}

	// Restore stdout and capture output
	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Test that key sections are present in output
	expectedSections := []string{
		"GitHub Token Permissions Required for ownershit",
		"CLASSIC PERSONAL ACCESS TOKEN SCOPES:",
		"FINE-GRAINED PERSONAL ACCESS TOKEN PERMISSIONS:",
		"PERMISSION REQUIREMENTS BY OPERATION:",
		"SETUP INSTRUCTIONS:",
		"repo",
		"admin:org",
		"Repository permissions:",
		"Organization permissions:",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("permissionsCommand output missing section: %s", section)
		}
	}

	// Test that GitHub token setup URL is present
	if !strings.Contains(output, "https://github.com/settings/tokens") {
		t.Error("permissionsCommand output missing GitHub token setup URL")
	}
}
