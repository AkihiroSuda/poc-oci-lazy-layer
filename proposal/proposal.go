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
