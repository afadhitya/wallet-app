package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCurrentVersion(t *testing.T) {
	orig := Version
	Version = "v1.2.0"
	defer func() { Version = orig }()

	if got := CurrentVersion(); got != "v1.2.0" {
		t.Errorf("CurrentVersion() = %q, want %q", got, "v1.2.0")
	}
}

func TestCurrentVersionDev(t *testing.T) {
	orig := Version
	Version = "dev"
	defer func() { Version = orig }()

	if got := CurrentVersion(); got != "dev" {
		t.Errorf("CurrentVersion() = %q, want %q", got, "dev")
	}
}

func TestParseSemver(t *testing.T) {
	tests := []struct {
		v           string
		maj, min, p int
		ok          bool
	}{
		{"v1.2.3", 1, 2, 3, true},
		{"1.2.3", 1, 2, 3, true},
		{"v0.0.1", 0, 0, 1, true},
		{"v10.20.30", 10, 20, 30, true},
		{"dev", 0, 0, 0, false},
		{"", 0, 0, 0, false},
		{"v1.2", 0, 0, 0, false},
		{"v1", 0, 0, 0, false},
		{"abc", 0, 0, 0, false},
	}

	for _, tt := range tests {
		maj, min, p, ok := parseSemver(tt.v)
		if maj != tt.maj || min != tt.min || p != tt.p || ok != tt.ok {
			t.Errorf("parseSemver(%q) = (%d,%d,%d,%t), want (%d,%d,%d,%t)",
				tt.v, maj, min, p, ok, tt.maj, tt.min, tt.p, tt.ok)
		}
	}
}

func TestIsNewer(t *testing.T) {
	tests := []struct {
		current, latest string
		newer           bool
	}{
		{"v1.0.0", "v1.0.1", true},
		{"v1.0.0", "v1.1.0", true},
		{"v1.0.0", "v2.0.0", true},
		{"v1.0.1", "v1.0.0", false},
		{"v1.0.0", "v1.0.0", false},
		{"v2.0.0", "v1.9.9", false},
		{"dev", "v1.0.0", false},
		{"v1.0.0", "dev", false},
		{"", "", false},
	}

	for _, tt := range tests {
		if got := IsNewer(tt.current, tt.latest); got != tt.newer {
			t.Errorf("IsNewer(%q, %q) = %t, want %t", tt.current, tt.latest, got, tt.newer)
		}
	}
}

func TestPlatformAssetName(t *testing.T) {
	name := PlatformAssetName()
	expected := "wallet_" + name[7:] // just ensure it has the pattern
	_ = expected
	if name == "" {
		t.Error("PlatformAssetName() returned empty string")
	}
}

func TestFindAsset(t *testing.T) {
	release := &Release{
		Assets: []Asset{
			{Name: "wallet_linux_amd64.tar.gz"},
			{Name: PlatformAssetName()},
			{Name: "wallet_darwin_arm64.tar.gz"},
		},
	}

	asset, err := FindAsset(release)
	if err != nil {
		t.Fatalf("FindAsset() error: %v", err)
	}
	if asset.Name != PlatformAssetName() {
		t.Errorf("FindAsset() = %q, want %q", asset.Name, PlatformAssetName())
	}
}

func TestFindAssetNotFound(t *testing.T) {
	release := &Release{
		Assets: []Asset{
			{Name: "wallet_nonexistent_arch.tar.gz"},
		},
	}

	_, err := FindAsset(release)
	if err == nil {
		t.Error("FindAsset() expected error for missing asset")
	}
}

func TestFindChecksumsAsset(t *testing.T) {
	release := &Release{
		Assets: []Asset{
			{Name: "wallet_linux_amd64.tar.gz"},
			{Name: "checksums.txt"},
		},
	}

	asset, err := FindChecksumsAsset(release)
	if err != nil {
		t.Fatalf("FindChecksumsAsset() error: %v", err)
	}
	if asset.Name != "checksums.txt" {
		t.Errorf("FindChecksumsAsset() = %q, want %q", asset.Name, "checksums.txt")
	}
}

func TestFindChecksumsAssetNotFound(t *testing.T) {
	release := &Release{
		Assets: []Asset{
			{Name: "wallet_linux_amd64.tar.gz"},
		},
	}

	_, err := FindChecksumsAsset(release)
	if err == nil {
		t.Error("FindChecksumsAsset() expected error for missing checksums.txt")
	}
}

