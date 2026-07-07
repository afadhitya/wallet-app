package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/afadhitya/wallet-app/pkg/update"
)

func TestVersionText(t *testing.T) {
	orig := update.Version
	update.Version = "v1.2.0"
	defer func() { update.Version = orig }()

	cmd := newVersionCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version command error: %v", err)
	}

	output := stdout.String()
	if output != "v1.2.0\n" {
		t.Errorf("expected 'v1.2.0\\n', got %q", output)
	}
}

func TestVersionJSON(t *testing.T) {
	orig := update.Version
	update.Version = "v1.2.0"
	defer func() { update.Version = orig }()

	cmd := newVersionCmd()
	cmd.SetArgs([]string{"--json"})
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version --json error: %v", err)
	}

	var resp successResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if !resp.Success {
		t.Error("expected success: true")
	}
}

func TestVersionDev(t *testing.T) {
	orig := update.Version
	update.Version = "dev"
	defer func() { update.Version = orig }()

	cmd := newVersionCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version command error: %v", err)
	}

	output := stdout.String()
	if output != "dev\n" {
		t.Errorf("expected 'dev\\n', got %q", output)
	}
}

func TestVersionCheckAlreadyLatest(t *testing.T) {
	origVer := update.Version
	update.Version = "v1.2.0"
	defer func() { update.Version = origVer }()

	server, cleanup := mockReleaseServer(t, func(w http.ResponseWriter, r *http.Request) {
		release := map[string]interface{}{
			"tag_name":   "v1.2.0",
			"prerelease": false,
			"html_url":   "https://github.com/test/release",
			"assets":     []interface{}{},
			"assets_url": "",
		}
		_ = json.NewEncoder(w).Encode(release)
	})
	defer server.Close()
	defer cleanup()

	cmd := newVersionCmd()
	cmd.SetArgs([]string{"--check"})
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version --check error: %v", err)
	}
}

func TestVersionCheckUpdateAvailable(t *testing.T) {
	origVer := update.Version
	update.Version = "v1.0.0"
	defer func() { update.Version = origVer }()

	server, cleanup := mockReleaseServer(t, func(w http.ResponseWriter, r *http.Request) {
		release := map[string]interface{}{
			"tag_name":   "v1.1.0",
			"prerelease": false,
			"html_url":   "https://github.com/test/release",
			"assets":     []interface{}{},
			"assets_url": "",
		}
		_ = json.NewEncoder(w).Encode(release)
	})
	defer server.Close()
	defer cleanup()

	cmd := newVersionCmd()
	cmd.SetArgs([]string{"--check"})
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version --check error: %v", err)
	}
}

func TestVersionCheckJSON(t *testing.T) {
	origVer := update.Version
	update.Version = "v1.2.0"
	defer func() { update.Version = origVer }()

	server, cleanup := mockReleaseServer(t, func(w http.ResponseWriter, r *http.Request) {
		release := map[string]interface{}{
			"tag_name":   "v1.3.0",
			"prerelease": false,
			"html_url":   "https://github.com/test/release",
			"assets":     []interface{}{},
			"assets_url": "",
		}
		_ = json.NewEncoder(w).Encode(release)
	})
	defer server.Close()
	defer cleanup()

	cmd := newVersionCmd()
	cmd.SetArgs([]string{"--check", "--json"})
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version --check --json error: %v", err)
	}

	var resp successResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if !resp.Success {
		t.Error("expected success: true")
	}
}

func TestVersionCheckNetworkError(t *testing.T) {
	origVer := update.Version
	update.Version = "v1.2.0"
	defer func() { update.Version = origVer }()

	server, cleanup := mockReleaseServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer server.Close()
	defer cleanup()

	cmd := newVersionCmd()
	cmd.SetArgs([]string{"--check"})
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version --check error (network): %v", err)
	}
}

func TestVersionCheckNetworkErrorJSON(t *testing.T) {
	origVer := update.Version
	update.Version = "v1.0.0"
	defer func() { update.Version = origVer }()

	server, cleanup := mockReleaseServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer server.Close()
	defer cleanup()

	cmd := newVersionCmd()
	cmd.SetArgs([]string{"--check", "--json"})
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version --check --json error (network): %v", err)
	}

	var resp successResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if !resp.Success {
		t.Error("expected success: true on network error")
	}
}

func TestVersionCheckAlreadyLatestJSON(t *testing.T) {
	origVer := update.Version
	update.Version = "v1.2.0"
	defer func() { update.Version = origVer }()

	server, cleanup := mockReleaseServer(t, func(w http.ResponseWriter, r *http.Request) {
		release := map[string]interface{}{
			"tag_name":   "v1.2.0",
			"prerelease": false,
			"assets":     []interface{}{},
		}
		_ = json.NewEncoder(w).Encode(release)
	})
	defer server.Close()
	defer cleanup()

	cmd := newVersionCmd()
	cmd.SetArgs([]string{"--check", "--json"})
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version --check --json error: %v", err)
	}

	var resp successResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if !resp.Success {
		t.Error("expected success: true")
	}
}

func TestVersionHelp(t *testing.T) {
	cmd := newVersionCmd()
	if cmd.Use != "version" {
		t.Errorf("command use = %q, want %q", cmd.Use, "version")
	}

	var stdout, stderr bytes.Buffer
	root := NewRootCmd()
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"version", "--help"})
	err := root.Execute()
	if err != nil {
		t.Fatalf("root version --help error: %v", err)
	}
}

func TestVersionInRoot(t *testing.T) {
	root := NewRootCmd()
	var found bool
	for _, c := range root.Commands() {
		if c.Use == "version" {
			found = true
			break
		}
	}
	if !found {
		t.Error("version command not registered in root")
	}
}

func TestUpdateInRoot(t *testing.T) {
	root := NewRootCmd()
	var found bool
	for _, c := range root.Commands() {
		if c.Use == "update" {
			found = true
			break
		}
	}
	if !found {
		t.Error("update command not registered in root")
	}
}

func TestUpdateFlags(t *testing.T) {
	cmd := newUpdateCmd()
	force := cmd.Flags().Lookup("force")
	if force == nil {
		t.Error("update command missing --force flag")
	}
}

func TestVersionCheckFlag(t *testing.T) {
	cmd := newVersionCmd()
	check := cmd.Flags().Lookup("check")
	if check == nil {
		t.Error("version command missing --check flag")
	}
}

func TestCommandUsesRootJSON(t *testing.T) {
	orig := update.Version
	update.Version = "v1.0.0"
	defer func() { update.Version = orig }()

	cmd := newVersionCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	_ = cmd.Flags().Set("json", "true")

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version command error: %v", err)
	}

	var resp successResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if !resp.Success {
		t.Error("expected success: true")
	}
}
