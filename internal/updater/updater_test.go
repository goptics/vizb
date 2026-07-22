package updater

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UpdaterSuite struct {
	suite.Suite
}

func (s *UpdaterSuite) TestNewUsesCurrentRuntimeAndExecutable() {
	service, err := New("v1.2.3", "standalone")
	s.Require().NoError(err)
	s.Equal("v1.2.3", service.currentVersion)
	s.Equal("standalone", service.distribution)
	s.Equal(runtime.GOOS, service.goos)
	s.Equal(runtime.GOARCH, service.goarch)
	s.NotEmpty(service.rawExecutable)
	s.NotEmpty(service.canonicalExecutable)
	s.NotNil(service.client)
	s.NotNil(service.replace)
}

func (s *UpdaterSuite) TestDetectsInstallationManager() {
	tests := []struct {
		name        string
		ctx         detectionContext
		kind        installationKind
		manager     string
		instruction string
	}{
		{
			name: "winget link",
			ctx: detectionContext{
				GOOS:          "windows",
				RawExecutable: `C:\Users\me\AppData\Local\Microsoft\WinGet\Links\vizb.exe`,
				Distribution:  "standalone",
			},
			kind:        installationManaged,
			manager:     "WinGet",
			instruction: "winget upgrade --id goptics.vizb --exact",
		},
		{
			name: "winget canonical package",
			ctx: detectionContext{
				GOOS:                "windows",
				CanonicalExecutable: `C:\Users\me\AppData\Local\Microsoft\WinGet\Packages\goptics.vizb\vizb.exe`,
				Distribution:        "standalone",
			},
			kind:        installationManaged,
			manager:     "WinGet",
			instruction: "winget upgrade --id goptics.vizb --exact",
		},
		{
			name: "go install release",
			ctx: detectionContext{
				GOOS:  "linux",
				Build: buildInfo{MainPath: modulePath, MainVersion: "v1.2.3"},
			},
			kind:        installationManaged,
			manager:     "Go toolchain",
			instruction: "go install github.com/goptics/vizb@latest",
		},
		{
			name: "go install pseudo version",
			ctx: detectionContext{
				GOOS:  "linux",
				Build: buildInfo{MainPath: modulePath, MainVersion: "v1.2.4-0.20260718140803-914ada16708f"},
			},
			kind:        installationManaged,
			manager:     "Go toolchain",
			instruction: "go install github.com/goptics/vizb@latest",
		},
		{
			name: "standalone marker wins over build metadata",
			ctx: detectionContext{
				GOOS:         "linux",
				Distribution: "standalone",
				Build:        buildInfo{MainPath: modulePath, MainVersion: "v1.2.3"},
			},
			kind:    installationStandalone,
			manager: "Vizb installer",
		},
		{
			name: "future package distribution",
			ctx: detectionContext{
				GOOS:         "darwin",
				Distribution: "homebrew",
			},
			kind:    installationManaged,
			manager: "homebrew",
		},
		{
			name: "development build",
			ctx: detectionContext{
				GOOS:  "linux",
				Build: buildInfo{MainPath: modulePath, MainVersion: "(devel)"},
			},
			kind: installationUnknown,
		},
		{
			name: "winget-looking path on unix",
			ctx: detectionContext{
				GOOS:          "linux",
				RawExecutable: "/microsoft/winget/links/vizb",
			},
			kind: installationUnknown,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			got := detectInstallation(test.ctx)
			s.Equal(test.kind, got.Kind)
			s.Equal(test.manager, got.Manager)
			s.Equal(test.instruction, got.Instruction)
		})
	}
}

