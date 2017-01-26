package prestat

import (
	"archive/tar"
	"encoding/json"
	"errors"
	"io"

	"github.com/AkihiroSuda/poc-oci-lazy-layer/proposal"
)

func ConvertTar(tw *tar.Writer, tr *tar.Reader) error {
	defer tw.Flush()
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		convertedHdr, prestatBytes, err := convertTarHeader(hdr)
		if err != nil {
			return err
		}
		if err = tw.WriteHeader(convertedHdr); err != nil {
			return err
		}
		if prestatBytes != nil {
			if _, err = tw.Write(prestatBytes); err != nil {
				return err
			}
		}
	}
}

func convertTarHeader(hdr *tar.Header) (*tar.Header, []byte, error) {
	convertedHdr := cloneTarHeader(hdr)
	if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA {
		return convertedHdr, nil, nil
	}
	prestatBytes, err := prestat(hdr)
	if err != nil {
		return nil, nil, err
	}
	convertedHdr.Size = int64(len(prestatBytes))
	return convertedHdr, prestatBytes, nil
}

func cloneTarHeader(hdr *tar.Header) *tar.Header {
	if hdr == nil {
		return nil
	}
	cloned := *hdr
	for k, v := range hdr.Xattrs {
		cloned.Xattrs[k] = v
	}
	return &cloned
}

func prestat(hdr *tar.Header) ([]byte, error) {
	if hdr == nil {
		return nil, errors.New("nil *tar.Header")
	}
	var s proposal.Prestat
	s.Size = hdr.Size
	return json.Marshal(s)
}
