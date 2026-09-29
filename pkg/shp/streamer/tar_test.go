package streamer

import (
	"archive/tar"
	"io"
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
