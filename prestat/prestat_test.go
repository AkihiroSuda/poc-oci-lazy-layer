package prestat

import (
	"archive/tar"
	"flag"
	"os"
	"testing"
)

var (
	outTar string
	inTar  string
)

func init() {
	flag.StringVar(&outTar, "prestat.out", "", "path to output tar (not tar+gzip)")
	flag.StringVar(&inTar, "prestat.in", "", "path to input tar (not tar+gzip)")
}

func openTars(t *testing.T) (*tar.Writer, *tar.Reader) {
	if outTar == "" || inTar == "" {
		t.Skip("please specify -prestat.out and -prestat.in")
	}
	outFile, err := os.Create(outTar)
	if err != nil {
		t.Fatal(err)
	}
	inFile, err := os.Open(inTar)
	if err != nil {
		t.Fatal(err)
	}
	return tar.NewWriter(outFile), tar.NewReader(inFile)
}

// usage: go test -v -prestat.in /tmp/in.tar -prestat.out /tmp/out.tar .
func TestConvertTar(t *testing.T) {
	if err := ConvertTar(openTars(t)); err != nil {
		t.Fatal(err)
	}
	t.Logf("please verify %q manually (FIXME)", outTar)
}