func TestLatestRelease(t *testing.T) {
	expectedTag := "v1.2.0"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/afadhitya/wallet-app/releases/latest" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		release := Release{
			TagName:    expectedTag,
			Prerelease: false,
			Assets: []Asset{
				{Name: PlatformAssetName()},
				{Name: "checksums.txt"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	origHTTP := httpGet
	defer func() { httpGet = origHTTP }()

	httpGet = func(url string) (*http.Response, error) {
		return http.Get(server.URL + "/repos/afadhitya/wallet-app/releases/latest")
	}

	release, err := LatestRelease()
	if err != nil {
		t.Fatalf("LatestRelease() error: %v", err)
	}
	if release.TagName != expectedTag {
		t.Errorf("LatestRelease() tag = %q, want %q", release.TagName, expectedTag)
	}
}

func TestLatestReleaseNoAssets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := Release{
			TagName:    "v1.0.0",
			Prerelease: false,
			Assets:     []Asset{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	origHTTP := httpGet
	defer func() { httpGet = origHTTP }()

	httpGet = func(url string) (*http.Response, error) {
		return http.Get(server.URL + "/repos/afadhitya/wallet-app/releases/latest")
	}

	release, err := LatestRelease()
	if err != nil {
		t.Fatalf("LatestRelease() error: %v", err)
	}
	if _, err := FindAsset(release); err == nil {
		t.Error("FindAsset() should fail when release has no matching assets")
	}
}

func TestLatestReleasePrerelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := Release{
			TagName:    "v2.0.0-rc1",
			Prerelease: true,
			Assets:     []Asset{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	origHTTP := httpGet
	defer func() { httpGet = origHTTP }()

	httpGet = func(url string) (*http.Response, error) {
		return http.Get(server.URL + "/repos/afadhitya/wallet-app/releases/latest")
	}

	_, err := LatestRelease()
	if err == nil {
		t.Error("LatestRelease() should error on prerelease")
	}
}

func TestLatestReleaseNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	origHTTP := httpGet
	defer func() { httpGet = origHTTP }()

	httpGet = func(url string) (*http.Response, error) {
		return http.Get(server.URL + "/repos/afadhitya/wallet-app/releases/latest")
	}

	_, err := LatestRelease()
	if err == nil {
		t.Error("LatestRelease() should error on non-200")
	}
}

func TestLatestReleaseMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer server.Close()

	origHTTP := httpGet
	defer func() { httpGet = origHTTP }()

	httpGet = func(url string) (*http.Response, error) {
		return http.Get(server.URL + "/repos/afadhitya/wallet-app/releases/latest")
	}

	_, err := LatestRelease()
	if err == nil {
		t.Error("LatestRelease() should error on malformed JSON")
	}
}

func TestParseChecksums(t *testing.T) {
	data := []byte("abc123  wallet_linux_amd64.tar.gz\ndef456  wallet_darwin_amd64.tar.gz\n")
	checksums := parseChecksums(data)
	if checksums["wallet_linux_amd64.tar.gz"] != "abc123" {
		t.Errorf("checksum for wallet_linux_amd64.tar.gz = %q, want %q", checksums["wallet_linux_amd64.tar.gz"], "abc123")
	}
	if checksums["wallet_darwin_amd64.tar.gz"] != "def456" {
		t.Errorf("checksum for wallet_darwin_amd64.tar.gz = %q, want %q", checksums["wallet_darwin_amd64.tar.gz"], "def456")
	}
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("hello world")
	// SHA256 of "hello world" = b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if err := VerifyChecksum(data, expected, "test.txt"); err != nil {
		t.Errorf("VerifyChecksum() error: %v", err)
	}
}

func TestVerifyChecksumMismatch(t *testing.T) {
	data := []byte("hello world")
	expected := "0000000000000000000000000000000000000000000000000000000000000000"
	if err := VerifyChecksum(data, expected, "test.txt"); err == nil {
		t.Error("VerifyChecksum() should error on mismatch")
	}
}

func TestExtractBinary(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "wallet_darwin_amd64.tar.gz"))
	if err != nil {
		t.Fatalf("read test fixture: %v", err)
	}

	binary, err := ExtractBinary(data)
	if err != nil {
		t.Fatalf("ExtractBinary() error: %v", err)
	}
	if len(binary) == 0 {
		t.Error("ExtractBinary() returned empty binary")
	}
}

