package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/AkihiroSuda/poc-oci-lazy-layer/convert"
)

func main() {
	outImage := flag.String("out", "", "path to output image")
	inImage := flag.String("in", "", "path to input image")
	ref := flag.String("ref", "latest", "ref (e.g. \"latest\")")
	flag.Parse()
	if err := xmain(*outImage, *inImage, *ref, flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("converted %q to %q\n", *inImage, *outImage)
}

func xmain(outImage, inImage, ref string, args []string) error {
	if outImage == "" || inImage == "" || ref == "" {
		return errors.New("please specify flags")
	}
	if len(args) != 0 {
		return fmt.Errorf("extra args: %v", args)
	}
	return convert.ConvertImage(outImage, inImage, ref)
}