func (s *UpdaterSuite) TestAssetNamingAndPlatformValidation() {
	tests := []struct {
		name      string
		version   string
		goos      string
		goarch    string
		assetName string
		extension string
		wantErr   string
	}{
		{name: "linux amd64", version: "v1.2.3", goos: "linux", goarch: "amd64", assetName: "vizb@1.2.3-linux-amd64.tar.gz", extension: ".tar.gz"},
		{name: "darwin arm64", version: "1.2.3", goos: "darwin", goarch: "arm64", assetName: "vizb@1.2.3-darwin-arm64.tar.gz", extension: ".tar.gz"},
		{name: "windows unsupported", version: "v1.2.3", goos: "windows", goarch: "arm64", wantErr: "unsupported operating system"},
		{name: "unsupported os", version: "v1.2.3", goos: "freebsd", goarch: "amd64", wantErr: "unsupported operating system"},
		{name: "unsupported arch", version: "v1.2.3", goos: "linux", goarch: "386", wantErr: "unsupported architecture"},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			asset, err := assetFor(test.version, test.goos, test.goarch)
			if test.wantErr != "" {
				s.Require().ErrorContains(err, test.wantErr)
				return
			}
			s.Require().NoError(err)
			s.Equal(test.assetName, asset.Name)
			s.Equal(test.extension, asset.Extension)
		})
	}
}

func (s *UpdaterSuite) TestStandaloneUpdateDownloadsVerifiesAndReplaces() {
	archive := tarGzipArchive(s.T(), "vizb", []byte("new-vizb"))
	server := newReleaseServer(s.T(), "v1.1.0", "vizb@1.1.0-linux-amd64.tar.gz", archive, false)
	defer server.Close()

	target := filepath.Join(s.T().TempDir(), "vizb")
	s.Require().NoError(os.WriteFile(target, []byte("old-vizb"), 0o751))

	service := standaloneUpdater(server, target, "v1.0.0", "linux", "amd64")
	var stdout bytes.Buffer
	s.Require().NoError(service.Run(context.Background(), strings.NewReader(""), &stdout, io.Discard))

	installed, err := os.ReadFile(target)
	s.Require().NoError(err)
	s.Equal("new-vizb", string(installed))
	info, err := os.Stat(target)
	s.Require().NoError(err)
	s.Equal(os.FileMode(0o751), info.Mode().Perm())
	s.Contains(stdout.String(), "Updated vizb from v1.0.0 to v1.1.0")
}

func (s *UpdaterSuite) TestChecksumFailurePreservesExecutable() {
	archive := tarGzipArchive(s.T(), "vizb", []byte("new-vizb"))
	server := newReleaseServer(s.T(), "v1.1.0", "vizb@1.1.0-linux-amd64.tar.gz", archive, true)
	defer server.Close()

	target := filepath.Join(s.T().TempDir(), "vizb")
	s.Require().NoError(os.WriteFile(target, []byte("old-vizb"), 0o755))
	service := standaloneUpdater(server, target, "v1.0.0", "linux", "amd64")

	err := service.Run(context.Background(), strings.NewReader(""), io.Discard, io.Discard)
	s.Require().ErrorContains(err, "checksum mismatch")
	installed, readErr := os.ReadFile(target)
	s.Require().NoError(readErr)
	s.Equal("old-vizb", string(installed))
}

func (s *UpdaterSuite) TestMissingChecksumPreservesExecutable() {
	archive := tarGzipArchive(s.T(), "vizb", []byte("new-vizb"))
	server := newReleaseServer(s.T(), "v1.1.0", "another-asset.tar.gz", archive, false)
	defer server.Close()

	target := filepath.Join(s.T().TempDir(), "vizb")
	s.Require().NoError(os.WriteFile(target, []byte("old-vizb"), 0o755))
	service := standaloneUpdater(server, target, "v1.0.0", "linux", "amd64")

	err := service.Run(context.Background(), strings.NewReader(""), io.Discard, io.Discard)
	s.Require().ErrorContains(err, "checksum for vizb@1.1.0-linux-amd64.tar.gz not found")
	installed, readErr := os.ReadFile(target)
	s.Require().NoError(readErr)
	s.Equal("old-vizb", string(installed))
}

