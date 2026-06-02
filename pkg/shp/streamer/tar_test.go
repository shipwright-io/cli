package streamer

import (
	"archive/tar"
	"io"
	"os"
	"strings"
	"testing"

	o "github.com/onsi/gomega"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func Test_Tar(t *testing.T) {
	g := o.NewGomegaWithT(t)

	baseDir := "../../.."
	ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()

	tarHelper, err := NewTar(baseDir, &ioStreams)
	g.Expect(err).To(o.BeNil())

	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()

	// Ensure symlinks with a target outside the source tree are rejected
	symlink := baseDir + "/test/" + "symlink_for_test_tar"
	if err := os.Symlink("../../path/outside/source/tree", symlink); err != nil {
		t.Fatalf("Failed to setup test symlink pointing outside the source tree: %v", err)
	}
	t.Cleanup(func() {
		cleanupErr := os.Remove(symlink)
		if cleanupErr != nil {
			t.Logf("Failed to cleanup test symlink %q: %v", symlink, cleanupErr)
		}
	})

	err = tarHelper.Create(io.Discard)
	g.Expect(err.Error()).To(o.ContainSubstring("points outside the source directory"))

	// Override the symlink with a target inside the source tree and retest
	if err := os.Remove(symlink); err != nil {
		t.Fatalf("Failed to cleanup test symlink %q: %v", symlink, err)
	}
	if err := os.Symlink("../path/inside/source/tree", symlink); err != nil {
		t.Fatalf("Failed to setup test symlink pointing inside the source tree: %v", err)
	}

	go func() {
		err := tarHelper.Create(writer)
		g.Expect(err).To(o.BeNil())
	}()

	tarReader := tar.NewReader(reader)
	counter, symlinks := 0, 0
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

		if header.Linkname != "" {
			symlinks++
		}

		// making sure that undesired entries are not present on the list of files captured by the
		// tar helper
		g.Expect(strings.HasPrefix(name, ".git/")).To(o.BeFalse())
		g.Expect(strings.HasPrefix(name, "_output/")).To(o.BeFalse())
	}
	g.Expect(counter > 10).To(o.BeTrue())
	g.Expect(symlinks > 0).To(o.BeTrue())
}
