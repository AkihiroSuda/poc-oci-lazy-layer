package image

import (
	"flag"
	"testing"
)

var (
	outImage  string
	inImage   string
	ref       string
	removeOld bool
)

func init() {
	flag.StringVar(&outImage, "image.out", "", "path to output image dir")
	flag.StringVar(&inImage, "image.in", "", "path to input image dir")
	flag.StringVar(&ref, "image.ref", "latest", "image ref")
}

// usage: go test -v -image.in /tmp/in -image.out /tmp/out -image.ref latest .
func TestImage(t *testing.T) {
	if outImage == "" || inImage == "" {
		t.Skip("please specify -image.out and -image.in")
	}
	if err := Copy(outImage, inImage); err != nil {
		t.Fatal(err)
	}
	manifest, _, err := ReadManifest(outImage, ref)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Annotations == nil {
		manifest.Annotations = make(map[string]string)
	}
	t.Logf("annotating foo=bar")
	manifest.Annotations["foo"] = "bar"
	_, err = ReplaceManifest(outImage, ref, manifest)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("please verify %q manually (FIXME)", outImage)
}