func (s *UpdaterSuite) TestDownloadAndReplacementFailuresPreserveExecutable() {
	s.Run("archive download", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/latest":
				http.Redirect(w, r, "/releases/tag/v1.1.0", http.StatusFound)
			case r.URL.Path == "/releases/tag/v1.1.0":
				w.WriteHeader(http.StatusOK)
			case strings.HasSuffix(r.URL.Path, "/checksums.txt"):
				_, _ = io.WriteString(w, strings.Repeat("0", 64)+"  vizb@1.1.0-linux-amd64.tar.gz\n")
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		target := filepath.Join(s.T().TempDir(), "vizb")
		s.Require().NoError(os.WriteFile(target, []byte("old-vizb"), 0o755))
		service := standaloneUpdater(server, target, "v1.0.0", "linux", "amd64")
		err := service.Run(context.Background(), strings.NewReader(""), io.Discard, io.Discard)
		s.Require().ErrorContains(err, "404 Not Found")
		installed, readErr := os.ReadFile(target)
		s.Require().NoError(readErr)
		s.Equal("old-vizb", string(installed))
	})

	s.Run("atomic replacement", func() {
		archive := tarGzipArchive(s.T(), "vizb", []byte("new-vizb"))
		server := newReleaseServer(s.T(), "v1.1.0", "vizb@1.1.0-linux-amd64.tar.gz", archive, false)
		defer server.Close()

		target := filepath.Join(s.T().TempDir(), "vizb")
		s.Require().NoError(os.WriteFile(target, []byte("old-vizb"), 0o755))
		service := standaloneUpdater(server, target, "v1.0.0", "linux", "amd64")
		service.replace = func(_, _ string) error { return os.ErrPermission }
		err := service.Run(context.Background(), strings.NewReader(""), io.Discard, io.Discard)
		s.Require().ErrorContains(err, "permission denied")
		installed, readErr := os.ReadFile(target)
		s.Require().NoError(readErr)
		s.Equal("old-vizb", string(installed))
	})
}

func (s *UpdaterSuite) TestAlreadyCurrentAndNewerBuildDoNotInstall() {
	for _, current := range []string{"v1.1.0", "v1.2.0"} {
		s.Run(current, func() {
			server := newReleaseServer(s.T(), "v1.1.0", "unused", nil, false)
			defer server.Close()

			target := filepath.Join(s.T().TempDir(), "vizb")
			s.Require().NoError(os.WriteFile(target, []byte("current"), 0o755))
			service := standaloneUpdater(server, target, current, "linux", "amd64")
			service.replace = func(_, _ string) error {
				s.FailNow("replacement must not run")
				return nil
			}

			var stdout bytes.Buffer
			s.Require().NoError(service.Run(context.Background(), strings.NewReader(""), &stdout, io.Discard))
			if current == "v1.1.0" {
				s.Contains(stdout.String(), "already up to date")
			} else {
				s.Contains(stdout.String(), "no downgrade performed")
			}
		})
	}
}

func (s *UpdaterSuite) TestManagedAndUnknownBuildsNeverContactReleaseServer() {
	tests := []struct {
		name    string
		service *Updater
		output  string
		wantErr string
	}{
		{
			name: "go toolchain",
			service: &Updater{
				build: buildInfo{MainPath: modulePath, MainVersion: "v1.0.0"},
				goos:  "linux",
			},
			output: "go install github.com/goptics/vizb@latest",
		},
		{
			name: "winget",
			service: &Updater{
				goos:          "windows",
				rawExecutable: `C:\Users\me\AppData\Local\Microsoft\WinGet\Links\vizb.exe`,
				distribution:  "standalone",
			},
			output: "winget upgrade --id goptics.vizb --exact",
		},
		{
			name:    "development",
			service: &Updater{goos: "linux"},
			wantErr: "cannot safely update",
		},
		{
			name:    "future manager",
			service: &Updater{goos: "darwin", distribution: "homebrew"},
			wantErr: "managed by homebrew",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			var stdout bytes.Buffer
			err := test.service.Run(context.Background(), strings.NewReader(""), &stdout, io.Discard)
			if test.wantErr != "" {
				s.Require().ErrorContains(err, test.wantErr)
			} else {
				s.Require().NoError(err)
				s.Contains(stdout.String(), test.output)
			}
		})
	}
}

func (s *UpdaterSuite) TestWindowsStandalonePrintsReinstallCommandWithoutNetwork() {
	service := &Updater{
		currentVersion:      "devel",
		distribution:        "standalone",
		rawExecutable:       `C:\Users\me\AppData\Local\vizb\vizb.exe`,
		canonicalExecutable: `C:\Users\me\AppData\Local\vizb\vizb.exe`,
		goos:                "windows",
		goarch:              "amd64",
	}

	var stdout bytes.Buffer
	s.Require().NoError(service.Run(context.Background(), strings.NewReader(""), &stdout, io.Discard))
	s.Contains(stdout.String(), "not supported on Windows yet")
	s.Contains(stdout.String(), "irm https://vizb.goptics.org/install.ps1 | iex")
}

