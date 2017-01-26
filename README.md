# [PROPOSAL] Lazy-loadable layer support for OCI image spec

TLDR: Let files appear before pulling layers

## Proposed change to the image spec

Please refer to [`proposal/proposal.go`](proposal/proposal.go).

```go
package proposal

import (
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	spec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Prestat is a type newly added for my proposal
type Prestat struct {
	// Size is the size in bytes of the corresponding file in the layer.
	Size int64 `json:"size"`
}

// LayerDescriptor is a type that denotes the diff between the upstream spec and my proposal
type LayerDescriptor struct {
	spec.Descriptor
	// PrestatDigest is an OPTIONAL digest value of the "prestat" file of the layer.
	//
	// A "prestat" file MUST have the same mediatype, tar structure, digest algorithm.
	// A regular file in the "prestat" tar MUST be a JSON contains a valid Prestat JSON string.
	//
	// <TLDR> Using the "prestat" tar, a runtime can let files appear before pulling the corresponding layer.</TLDR>
	PrestatDigest digest.Digest `json:"prestat_digest,omitempty"`
}

// Manifest is a type that denotes the diff between the upstream spec and my proposal
type Manifest struct {
	specs.Versioned
	Config spec.Descriptor `json:"config"`
	// Layers is changed with the proposal. other fields are unchanged.
	Layers      []LayerDescriptor `json:"layers"`
	Annotations map[string]string `json:"annotations,omitempty"`
}
```

## POC (Converter)

Suppose that you have already converted a Docker image `busybox@sha256:817a12c32a39bbe394944ba49de563e085f1d3c5266eb8e9723256bc4448680e` to an OCI image `/tmp/image-busybox`. (via skopeo or something else)

Then convert the image via `poc-converter`.
```console
$ go run ./cmd/poc-converter/main.go -in /tmp/image-busybox -out /tmp/image-out -ref latest
converted "/tmp/image-busybox" to "/tmp/image-out"
```

You can see that the manifest contains a `prestat_digest` value:
```console
$ jq . < /tmp/image-out/refs/latest
{
    "mediaType": "application/vnd.oci.image.manifest.v1+json",
    "digest": "sha256:cb8472e2cf25fade21595a0f98821e62982a58e2e65e72e85a52ae2b85bf2adf",
    "size": 439
}
$ jq . < /tmp/image-out/blobs/sha256/cb8472e2cf25fade21595a0f98821e62982a58e2e65e72e85a52ae2b85bf2adf
{
    "schemaVersion": 2,
    "config": {
        "mediaType": "application/vnd.oci.image.config.v1+json",
	"digest": "sha256:7968321274dc6b6171697c33df7815310468e694ac5be0ec03ff053bb135e768",
	"size": 1465
    },
    "layers": [
	{
	    "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
	    "digest": "sha256:4b0bc1c4050b03c95ef2a8e36e25feac42fd31283e8c30b3ee5df6b043155d3c",
	    "size": 677628,
	    "prestat_digest": "sha256:4e53c89b7dabb5e7883ccbb6f9600f009521ef02d40ce9909ee5610f8e3af7d4"
	}
    ]
}
```

A regular file in the prestart archive `4e53c...` contains the `size` of the corresponding file in the layer archive `4b0bc...`.
```console
$ tar tzvf /tmp/image-out/blobs/sha256/4b0bc1c4050b03c95ef2a8e36e25feac42fd31283e8c30b3ee5df6b043155d3c bin/[
-rwxr-xr-x 0/0         1028368 2017-01-12 18:10 bin/[
$ tar tzvf /tmp/image-out/blobs/sha256/4e53c89b7dabb5e7883ccbb6f9600f009521ef02d40ce9909ee5610f8e3af7d4 bin/[
-rwxr-xr-x 0/0              16 2017-01-12 18:10 bin/[
$ tar xzOf /tmp/image-out/blobs/sha256/4e53c89b7dabb5e7883ccbb6f9600f009521ef02d40ce9909ee5610f8e3af7d4 bin/[
{"size":1028368}
```

## POC (runtime RootFS)

Not implemented yet

Plan:

 * FUSE
 * Read-only for ease of implementation
 * Emulate slow network so as to show the effect of lazy layer distribution

## Similar work

- [Harter, Tyler, et al. "Slacker: Fast Distribution with Lazy Docker Containers." FAST. 2016.](https://www.usenix.org/conference/fast16/technical-sessions/presentation/harter)
- [Lestaris, George. "Alternatives to layer-based image distribution: using CERN filesystem for images." Container Camp UK. 2016.](http://www.slideshare.net/glestaris/alternatives-to-layerbased-image-distribution-using-cern-filesystem-for-images)
- [Blomer, Jakob, et al. "A Novel Approach for Distributing and Managing Container Images: Integrating CernVM File System and Mesos." MesosCon NA. 2016.](https://mesosconna2016.sched.com/event/6jtr/a-novel-approach-for-distributing-and-managing-container-images-integrating-cernvm-file-system-and-mesos-jakob-blomer-cern-jie-yu-artem-harutyunyan-mesosphere)
