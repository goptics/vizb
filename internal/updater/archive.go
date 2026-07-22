package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const maxBinarySize = 256 << 20

func extractBinary(archivePath, extension, destination, binaryName string) error {
	switch extension {
	case ".tar.gz":
		return extractTarGzipBinary(archivePath, destination, binaryName)
	case ".zip":
		return extractZipBinary(archivePath, destination, binaryName)
	default:
		return fmt.Errorf("unsupported archive format %q", extension)
	}
}

func extractTarGzipBinary(archivePath, destination, binaryName string) error {
	archive, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open release archive: %w", err)
	}
	defer archive.Close()

	gzipReader, err := gzip.NewReader(archive)
	if err != nil {
		return fmt.Errorf("open gzip stream: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read release archive: %w", err)
		}
		if cleanArchiveName(header.Name) != binaryName || header.Typeflag != tar.TypeReg {
			continue
		}
		if header.Size < 0 || header.Size > maxBinarySize {
			return fmt.Errorf("binary in release archive is too large")
		}
		return writeExtractedBinary(destination, io.LimitReader(tarReader, header.Size), header.Size)
	}

	return fmt.Errorf("binary %q not found in release archive", binaryName)
}

func extractZipBinary(archivePath, destination, binaryName string) error {
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open release archive: %w", err)
	}
	defer archive.Close()

	for _, file := range archive.File {
		if cleanArchiveName(file.Name) != binaryName || !file.Mode().IsRegular() {
			continue
		}
		if file.UncompressedSize64 > maxBinarySize {
			return fmt.Errorf("binary in release archive is too large")
		}

		reader, err := file.Open()
		if err != nil {
			return fmt.Errorf("open binary in release archive: %w", err)
		}
		err = writeExtractedBinary(destination, reader, int64(file.UncompressedSize64))
		closeErr := reader.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return fmt.Errorf("close binary in release archive: %w", closeErr)
		}
		return nil
	}

	return fmt.Errorf("binary %q not found in release archive", binaryName)
}

func cleanArchiveName(name string) string {
	return strings.TrimPrefix(filepath.ToSlash(name), "./")
}

func writeExtractedBinary(destination string, source io.Reader, expectedSize int64) error {
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700)
	if err != nil {
		return fmt.Errorf("create extracted binary: %w", err)
	}

	written, copyErr := io.Copy(output, source)
	syncErr := output.Sync()
	closeErr := output.Close()
	if copyErr != nil {
		return fmt.Errorf("extract binary: %w", copyErr)
	}
	if written != expectedSize {
		return fmt.Errorf("extract binary: expected %d bytes, wrote %d", expectedSize, written)
	}
	if syncErr != nil {
		return fmt.Errorf("sync extracted binary: %w", syncErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close extracted binary: %w", closeErr)
	}
	return nil
}