func (s *UpdaterSuite) TestReleaseResolutionAndErrors() {
	s.Run("invalid latest tag", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/latest" {
				http.Redirect(w, r, "/releases/tag/not-a-version", http.StatusFound)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		_, err := latestReleaseVersion(context.Background(), server.Client(), server.URL+"/latest", "vizb/test")
		s.Require().ErrorContains(err, "invalid semantic version")
	})

	s.Run("server error", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "no", http.StatusBadGateway)
		}))
		defer server.Close()

		_, err := latestReleaseVersion(context.Background(), server.Client(), server.URL, "vizb/test")
		s.Require().ErrorContains(err, "502 Bad Gateway")
	})

	s.Run("unsupported platform before network", func() {
		service := &Updater{distribution: "standalone", currentVersion: "v1.0.0", goos: "freebsd", goarch: "amd64"}
		err := service.Run(context.Background(), strings.NewReader(""), io.Discard, io.Discard)
		s.Require().ErrorContains(err, "unsupported operating system")
	})

	s.Run("invalid current version before network", func() {
		service := &Updater{distribution: "standalone", currentVersion: "devel", goos: "linux", goarch: "amd64"}
		err := service.Run(context.Background(), strings.NewReader(""), io.Discard, io.Discard)
		s.Require().ErrorContains(err, "invalid semantic version")
	})
}

func (s *UpdaterSuite) TestMissingBinaryInArchive() {
	missingArchive := tarGzipArchive(s.T(), "other", []byte("no"))
	missingPath := filepath.Join(s.T().TempDir(), "release.tar.gz")
	s.Require().NoError(os.WriteFile(missingPath, missingArchive, 0o600))
	err := extractBinary(missingPath, ".tar.gz", filepath.Join(s.T().TempDir(), "vizb"), "vizb")
	s.Require().ErrorContains(err, `binary "vizb" not found`)
}

func (s *UpdaterSuite) TestReplaceFailurePreservesTarget() {
	target := filepath.Join(s.T().TempDir(), "vizb")
	s.Require().NoError(os.WriteFile(target, []byte("old"), 0o755))

	err := replaceExecutable(filepath.Join(s.T().TempDir(), "missing"), target)
	s.Require().ErrorContains(err, "open verified replacement")
	got, readErr := os.ReadFile(target)
	s.Require().NoError(readErr)
	s.Equal("old", string(got))

	candidate := filepath.Join(s.T().TempDir(), "candidate")
	s.Require().NoError(os.WriteFile(candidate, []byte("new"), 0o755))
	err = replaceExecutable(candidate, s.T().TempDir())
	s.Require().ErrorContains(err, "not a regular file")
}

func standaloneUpdater(server *httptest.Server, executable, currentVersion, goos, goarch string) *Updater {
	return &Updater{
		currentVersion:      currentVersion,
		distribution:        "standalone",
		rawExecutable:       executable,
		canonicalExecutable: executable,
		goos:                goos,
		goarch:              goarch,
		client:              server.Client(),
		latestURL:           server.URL + "/latest",
		releaseDownloadURL:  server.URL + "/download",
		replace:             replaceExecutable,
	}
}

func newReleaseServer(t *testing.T, tag, checksumAsset string, archive []byte, corruptChecksum bool) *httptest.Server {
	t.Helper()
	digest := sha256.Sum256(archive)
	if corruptChecksum {
		digest = sha256.Sum256([]byte("different"))
	}
	checksums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(digest[:]), checksumAsset)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/latest":
			http.Redirect(w, r, "/releases/tag/"+tag, http.StatusFound)
		case r.URL.Path == "/releases/tag/"+tag:
			if r.Header.Get("User-Agent") == "" {
				t.Error("release request is missing User-Agent")
			}
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/download/"+tag+"/checksums.txt":
			_, _ = io.WriteString(w, checksums)
		case strings.HasPrefix(r.URL.Path, "/download/"+tag+"/"):
			_, _ = w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
}

func tarGzipArchive(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)
	if err := tarWriter.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func TestUpdaterSuite(t *testing.T) {
	suite.Run(t, new(UpdaterSuite))
}
