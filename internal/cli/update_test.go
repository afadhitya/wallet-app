package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/afadhitya/wallet-app/pkg/update"
)

func mockReleaseServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, func()) {
	t.Helper()
	server := httptest.NewServer(handler)
	origHTTP := update.GetHTTPGet()
	update.SetHTTPGet(func(url string) (*http.Response, error) {
		return http.Get(server.URL + url[len("https://api.github.com"):])
	})
	return server, func() { update.SetHTTPGet(origHTTP) }
}

func TestUpdateSuccess(t *testing.T) {
	origVer := update.Version
	update.Version = "v1.0.0"
	defer func() { update.Version = origVer }()

	var baseURL, checksumPath, archivePath string
	server, cleanup := mockReleaseServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/afadhitya/wallet-app/releases/latest":
			release := map[string]interface{}{
				"tag_name":   "v1.1.0",
				"prerelease": false,
				"assets": []map[string]interface{}{
					{
						"name":                 update.PlatformAssetName(),
						"browser_download_url": baseURL + "/dl/archive",
					},
					{
						"name":                 "checksums.txt",
						"browser_download_url": baseURL + "/dl/checksums",
					},
				},
			}
			_ = json.NewEncoder(w).Encode(release)
		case "/dl/archive":
			data, _ := os.ReadFile(archivePath)
			_, _ = w.Write(data)
		case "/dl/checksums":
			data, _ := os.ReadFile(checksumPath)
			_, _ = w.Write(data)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer server.Close()
	defer cleanup()

	baseURL = server.URL
	checksumPath = filepath.Join("..", "..", "pkg", "update", "testdata", "checksums.txt")
	archivePath = filepath.Join("..", "..", "pkg", "update", "testdata", "wallet_darwin_amd64.tar.gz")

	tmpDir := t.TempDir()
	execPath := filepath.Join(tmpDir, "wallet")
	_ = os.WriteFile(execPath, []byte("old binary"), 0755)

	origExec := update.GetExecutablePath()
	update.SetExecutablePath(func() (string, error) { return execPath, nil })
	defer func() { update.SetExecutablePath(origExec) }()

	cmd := newUpdateCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("update command error: %v", err)
	}
}

func TestUpdateJSONSuccess(t *testing.T) {
	origVer := update.Version
	update.Version = "v1.0.0"
	defer func() { update.Version = origVer }()

	var baseURL, checksumPath, archivePath string
	server, cleanup := mockReleaseServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/afadhitya/wallet-app/releases/latest":
			release := map[string]interface{}{
				"tag_name":   "v1.1.0",
				"prerelease": false,
				"assets": []map[string]interface{}{
					{
						"name":                 update.PlatformAssetName(),
						"browser_download_url": baseURL + "/dl/archive",
					},
					{
						"name":                 "checksums.txt",
						"browser_download_url": baseURL + "/dl/checksums",
					},
				},
			}
			_ = json.NewEncoder(w).Encode(release)
		case "/dl/archive":
			data, _ := os.ReadFile(archivePath)
			_, _ = w.Write(data)
		case "/dl/checksums":
			data, _ := os.ReadFile(checksumPath)
			_, _ = w.Write(data)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer server.Close()
	defer cleanup()

	baseURL = server.URL
	checksumPath = filepath.Join("..", "..", "pkg", "update", "testdata", "checksums.txt")
	archivePath = filepath.Join("..", "..", "pkg", "update", "testdata", "wallet_darwin_amd64.tar.gz")

	tmpDir := t.TempDir()
	execPath := filepath.Join(tmpDir, "wallet")
	_ = os.WriteFile(execPath, []byte("old binary"), 0755)

	origExec := update.GetExecutablePath()
	update.SetExecutablePath(func() (string, error) { return execPath, nil })
	defer func() { update.SetExecutablePath(origExec) }()

	cmd := newUpdateCmd()
	cmd.SetArgs([]string{"--json"})
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("update --json error: %v", err)
	}

	var resp successResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if !resp.Success {
		t.Error("expected success: true")
	}
}

func TestUpdateAlreadyLatest(t *testing.T) {
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

	cmd := newUpdateCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	_ = cmd.Execute()
}

func TestUpdateAlreadyLatestJSON(t *testing.T) {
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

	cmd := newUpdateCmd()
	cmd.SetArgs([]string{"--json"})
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	_ = cmd.Execute()
}

