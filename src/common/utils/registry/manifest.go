package registry

import "github.com/docker/distribution"

//UnMarshal converts []byte to be Manifest
func UnMarshal(mediaType string, data []byte) (distribution.Manifest, distribution.Descriptor, error) {
	return distribution.UnmarshalManifest(mediaType, data)
}