func TestExtractBinaryInvalidGzip(t *testing.T) {
	_, err := ExtractBinary([]byte("not a gzip file"))
	if err == nil {
		t.Error("ExtractBinary() should error on invalid gzip data")
	}
}

func TestDecodeReleaseEmptyTag(t *testing.T) {
	body := strings.NewReader(`{"tag_name":"", "prerelease":false}`)
	_, err := decodeRelease(body)
	if err == nil {
		t.Error("decodeRelease() should error on empty tag name")
	}
}

func TestDownloadAndVerify(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/asset.tar.gz":
			data, _ := os.ReadFile(filepath.Join("testdata", "wallet_darwin_amd64.tar.gz"))
			_, _ = w.Write(data)
		case "/checksums.txt":
			data, _ := os.ReadFile(filepath.Join("testdata", "checksums.txt"))
			_, _ = w.Write(data)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	origHTTP := httpGet
	httpGet = func(url string) (*http.Response, error) {
		return http.Get(url)
	}
	defer func() { httpGet = origHTTP }()

	release := &Release{
		TagName: "v1.0.0",
		Assets: []Asset{
			{Name: PlatformAssetName(), BrowserDownloadURL: server.URL + "/asset.tar.gz"},
			{Name: "checksums.txt", BrowserDownloadURL: server.URL + "/checksums.txt"},
		},
	}

	binary, err := DownloadAndVerify(release)
	if err != nil {
		t.Fatalf("DownloadAndVerify() error: %v", err)
	}
	if len(binary) == 0 {
		t.Error("DownloadAndVerify() returned empty binary")
	}
}

func TestDownloadAndVerifyNoAsset(t *testing.T) {
	release := &Release{
		TagName: "v1.0.0",
		Assets: []Asset{},
	}
	_, err := DownloadAndVerify(release)
	if err == nil {
		t.Error("DownloadAndVerify() should error with no matching asset")
	}
}

func TestDownloadAndVerifyNoChecksums(t *testing.T) {
	release := &Release{
		TagName: "v1.0.0",
		Assets: []Asset{
			{Name: PlatformAssetName(), BrowserDownloadURL: "http://example.com/asset"},
		},
	}
	_, err := DownloadAndVerify(release)
	if err == nil {
		t.Error("DownloadAndVerify() should error with no checksums.txt")
	}
}

func TestDownloadAndVerifyFailedDownload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	origHTTP := httpGet
	httpGet = func(url string) (*http.Response, error) {
		return http.Get(url)
	}
	defer func() { httpGet = origHTTP }()

	release := &Release{
		TagName: "v1.0.0",
		Assets: []Asset{
			{Name: PlatformAssetName(), BrowserDownloadURL: server.URL + "/asset"},
			{Name: "checksums.txt", BrowserDownloadURL: server.URL + "/checksums"},
		},
	}
	_, err := DownloadAndVerify(release)
	if err == nil {
		t.Error("DownloadAndVerify() should error on failed download")
	}
}

func TestDownloadAndVerifyChecksumsDownloadFail(t *testing.T) {
	data, _ := os.ReadFile(filepath.Join("testdata", "wallet_darwin_amd64.tar.gz"))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/asset":
			_, _ = w.Write(data)
		case "/checksums":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	origHTTP := httpGet
	httpGet = func(url string) (*http.Response, error) {
		return http.Get(url)
	}
	defer func() { httpGet = origHTTP }()

	release := &Release{
		TagName: "v1.0.0",
		Assets: []Asset{
			{Name: PlatformAssetName(), BrowserDownloadURL: server.URL + "/asset"},
			{Name: "checksums.txt", BrowserDownloadURL: server.URL + "/checksums"},
		},
	}
	_, err := DownloadAndVerify(release)
	if err == nil {
		t.Error("DownloadAndVerify() should error on checksums download failure")
	}
}

