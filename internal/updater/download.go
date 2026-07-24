package updater

import (
	"bufio"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	maxArchiveSize  = 512 << 20
	maxChecksumSize = 1 << 20
)

func downloadFile(ctx context.Context, client *http.Client, sourceURL, destination, userAgent string, maxSize int64) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return fmt.Errorf("create download request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", sourceURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("download %s: server returned %s", sourceURL, resp.Status)
	}
	if resp.ContentLength > maxSize {
		return fmt.Errorf("download %s: response exceeds %d bytes", sourceURL, maxSize)
	}

	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create download destination: %w", err)
	}

	written, copyErr := io.Copy(output, io.LimitReader(resp.Body, maxSize+1))
	syncErr := output.Sync()
	closeErr := output.Close()
	if copyErr != nil {
		return fmt.Errorf("save download: %w", copyErr)
	}
	if written > maxSize {
		return fmt.Errorf("download %s: response exceeds %d bytes", sourceURL, maxSize)
	}
	if syncErr != nil {
		return fmt.Errorf("sync download: %w", syncErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close download: %w", closeErr)
	}
	return nil
}

func expectedChecksum(checksumsPath, assetName string) ([]byte, error) {
	checksums, err := os.Open(checksumsPath)
	if err != nil {
		return nil, fmt.Errorf("open checksums: %w", err)
	}
	defer checksums.Close()

	scanner := bufio.NewScanner(io.LimitReader(checksums, maxChecksumSize+1))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 || strings.TrimPrefix(fields[1], "*") != assetName {
			continue
		}
		digest, err := hex.DecodeString(fields[0])
		if err != nil || len(digest) != sha256.Size {
			return nil, fmt.Errorf("invalid SHA-256 checksum for %s", assetName)
		}
		return digest, nil
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read checksums: %w", err)
	}
	return nil, fmt.Errorf("checksum for %s not found", assetName)
}

func verifyChecksum(filePath string, expected []byte) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open downloaded archive: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("hash downloaded archive: %w", err)
	}
	actual := hash.Sum(nil)
	if subtle.ConstantTimeCompare(actual, expected) != 1 {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", hex.EncodeToString(expected), hex.EncodeToString(actual))
	}
	return nil
}
