package streamer

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	o "github.com/onsi/gomega"
)

func Test_Tar(t *testing.T) {
	g := o.NewGomegaWithT(t)

	tarHelper, err := NewTar("../../..")
	g.Expect(err).To(o.BeNil())

	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()

	go func() {
		err := tarHelper.Create(writer)
		g.Expect(err).To(o.BeNil())
	}()

	tarReader := tar.NewReader(reader)
	counter := 0
	foundGitIgnore := false
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			g.Expect(err).To(o.BeNil())
		}
		counter++
		name := header.Name

		cleanName := filepath.ToSlash(name)
		if cleanName == ".gitignore" {
			foundGitIgnore = true
		}

		// making sure that undesired entries are not present on the list of files caputured by the
		// tar helper
		g.Expect(strings.Split(cleanName, "/")).NotTo(o.ContainElement(".git"), "should not contain a .git path component")
		g.Expect(strings.HasPrefix(cleanName, "_output/")).To(o.BeFalse())
	}
	g.Expect(foundGitIgnore).To(o.BeTrue(), "expected .gitignore to be included in the tarball")
	g.Expect(counter > 10).To(o.BeTrue())
}

func Test_TarGitIgnoreRelativeToSource(t *testing.T) {
	g := o.NewGomegaWithT(t)

	// the source directory lives inside a directory named "src", which must not be matched by the
	// "src/" pattern, since git ignore patterns are relative to the directory holding .gitignore
	src := filepath.Join(t.TempDir(), "src", "app")
	files := map[string]string{
		".gitignore":          "src/\n/build\n*.log\n",
		"Dockerfile":          "FROM scratch\n",
		"main.go":             "package main\n",
		"debug.log":           "ignored\n",
		"build/output":        "ignored\n",
		"sub/build/output":    "kept\n",
		"sub/src/main.go":     "ignored\n",
		"sub/nested/info.log": "ignored\n",
	}
	for name, content := range files {
		fpath := filepath.Join(src, filepath.FromSlash(name))
		g.Expect(os.MkdirAll(filepath.Dir(fpath), 0o750)).To(o.Succeed())
		g.Expect(os.WriteFile(fpath, []byte(content), 0o600)).To(o.Succeed())
	}

	tarHelper, err := NewTar(src)
	g.Expect(err).To(o.BeNil())

	var buf bytes.Buffer
	g.Expect(tarHelper.Create(&buf)).To(o.Succeed())

	names := []string{}
	tarReader := tar.NewReader(&buf)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		g.Expect(err).To(o.BeNil())
		names = append(names, header.Name)
	}

	g.Expect(names).To(o.ConsistOf(
		".gitignore",
		"Dockerfile",
		"main.go",
		"sub/build/output",
	))
}