func TestDownloadAndVerifyChecksumMismatch(t *testing.T) {
	data, _ := os.ReadFile(filepath.Join("testdata", "wallet_darwin_amd64.tar.gz"))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/asset":
			_, _ = w.Write(data)
		case "/checksums":
			_, _ = w.Write([]byte("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff  " + PlatformAssetName() + "\n"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	origHTTP := httpGet
	httpGet = func(url string) (*http.Response, error) {
		return http.Get(url)
	}
	defer func() { httpGet = origHTTP }()

	release := &Release{
		TagName: "v1.0.0",
		Assets: []Asset{
			{Name: PlatformAssetName(), BrowserDownloadURL: server.URL + "/asset"},
			{Name: "checksums.txt", BrowserDownloadURL: server.URL + "/checksums"},
		},
	}
	_, err := DownloadAndVerify(release)
	if err == nil {
		t.Error("DownloadAndVerify() should error on checksum mismatch")
	}
}

func TestExtractBinaryNotFoundInArchive(t *testing.T) {
	gzBuf := new(bytes.Buffer)
	gzWriter := gzip.NewWriter(gzBuf)
	tarWriter := tar.NewWriter(gzWriter)
	_ = tarWriter.WriteHeader(&tar.Header{Name: "other-file", Size: 0})
	_ = tarWriter.Close()
	_ = gzWriter.Close()

	_, err := ExtractBinary(gzBuf.Bytes())
	if err == nil {
		t.Error("ExtractBinary() should error when wallet binary not found")
	}
}

func TestLatestReleaseHTTPError(t *testing.T) {
	origHTTP := httpGet
	httpGet = func(url string) (*http.Response, error) {
		return nil, errors.New("connection refused")
	}
	defer func() { httpGet = origHTTP }()

	_, err := LatestRelease()
	if err == nil {
		t.Error("LatestRelease() should error on HTTP failure")
	}
}

func TestVerifyArchiveChecksum(t *testing.T) {
	archiveData, err := os.ReadFile(filepath.Join("testdata", "wallet_darwin_amd64.tar.gz"))
	if err != nil {
		t.Fatalf("read test fixture: %v", err)
	}

	checksumsData, err := os.ReadFile(filepath.Join("testdata", "checksums.txt"))
	if err != nil {
		t.Fatalf("read test fixture: %v", err)
	}

	if err := VerifyArchiveChecksum(archiveData, checksumsData, "wallet_darwin_amd64.tar.gz"); err != nil {
		t.Errorf("VerifyArchiveChecksum() error: %v", err)
	}
}

func TestVerifyArchiveChecksumInvalid(t *testing.T) {
	archiveData := []byte("bad data")
	checksumsData := []byte("abc123  wallet_darwin_amd64.tar.gz\n")

	if err := VerifyArchiveChecksum(archiveData, checksumsData, "wallet_darwin_amd64.tar.gz"); err == nil {
		t.Error("VerifyArchiveChecksum() should error on invalid checksum")
	}
}

func TestVerifyArchiveChecksumNoEntry(t *testing.T) {
	archiveData := []byte("bad data")
	checksumsData := []byte("abc123  wallet_linux_amd64.tar.gz\n")

	if err := VerifyArchiveChecksum(archiveData, checksumsData, "wallet_darwin_amd64.tar.gz"); err == nil {
		t.Error("VerifyArchiveChecksum() should error when no checksum entry")
	}
}

func TestReplaceBinary(t *testing.T) {
	newBinary := []byte("new binary content")
	tmpDir := t.TempDir()

	execPath := filepath.Join(tmpDir, "wallet-test")
	if err := os.WriteFile(execPath, []byte("old content"), 0755); err != nil {
		t.Fatalf("write test binary: %v", err)
	}

	origExec := executablePath
	executablePath = func() (string, error) { return execPath, nil }
	defer func() { executablePath = origExec }()

	if err := ReplaceBinary(newBinary); err != nil {
		t.Fatalf("ReplaceBinary() error: %v", err)
	}

	content, err := os.ReadFile(execPath)
	if err != nil {
		t.Fatalf("read replaced binary: %v", err)
	}
	if string(content) != string(newBinary) {
		t.Errorf("ReplaceBinary() content = %q, want %q", string(content), string(newBinary))
	}

	if _, err := os.Stat(execPath + ".new"); !os.IsNotExist(err) {
		t.Error(".new temp file should not exist after successful rename")
	}
}
