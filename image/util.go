package image

import (
	"log"
	"os/exec"
)

func copyDirectory(dst, src string) error {
	// terrible implementation, but it does not hurt for POC...
	out, err := exec.Command("rm", "-rf", dst).CombinedOutput()
	if len(out) != 0 {
		log.Printf("output from `rm -rf %s`: %s", dst, out)
	}
	if err != nil {
		return err
	}
	out, err = exec.Command("cp", "-a", src, dst).CombinedOutput()
	if len(out) != 0 {
		log.Printf("output from `cp -a %s %s`: %s", src, dst, out)
	}
	return err
}
