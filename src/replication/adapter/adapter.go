package adapter

import (
	"github.com/chenxull/goGridhub/gridhub/src/replication/model"
	"github.com/docker/distribution"
	"io"
)

// const definition
const (
	UserAgentReplication = "harbor-replication-service"
	MaxConcurrency       = 100
)

type Factory interface {
	Create(*model.Registry) (Adapter, error)
	AdapterPattern() *model.AdapterPattern
}

// Adapter interface defines the capabilities of registry
type Adapter interface {
	//Info return the information of this adapter
	Info() (*model.RegistryInfo, error)
	// PrepareForPush does the prepare work that needed for pushing/uploading the resources
	// eg: create the namespace or repository
	PrepareForPush([]*model.Resource) error
	// HealthCheck checks health status of registry
	HealthCheck() (model.HealthStatus, error)
}

// ImageRegistry defines the capabilities that an image registry should have
type ImageRegistry interface {
	FetchImage(filters []*model.Filter) ([]*model.Resource, error)
	ManifestExist(repository, reference string) (exist bool, digest string, err error)
	PullManifest(repository, reference string, accepttedMediaTypes []string) (manifest distribution.Manifest, digest string, err error)
	PushManifest(repository, reference, mediaType string, payload []byte) error

	// the "reference" can be "tag" or "digest", the function needs to handle both
	DeleteManifest(repository, reference string) error
	BlobExist(repository, digest string) (exist bool, err error)
	PullBlob(repository, digest string) (size int64, blob io.ReadCloser, err error)
	PushBlob(repository, digest string, size int64, blob io.Reader) error
}

// ChartRegistry defines the capabilities that a chart registry should have
type ChartRegistry interface {
	FetchCharts(filters []*model.Filter) ([]*model.Resource, error)
	ChartExist(name, version string) (bool, error)
	DownloadChart(name, version string) (io.ReadCloser, error)
	UploadChart(name, version string, chart io.Reader) error
	DeleteChart(name, version string) error
}

// Repository defines an repository object, it can be image repository, chart repository and etc.
type Repository struct {
	ResourceType string `json:"resource_type"`
	Name         string `json:"name"`
}

// GetName returns the name
func (r *Repository) GetName() string {
	return r.Name
}

// GetResourceType returns the resource type
func (r *Repository) GetResourceType() string {
	return r.ResourceType
}
