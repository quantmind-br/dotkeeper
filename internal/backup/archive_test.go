package backup

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateArchive_Basic(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0755); err != nil {
		t.Fatal(err)
	}

	files, err := CollectFiles([]string{file1, file2}, nil)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := CreateArchive(files, &buf); err != nil {
		t.Fatalf("CreateArchive failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Archive is empty")
	}

	verifyArchive(t, &buf, files)
}

func TestCreateArchive_PreservesPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	file := filepath.Join(tmpDir, "executable.sh")
	if err := os.WriteFile(file, []byte("#!/bin/bash\necho hello"), 0755); err != nil {
		t.Fatal(err)
	}

	files, err := CollectFiles([]string{file}, nil)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := CreateArchive(files, &buf); err != nil {
		t.Fatalf("CreateArchive failed: %v", err)
	}

	gzr, err := gzip.NewReader(&buf)
	if err != nil {
		t.Fatal(err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	header, err := tr.Next()
	if err != nil {
		t.Fatal(err)
	}

	if header.Mode != 0755 {
		t.Errorf("Expected mode 0755, got %o", header.Mode)
	}
}

func TestCreateArchive_PreservesTimestamps(t *testing.T) {
	tmpDir := t.TempDir()

	file := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(file, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	modTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(file, modTime, modTime); err != nil {
		t.Fatal(err)
	}

	files, err := CollectFiles([]string{file}, nil)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := CreateArchive(files, &buf); err != nil {
		t.Fatalf("CreateArchive failed: %v", err)
	}

	gzr, err := gzip.NewReader(&buf)
	if err != nil {
		t.Fatal(err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	header, err := tr.Next()
	if err != nil {
		t.Fatal(err)
	}

	if !header.ModTime.Equal(modTime) {
		t.Errorf("Expected modtime %v, got %v", modTime, header.ModTime)
	}
}

func TestCreateArchive_Streaming(t *testing.T) {
	tmpDir := t.TempDir()

	largeFile := filepath.Join(tmpDir, "large.bin")
	largeContent := make([]byte, 10*1024*1024)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	if err := os.WriteFile(largeFile, largeContent, 0644); err != nil {
		t.Fatal(err)
	}

	files, err := CollectFiles([]string{largeFile}, nil)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := CreateArchive(files, &buf); err != nil {
		t.Fatalf("CreateArchive failed: %v", err)
	}

	gzr, err := gzip.NewReader(&buf)
	if err != nil {
		t.Fatal(err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	header, err := tr.Next()
	if err != nil {
		t.Fatal(err)
	}

	if header.Size != int64(len(largeContent)) {
		t.Errorf("Expected size %d, got %d", len(largeContent), header.Size)
	}

	extractedContent, err := io.ReadAll(tr)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(extractedContent, largeContent) {
		t.Error("Extracted content doesn't match original")
	}
}

func TestCreateArchive_EmptyFiles(t *testing.T) {
	files := []FileInfo{}

	var buf bytes.Buffer
	if err := CreateArchive(files, &buf); err != nil {
		t.Fatalf("CreateArchive failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Archive should have gzip header even with no files")
	}
}

func verifyArchive(t *testing.T, buf *bytes.Buffer, expectedFiles []FileInfo) {
	gzr, err := gzip.NewReader(buf)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	fileCount := 0
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to read tar header: %v", err)
		}

		if fileCount >= len(expectedFiles) {
			t.Errorf("More files in archive than expected")
			break
		}

		expected := expectedFiles[fileCount]

		if header.Name != expected.Path {
			t.Errorf("Expected name %s, got %s", expected.Path, header.Name)
		}

		if expected.LinkTarget != "" {
			if header.Typeflag != tar.TypeSymlink {
				t.Errorf("Expected symlink entry for %s", expected.Path)
			}
			if header.Linkname != expected.LinkTarget {
				t.Errorf("Expected linkname %s, got %s", expected.LinkTarget, header.Linkname)
			}
		} else if header.Size != expected.Size {
			t.Errorf("Expected size %d, got %d", expected.Size, header.Size)
		}

		fileCount++
	}

	if fileCount != len(expectedFiles) {
		t.Errorf("Expected %d files, got %d", len(expectedFiles), fileCount)
	}
}
