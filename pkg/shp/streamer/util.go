package streamer

import (
	"archive/tar"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func trimPrefix(prefix, fpath string) string {
	return strings.TrimPrefix(strings.Replace(fpath, prefix, "", -1), string(filepath.Separator))
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

	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(tw, f); err != nil {
		return err
	}
	return f.Close()
}
