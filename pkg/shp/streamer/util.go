package streamer

import (
	"archive/tar"
	"fmt"
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

func writeFileToTar(tw *tar.Writer, t *Tar, fpath string, stat fs.FileInfo) error {
	header, err := tar.FileInfoHeader(stat, stat.Name())
	if err != nil {
		return err
	}

	header.Name = trimPrefix(t.src, fpath)

	// Symlink
	if stat.Mode()&fs.ModeSymlink != 0 {
		target, err := os.Readlink(fpath)
		if err != nil {
			return err
		}

		// resolving target to absolute path, to determine if it is pointing
		// outside the source directory.
		var absSrc, absTarget string

		if absSrc, err = filepath.Abs(t.src); err != nil {
			return err
		}

		if filepath.IsAbs(target) {
			absTarget = target
		} else {
			fullTarget := filepath.Join(filepath.Dir(fpath), target)
			if absTarget, err = filepath.Abs(fullTarget); err != nil {
				return err
			}
		}

		if outside, _ := isSymlinkTargetOutsideOfDir(absSrc, absTarget); outside {
			relPath, _ := filepath.Rel(t.src, fpath)
			return fmt.Errorf("symlink %q points outside the source directory %q (target: %q)", relPath, absSrc, target)
		}

		header.Linkname = target
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	// Copy regular file content
	if stat.Mode().IsRegular() {
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

	return nil
}

// isSymlinkTargetOutsideOfDir checks if the provided link target resolves to an absolute path outside
// of the provided directory. Symlinks to absolute paths outside of a given "root" directory is a common
// attack vector, as these can potentially bypass operating system permission checks that block access to
// sensitive data. See [CWE-61](https://cwe.mitre.org/data/definitions/61.html).
func isSymlinkTargetOutsideOfDir(dir, target string) (bool, error) {
	realTarget, err := filepath.EvalSymlinks(target)
	if err != nil {
		realTarget = filepath.Clean(target)
	}

	rel, err := filepath.Rel(dir, realTarget)
	if err != nil {
		return false, err
	}

	return !filepath.IsLocal(rel), nil
}