func TestUpdateForceAlreadyLatest(t *testing.T) {
	origVer := update.Version
	update.Version = "v1.2.0"
	defer func() { update.Version = origVer }()

	var baseURL, checksumPath, archivePath string
	server, cleanup := mockReleaseServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/afadhitya/wallet-app/releases/latest":
			release := map[string]interface{}{
				"tag_name":   "v1.2.0",
				"prerelease": false,
				"assets": []map[string]interface{}{
					{
						"name":                 update.PlatformAssetName(),
						"browser_download_url": baseURL + "/dl/archive",
					},
					{
						"name":                 "checksums.txt",
						"browser_download_url": baseURL + "/dl/checksums",
					},
				},
			}
			_ = json.NewEncoder(w).Encode(release)
		case "/dl/archive":
			data, _ := os.ReadFile(archivePath)
			_, _ = w.Write(data)
		case "/dl/checksums":
			data, _ := os.ReadFile(checksumPath)
			_, _ = w.Write(data)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer server.Close()
	defer cleanup()

	baseURL = server.URL
	checksumPath = filepath.Join("..", "..", "pkg", "update", "testdata", "checksums.txt")
	archivePath = filepath.Join("..", "..", "pkg", "update", "testdata", "wallet_darwin_amd64.tar.gz")

	tmpDir := t.TempDir()
	execPath := filepath.Join(tmpDir, "wallet")
	_ = os.WriteFile(execPath, []byte("old binary"), 0755)

	origExec := update.GetExecutablePath()
	update.SetExecutablePath(func() (string, error) { return execPath, nil })
	defer func() { update.SetExecutablePath(origExec) }()

	cmd := newUpdateCmd()
	cmd.SetArgs([]string{"--force"})
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("update --force error: %v", err)
	}
}

func TestUpdateNetworkError(t *testing.T) {
	origVer := update.Version
	update.Version = "v1.0.0"
	defer func() { update.Version = origVer }()

	server, cleanup := mockReleaseServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer server.Close()
	defer cleanup()

	cmd := newUpdateCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err == nil {
		t.Error("expected error on network failure")
	}
}

func TestUpdateNetworkErrorJSON(t *testing.T) {
	origVer := update.Version
	update.Version = "v1.0.0"
	defer func() { update.Version = origVer }()

	server, cleanup := mockReleaseServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer server.Close()
	defer cleanup()

	cmd := newUpdateCmd()
	cmd.SetArgs([]string{"--json"})
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	_ = cmd.Execute()
}

func TestUpdateChecksumError(t *testing.T) {
	origVer := update.Version
	update.Version = "v1.0.0"
	defer func() { update.Version = origVer }()

	var baseURL, checksumPath string
	server, cleanup := mockReleaseServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/afadhitya/wallet-app/releases/latest":
			release := map[string]interface{}{
				"tag_name":   "v1.1.0",
				"prerelease": false,
				"assets": []map[string]interface{}{
					{
						"name":                 update.PlatformAssetName(),
						"browser_download_url": baseURL + "/dl/archive",
					},
					{
						"name":                 "checksums.txt",
						"browser_download_url": baseURL + "/dl/checksums",
					},
				},
			}
			_ = json.NewEncoder(w).Encode(release)
		case "/dl/archive":
			_, _ = w.Write([]byte("tampered data"))
		case "/dl/checksums":
			data, _ := os.ReadFile(checksumPath)
			_, _ = w.Write(data)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer server.Close()
	defer cleanup()

	baseURL = server.URL
	checksumPath = filepath.Join("..", "..", "pkg", "update", "testdata", "checksums.txt")

	cmd := newUpdateCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err == nil {
		t.Error("expected error on checksum mismatch")
	}
}

func TestUpdatePermissionError(t *testing.T) {
	origVer := update.Version
	update.Version = "v1.0.0"
	defer func() { update.Version = origVer }()

	var baseURL, checksumPath, archivePath string
	server, cleanup := mockReleaseServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/afadhitya/wallet-app/releases/latest":
			release := map[string]interface{}{
				"tag_name":   "v1.1.0",
				"prerelease": false,
				"assets": []map[string]interface{}{
					{
						"name":                 update.PlatformAssetName(),
						"browser_download_url": baseURL + "/dl/archive",
					},
					{
						"name":                 "checksums.txt",
						"browser_download_url": baseURL + "/dl/checksums",
					},
				},
			}
			_ = json.NewEncoder(w).Encode(release)
		case "/dl/archive":
			data, _ := os.ReadFile(archivePath)
			_, _ = w.Write(data)
		case "/dl/checksums":
			data, _ := os.ReadFile(checksumPath)
			_, _ = w.Write(data)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer server.Close()
	defer cleanup()

	baseURL = server.URL
	checksumPath = filepath.Join("..", "..", "pkg", "update", "testdata", "checksums.txt")
	archivePath = filepath.Join("..", "..", "pkg", "update", "testdata", "wallet_darwin_amd64.tar.gz")

	execPath := "/nonexistent/path/wallet"
	origExec := update.GetExecutablePath()
	update.SetExecutablePath(func() (string, error) { return execPath, nil })
	defer func() { update.SetExecutablePath(origExec) }()

	cmd := newUpdateCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err == nil {
		t.Error("expected error on permission failure")
	}
}

func TestClassifyUpdateErrors(t *testing.T) {
	tests := []struct {
		err  error
		code string
	}{
		{update.ErrChecksumMismatch, ErrCodeUpdateChecksumMismatch},
		{update.ErrNetworkError, ErrCodeUpdateNetworkError},
		{update.ErrPermission, ErrCodeUpdatePermission},
		{update.ErrAlreadyLatest, ErrCodeUpdateAlreadyLatest},
		{update.ErrUpdateFailed, ErrCodeUpdateFailed},
	}

	for _, tt := range tests {
		code, _ := classifyError(tt.err)
		if code != tt.code {
			t.Errorf("classifyError(%v) = %q, want %q", tt.err, code, tt.code)
		}
	}
}
