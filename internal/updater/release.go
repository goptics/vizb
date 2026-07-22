package updater

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"golang.org/x/mod/semver"
)

type releaseAsset struct {
	Name      string
	Extension string
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	if version != "" && version[0] != 'v' {
		version = "v" + version
	}
	return version
}

func latestReleaseVersion(ctx context.Context, client *http.Client, latestURL, userAgent string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, latestURL, nil)
	if err != nil {
		return "", fmt.Errorf("create latest release request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("resolve latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("resolve latest release: server returned %s", resp.Status)
	}

	tag, err := url.PathUnescape(path.Base(resp.Request.URL.Path))
	if err != nil {
		return "", fmt.Errorf("decode latest release tag: %w", err)
	}
	tag = normalizeVersion(tag)
	if !semver.IsValid(tag) {
		return "", fmt.Errorf("latest release returned invalid semantic version %q", tag)
	}
	return tag, nil
}

func assetFor(version, goos, goarch string) (releaseAsset, error) {
	if goos != "linux" && goos != "darwin" && goos != "windows" {
		return releaseAsset{}, fmt.Errorf("unsupported operating system %q", goos)
	}
	if goarch != "amd64" && goarch != "arm64" {
		return releaseAsset{}, fmt.Errorf("unsupported architecture %q", goarch)
	}

	extension := ".tar.gz"
	if goos == "windows" {
		extension = ".zip"
	}
	releaseVersion := strings.TrimPrefix(normalizeVersion(version), "v")
	return releaseAsset{
		Name:      fmt.Sprintf("vizb@%s-%s-%s%s", releaseVersion, goos, goarch, extension),
		Extension: extension,
	}, nil
}
