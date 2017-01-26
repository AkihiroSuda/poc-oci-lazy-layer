package image

import (
	_ "crypto/sha256"
	_ "crypto/sha512"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/opencontainers/go-digest"
	spec "github.com/opencontainers/image-spec/specs-go/v1"
)

func Copy(dst, src string) error {
	return copyDirectory(dst, src)
}

func blobPath(img string, d digest.Digest) string {
	return filepath.Join(img, "blobs", d.Algorithm().String(), d.Hex())
}

func GetBlobReader(img string, d digest.Digest) (io.Reader, error) {
	return os.Open(blobPath(img, d))
}

func ReadBlob(img string, d digest.Digest) ([]byte, error) {
	r, err := GetBlobReader(img, d)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(r)
}

func WriteBlob(img string, b []byte) (digest.Digest, error) {
	d := digest.FromBytes(b)
	return d, ioutil.WriteFile(blobPath(img, d), b, 0444)
}

type BlobWriter struct {
	img      string
	digester digest.Digester
	f        *os.File
	closed   bool
}

func NewBlobWriter(img string, algo digest.Algorithm) (*BlobWriter, error) {
	f, err := ioutil.TempFile("", "blobwriter")
	if err != nil {
		return nil, err
	}
	return &BlobWriter{
		img:      img,
		digester: algo.Digester(),
		f:        f,
	}, nil
}

func (bw *BlobWriter) Write(b []byte) (int, error) {
	n, err := bw.f.Write(b)
	if err != nil {
		return n, err
	}
	return bw.digester.Hash().Write(b)
}

func (bw *BlobWriter) Close() error {
	oldPath := bw.f.Name()
	if err := bw.f.Close(); err != nil {
		return err
	}
	newPath := blobPath(bw.img, bw.digester.Digest())
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}
	bw.closed = true
	return nil
}

// Digest returns nil if unclosed
func (bw *BlobWriter) Digest() *digest.Digest {
	if !bw.closed {
		return nil
	}
	d := bw.digester.Digest()
	return &d
}

func DeleteBlob(img string, d digest.Digest) error {
	return os.Remove(blobPath(img, d))
}

func refPath(img, ref string) string {
	return filepath.Join(img, "refs", ref)
}

func ReadManifestDescriptorBytes(img, ref string) ([]byte, error) {
	return ioutil.ReadFile(refPath(img, ref))
}

func ReadManifestDescriptor(img, ref string) (*spec.Descriptor, error) {
	b, err := ReadManifestDescriptorBytes(img, ref)
	if err != nil {
		return nil, err
	}
	var desc spec.Descriptor
	if err = json.Unmarshal(b, &desc); err != nil {
		return nil, err
	}
	return &desc, nil
}

func WriteManifestDescriptorBytes(img, ref string, b []byte) error {
	return ioutil.WriteFile(refPath(img, ref), b, 0644)
}

func WriteManifestDescriptor(img, ref string, desc *spec.Descriptor) error {
	b, err := json.Marshal(desc)
	if err != nil {
		return err
	}
	return WriteManifestDescriptorBytes(img, ref, b)
}

func ReadManifestBytes(img, ref string) ([]byte, digest.Digest, error) {
	desc, err := ReadManifestDescriptor(img, ref)
	if err != nil {
		return nil, "", err
	}
	b, err := ReadBlob(img, desc.Digest)
	// TODO: verify digest?
	return b, desc.Digest, err
}

func ReadManifest(img, ref string) (*spec.Manifest, digest.Digest, error) {
	b, d, err := ReadManifestBytes(img, ref)
	if err != nil {
		return nil, "", err
	}
	var manifest spec.Manifest
	if err = json.Unmarshal(b, &manifest); err != nil {
		return nil, "", err
	}
	return &manifest, d, err
}

func ReplaceManifestBytes(img, ref string, b []byte) (digest.Digest, error) {
	d, err := WriteBlob(img, b)
	if err != nil {
		return "", err
	}
	desc, err := ReadManifestDescriptor(img, ref)
	if err != nil {
		return "", err
	}
	desc.Digest = d
	desc.Size = int64(len(b))
	if err = WriteManifestDescriptor(img, ref, desc); err != nil {
		return "", err
	}
	return d, nil
}

func ReplaceManifest(img, ref string, manifest *spec.Manifest) (digest.Digest, error) {
	b, err := json.Marshal(manifest)
	if err != nil {
		return "", err
	}
	return ReplaceManifestBytes(img, ref, b)
}
