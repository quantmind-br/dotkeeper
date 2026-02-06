package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"time"
)

func CreateArchive(files []FileInfo, writer io.Writer) error {
	gzw := gzip.NewWriter(writer)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	for _, fileInfo := range files {
		if err := addFileToArchive(tw, fileInfo); err != nil {
			return fmt.Errorf("failed to add %s to archive: %w", fileInfo.Path, err)
		}
	}

	return nil
}

func addFileToArchive(tw *tar.Writer, fileInfo FileInfo) error {
	if fileInfo.LinkTarget != "" {
		header := &tar.Header{
			Typeflag: tar.TypeSymlink,
			Name:     fileInfo.Path,
			Linkname: fileInfo.LinkTarget,
			Mode:     int64(fileInfo.Mode),
			ModTime:  time.Unix(fileInfo.ModTime, 0),
		}
		return tw.WriteHeader(header)
	}

	file, err := os.Open(fileInfo.Path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	header := &tar.Header{
		Name:    fileInfo.Path,
		Size:    fileInfo.Size,
		Mode:    int64(fileInfo.Mode),
		ModTime: time.Unix(fileInfo.ModTime, 0),
	}

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	if _, err := io.Copy(tw, file); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}
