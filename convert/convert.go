package convert

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"

	"github.com/AkihiroSuda/poc-oci-lazy-layer/image"
	"github.com/AkihiroSuda/poc-oci-lazy-layer/prestat"
	"github.com/AkihiroSuda/poc-oci-lazy-layer/proposal"
	"github.com/opencontainers/go-digest"
	spec "github.com/opencontainers/image-spec/specs-go/v1"
)

func convertManifest(oldManifest *spec.Manifest) *proposal.Manifest {
	var manifest proposal.Manifest
	manifest.SchemaVersion = oldManifest.SchemaVersion
	manifest.Config = oldManifest.Config
	manifest.Annotations = make(map[string]string)
	for _, l := range oldManifest.Layers {
		var layerDesc proposal.LayerDescriptor
		layerDesc.MediaType = l.MediaType
		layerDesc.Digest = l.Digest
		layerDesc.Size = l.Size
		layerDesc.URLs = l.URLs
		// this line is the new proposal.
		// the value is empty by default.
		layerDesc.PrestatDigest = ""
		manifest.Layers = append(manifest.Layers, layerDesc)
	}
	for k, v := range oldManifest.Annotations {
		manifest.Annotations[k] = v
	}
	return &manifest
}

func getLayerTarReader(img, mediaType string, d digest.Digest) (*tar.Reader, error) {
	if mediaType != spec.MediaTypeImageLayerGzip {
		return nil, fmt.Errorf("unsupported (yet) media type %q for layer %q",
			mediaType, d)
	}
	blobReader, err := image.GetBlobReader(img, d)
	if err != nil {
		return nil, err
	}
	gzReader, err := gzip.NewReader(blobReader)
	if err != nil {
		return nil, err
	}
	return tar.NewReader(gzReader), nil
}

func getLayerTarWriter(img, mediaType string, algo digest.Algorithm) (*tar.Writer, *gzip.Writer, *image.BlobWriter, error) {
	if mediaType != spec.MediaTypeImageLayerGzip {
		return nil, nil, nil, fmt.Errorf("unsupported (yet) media type", mediaType)
	}
	blobWriter, err := image.NewBlobWriter(img, algo)
	if err != nil {
		return nil, nil, nil, err
	}
	gzWriter := gzip.NewWriter(blobWriter)
	return tar.NewWriter(gzWriter), gzWriter, blobWriter, nil
}

func convertLayer(img string, l *proposal.LayerDescriptor) error {
	tarReader, err := getLayerTarReader(img, l.MediaType, l.Digest)
	if err != nil {
		return err
	}
	tarWriter, gzWriter, blobWriter, err := getLayerTarWriter(img, l.MediaType, l.Digest.Algorithm())
	if err != nil {
		return err
	}
	if err = prestat.ConvertTar(tarWriter, tarReader); err != nil {
		return err
	}
	tarWriter.Close()
	gzWriter.Close()
	blobWriter.Close()
	prestatDigest := blobWriter.Digest()
	if prestatDigest == nil {
		return fmt.Errorf("nil prestatDigest for layer %q", l.Digest)
	}
	l.PrestatDigest = *prestatDigest
	return nil
}

func convertImage(img, ref string) error {
	oldManifest, _, err := image.ReadManifest(img, ref)
	if err != nil {
		return err
	}
	manifest := convertManifest(oldManifest)
	for i, _ := range manifest.Layers {
		if err = convertLayer(img, &manifest.Layers[i]); err != nil {
			return err
		}
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	_, err = image.ReplaceManifestBytes(img, ref, manifestBytes)
	// TODO: remove old manifest if orphan
	return err
}

func ConvertImage(dst, src, ref string) error {
	if err := image.Copy(dst, src); err != nil {
		return err
	}
	return convertImage(dst, ref)
}
