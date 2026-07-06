package update

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

var Version = "dev"

func CurrentVersion() string {
	return Version
}

func parseSemver(v string) (major, minor, patch int, ok bool) {
	v = strings.TrimPrefix(v, "v")
	_, err := fmt.Sscanf(v, "%d.%d.%d", &major, &minor, &patch)
	if err != nil {
		return 0, 0, 0, false
	}
	return major, minor, patch, true
}

func IsNewer(current, latest string) bool {
	cMaj, cMin, cPatch, cOk := parseSemver(current)
	lMaj, lMin, lPatch, lOk := parseSemver(latest)
	if !cOk || !lOk {
		return false
	}
	if lMaj != cMaj {
		return lMaj > cMaj
	}
	if lMin != cMin {
		return lMin > cMin
	}
	return lPatch > cPatch
}

type Release struct {
	TagName     string  `json:"tag_name"`
	Prerelease  bool    `json:"prerelease"`
	HTMLURL     string  `json:"html_url"`
	Assets      []Asset `json:"assets"`
	AssetsURL   string  `json:"assets_url"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

func getHTTPClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

var httpGet = func(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "wallet-cli")
	return getHTTPClient().Do(req)
}

func LatestRelease() (*Release, error) {
	resp, err := httpGet("https://api.github.com/repos/afadhitya/wallet-app/releases/latest")
	if err != nil {
		return nil, fmt.Errorf("fetch latest release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	release, err := decodeRelease(resp.Body)
	if err != nil {
		return nil, err
	}

	if release.Prerelease {
		return nil, errors.New("latest release is a prerelease")
	}

	return release, nil
}

func decodeRelease(body io.Reader) (*Release, error) {
	var release Release
	if err := json.NewDecoder(body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decode release: %w", err)
	}
	if release.TagName == "" {
		return nil, errors.New("release has no tag name")
	}
	return &release, nil
}

func PlatformAssetName() string {
	return fmt.Sprintf("wallet_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
}

func FindAsset(release *Release) (*Asset, error) {
	expected := PlatformAssetName()
	for _, a := range release.Assets {
		if a.Name == expected {
			return &a, nil
		}
	}
	return nil, fmt.Errorf("no asset found for platform %s/%s (%s)", runtime.GOOS, runtime.GOARCH, expected)
}

func FindChecksumsAsset(release *Release) (*Asset, error) {
	for _, a := range release.Assets {
		if a.Name == "checksums.txt" {
			return &a, nil
		}
	}
	return nil, errors.New("no checksums.txt found in release")
}

func downloadBytes(url string) ([]byte, error) {
	resp, err := httpGet(url)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: unexpected status %d", url, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	return data, nil
}

func parseChecksums(data []byte) map[string]string {
	checksums := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			checksums[parts[1]] = parts[0]
		}
	}
	return checksums
}

func VerifyChecksum(data []byte, expectedHash, filename string) error {
	hash := sha256.Sum256(data)
	got := hex.EncodeToString(hash[:])
	if got != expectedHash {
		return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", filename, expectedHash, got)
	}
	return nil
}

func VerifyArchiveChecksum(archiveData []byte, checksumsData []byte, assetName string) error {
	checksums := parseChecksums(checksumsData)
	expected, ok := checksums[assetName]
	if !ok {
		return fmt.Errorf("no checksum entry for %s", assetName)
	}
	return VerifyChecksum(archiveData, expected, assetName)
}

func ExtractBinary(archiveData []byte) ([]byte, error) {
	gzReader, err := gzip.NewReader(strings.NewReader(string(archiveData)))
	if err != nil {
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	defer func() { _ = gzReader.Close() }()

	tr := tar.NewReader(gzReader)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("tar reader: %w", err)
		}
		if hdr.Name == "wallet" || strings.HasSuffix(hdr.Name, "/wallet") {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("read binary from archive: %w", err)
			}
			return data, nil
		}
	}
	return nil, errors.New("wallet binary not found in archive")
}

func DownloadAndVerify(release *Release) ([]byte, error) {
	asset, err := FindAsset(release)
	if err != nil {
		return nil, err
	}

	checksumsAsset, err := FindChecksumsAsset(release)
	if err != nil {
		return nil, err
	}

	archiveData, err := downloadBytes(asset.BrowserDownloadURL)
	if err != nil {
		return nil, err
	}

	checksumsData, err := downloadBytes(checksumsAsset.BrowserDownloadURL)
	if err != nil {
		return nil, err
	}

	if err := VerifyArchiveChecksum(archiveData, checksumsData, asset.Name); err != nil {
		return nil, err
	}

	binary, err := ExtractBinary(archiveData)
	if err != nil {
		return nil, err
	}

	return binary, nil
}

var executablePath = os.Executable

func ReplaceBinary(newBinary []byte) error {
	execPath, err := executablePath()
	if err != nil {
		return fmt.Errorf("get current executable: %w", err)
	}

	tmpPath := execPath + ".new"

	if err := os.WriteFile(tmpPath, newBinary, 0755); err != nil {
		return fmt.Errorf("write new binary: %w", err)
	}

	if err := os.Rename(tmpPath, execPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("replace binary: %w", err)
	}

	return nil
}

var (
	ErrChecksumMismatch = errors.New("checksum mismatch")
	ErrNetworkError     = errors.New("network error")
	ErrPermission       = errors.New("permission error")
	ErrAlreadyLatest    = errors.New("already at latest version")
	ErrUpdateFailed     = errors.New("update failed")
)

func GetHTTPGet() func(string) (*http.Response, error) {
	return httpGet
}

func SetHTTPGet(fn func(string) (*http.Response, error)) {
	httpGet = fn
}

func GetExecutablePath() func() (string, error) {
	return executablePath
}

func SetExecutablePath(fn func() (string, error)) {
	executablePath = fn
}
