// Package imagerefs ...
package imagerefs

import (
	"fmt"
	"strings"

	"github.com/distribution/distribution/v3/reference"
	"github.com/opencontainers/go-digest"
)

// Parse ...
func Parse(name string) (string, string, string, string, error) {
	if !strings.Contains(name, "/") {
		return "", "", "", "", fmt.Errorf("invalid reference: %s", name)
	}

	named, err := reference.ParseNormalizedNamed(name)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to parse reference: %v, %s", err, name)
	}
	named = reference.TagNameOnly(named)
	domain := reference.Domain(named)
	path := reference.Path(named)
	tagged, ok := named.(reference.Tagged)
	if !ok {
		return "", "", "", "", fmt.Errorf("reference is not tagged: %v, %s", named, name)
	}
	tag := tagged.Tag()
	if !strings.Contains(path, "/") {
		return "", "", "", "", fmt.Errorf("invalid reference: %s", name)
	}
	parts := strings.Split(path, "/")
	ns := parts[0]
	repo := path
	return domain, ns, repo, tag, nil
}

// Reference ...
type Reference struct {
	isTag  bool
	tag    string
	digest digest.Digest
}

// NewReference ...
func NewReference(ref string) (*Reference, error) {
	isTag := false
	tag := ""
	dgest, err := digest.Parse(ref)
	if err != nil {
		isTag = true
		tag = ref
		if !reference.TagRegexp.MatchString(ref) {
			return nil, fmt.Errorf("not valid digest or tag")
		}
	}
	return &Reference{
		isTag:  isTag,
		tag:    tag,
		digest: dgest,
	}, nil
}

func (r *Reference) IsTag() bool {
	return r.isTag
}

func (r *Reference) Tag() string {
	return r.tag
}

func (r *Reference) Digest() digest.Digest {
	return r.digest
}
