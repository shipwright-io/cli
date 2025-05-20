package streamer

import (
	"archive/tar"
	"io"
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

		// making sure that undesired entries are not present on the list of files caputured by the
		// tar helper
		g.Expect(strings.HasPrefix(name, ".git/")).To(o.BeFalse())
		g.Expect(strings.HasPrefix(name, "_output/")).To(o.BeFalse())
	}
	g.Expect(counter > 10).To(o.BeTrue())
}
