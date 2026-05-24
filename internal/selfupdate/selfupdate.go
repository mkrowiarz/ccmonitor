// Package selfupdate replaces the running ccmonitor binary with the latest
// GitHub release build for the current platform.
package selfupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	repo        = "mkrowiarz/ccmonitor"
	latestURL   = "https://api.github.com/repos/" + repo + "/releases/latest"
	httpTimeout = 30 * time.Second
)

// release is the subset of the GitHub release API response we use.
type release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

// assetName returns the release asset filename for a platform, or an error if
// the platform is not one ccmonitor ships binaries for.
func assetName(goos, goarch string) (string, error) {
	switch goos {
	case "linux", "darwin":
	default:
		return "", fmt.Errorf("unsupported OS %q (no release binaries)", goos)
	}
	switch goarch {
	case "amd64", "arm64":
	default:
		return "", fmt.Errorf("unsupported architecture %q (no release binaries)", goarch)
	}
	return fmt.Sprintf("ccmonitor-%s-%s", goos, goarch), nil
}

// compareVersions returns -1 if a < b, 0 if equal, 1 if a > b. Leading "v" and
// any pre-release/build suffix (after "-") are ignored. Unparseable parts sort
// as 0, so "dev" compares as 0.0.0 and is treated as older than any release.
func compareVersions(a, b string) int {
	pa, pb := parseSemver(a), parseSemver(b)
	for i := 0; i < 3; i++ {
		if pa[i] < pb[i] {
			return -1
		}
		if pa[i] > pb[i] {
			return 1
		}
	}
	return 0
}

func parseSemver(v string) [3]int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	var out [3]int
	for i, part := range strings.SplitN(v, ".", 3) {
		if i > 2 {
			break
		}
		n, _ := strconv.Atoi(part)
		out[i] = n
	}
	return out
}

func fetchLatest(ctx context.Context) (*release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch latest release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %s", resp.Status)
	}
	var rel release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("decode release: %w", err)
	}
	if rel.TagName == "" {
		return nil, fmt.Errorf("no latest release found")
	}
	return &rel, nil
}

// Update checks the latest release and, if newer than current, downloads the
// matching binary and atomically replaces the running executable. Progress is
// written to out.
func Update(ctx context.Context, current string, out io.Writer) error {
	want, err := assetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Current version: %s\n", current)
	fmt.Fprintln(out, "Checking for updates…")
	rel, err := fetchLatest(ctx)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Latest version:  %s\n", rel.TagName)

	if compareVersions(current, rel.TagName) >= 0 {
		fmt.Fprintln(out, "Already up to date.")
		return nil
	}

	var assetURL string
	for _, a := range rel.Assets {
		if a.Name == want {
			assetURL = a.URL
			break
		}
	}
	if assetURL == "" {
		return fmt.Errorf("release %s has no asset %q", rel.TagName, want)
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate executable: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}

	fmt.Fprintf(out, "Downloading %s…\n", want)
	if err := downloadReplace(ctx, assetURL, exe); err != nil {
		return err
	}

	fmt.Fprintf(out, "Updated %s → %s\n", current, rel.TagName)
	return nil
}

// downloadReplace downloads url to a temp file beside exe, makes it executable,
// and atomically renames it over exe.
func downloadReplace(ctx context.Context, url, exe string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned %s", resp.Status)
	}

	dir := filepath.Dir(exe)
	tmp, err := os.CreateTemp(dir, ".ccmonitor-update-*")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w (no write permission? try reinstalling manually)", dir, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op after successful rename

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		return fmt.Errorf("write update: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0o755); err != nil {
		return err
	}
	if err := os.Rename(tmpName, exe); err != nil {
		return fmt.Errorf("replace %s: %w (no write permission? try reinstalling manually)", exe, err)
	}
	return nil
}
