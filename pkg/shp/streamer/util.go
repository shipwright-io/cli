package streamer

import (
	"archive/tar"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type writeCounter struct{ total int }

func (wc *writeCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.total += n
	return n, nil
}

func trimPrefix(prefix, fpath string) string {
	return strings.TrimPrefix(strings.ReplaceAll(fpath, prefix, ""), string(filepath.Separator))
}

func writeFileToTar(tw *tar.Writer, src, fpath string, stat fs.FileInfo) error {
	header, err := tar.FileInfoHeader(stat, stat.Name())
	if err != nil {
		return err
	}

	header.Name = trimPrefix(src, fpath)
	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	// #nosec G304 intentionally opening file from variable
	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(tw, f); err != nil {
		return err
	}
	return f.Close()
}
