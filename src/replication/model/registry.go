package model

import (
	"github.com/chenxull/goGridhub/gridhub/src/common/models"
	"time"
)

const (
	RegistryTypeHarbor RegistryType = "harbor"
)

// RegistryType indicates the type of registry
type RegistryType string

// CredentialType represents the supported credential types
// e.g: u/p, OAuth token
type CredentialType string

// const definitions
const (
	// CredentialTypeBasic indicates credential by user name, password
	CredentialTypeBasic = "basic"
	// CredentialTypeOAuth indicates credential by OAuth token
	CredentialTypeOAuth = "oauth"
	// CredentialTypeSecret is only used by the communication of Harbor internal components
	CredentialTypeSecret = "secret"
)

// Credential keeps the access key and/or secret for the related registry
type Credential struct {
	// Type of the credential
	Type CredentialType `json:"type"`
	// The key of the access account, for OAuth token, it can be empty
	AccessKey string `json:"access_key"`
	// The secret or password for the key
	AccessSecret string `json:"access_secret"`
}

// HealthStatus describes whether a target is healthy or not
type HealthStatus string

const (
	// Healthy indicates registry is healthy
	Healthy = "healthy"
	// Unhealthy indicates registry is unhealthy
	Unhealthy = "unhealthy"
	// Unknown indicates health status of registry is unknown
	Unknown = "unknown"
)

// Registry keeps the related info of registry
type Registry struct {
	ID          int64        `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Type        RegistryType `json:"type"`
	URL         string       `json:"type"`
	// TokenServiceURL is only used for local harbor instance to
	// avoid the requests passing through the external proxy for now
	TokenServiceURL string      `json:"token_service_url"`
	Credential      *Credential `json:"credential"`
	Insecure        bool        `json:"insecure"`
	Status          string      `json:"status"`
	CreationTime    time.Time   `json:"creation_time"`
	UpdateTime      time.Time   `json:"update_time"`
}

// RegistryQuery defines the query conditions for listing registries
type RegistryQuery struct {
	// Name is name of the registry to query
	Name string
	// Pagination specifies the pagination
	Pagination *models.Pagination
}

// FilterStyle ...
type FilterStyle struct {
	Type   FilterType `json:"type"`
	Style  string     `json:"style"`
	Values []string   `json:"values,omitempty"`
}

// EndpointPattern ...
type EndpointPattern struct {
	EndpointType EndpointType `json:"endpoint_type"`
	Endpoints    []*Endpoint  `json:"endpoints"`
}

// EndpointType ..
type EndpointType string

const (
	// EndpointPatternTypeStandard ...
	EndpointPatternTypeStandard EndpointType = "EndpointPatternTypeStandard"
	// EndpointPatternTypeFix ...
	EndpointPatternTypeFix EndpointType = "EndpointPatternTypeFix"
	// EndpointPatternTypeList ...
	EndpointPatternTypeList EndpointType = "EndpointPatternTypeList"
)

// Endpoint ...
type Endpoint struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CredentialPattern ...
type CredentialPattern struct {
	AccessKeyType    AccessKeyType    `json:"access_key_type"`
	AccessKeyData    string           `json:"access_key_data"`
	AccessSecretType AccessSecretType `json:"access_secret_type"`
	AccessSecretData string           `json:"access_secret_data"`
}

// AccessKeyType ..
type AccessKeyType string

const (
	// AccessKeyTypeStandard ...
	AccessKeyTypeStandard AccessKeyType = "AccessKeyTypeStandard"
	// AccessKeyTypeFix ...
	AccessKeyTypeFix AccessKeyType = "AccessKeyTypeFix"
)

// AccessSecretType ...
type AccessSecretType string

const (
	// AccessSecretTypeStandard ...
	AccessSecretTypeStandard AccessSecretType = "AccessSecretTypePass"
	// AccessSecretTypeFile ...
	AccessSecretTypeFile AccessSecretType = "AccessSecretTypeFile"
)

// RegistryInfo provides base info and capability declarations of the registry
type RegistryInfo struct {
	Type                     RegistryType   `json:"type"`
	Description              string         `json:"description"`
	SupportedResourceTypes   []ResourceType `json:"-"`
	SupportedResourceFilters []*FilterStyle `json:"supported_resource_filters"`
	SupportedTriggers        []TriggerType  `json:"supported_triggers"`
}

// AdapterPattern provides base info and capability declarations of the registry
type AdapterPattern struct {
	EndpointPattern   *EndpointPattern   `json:"endpoint_pattern"`
	CredentialPattern *CredentialPattern `json:"credential_pattern"`
}

// NewDefaultAdapterPattern ...
func NewDefaultAdapterPattern() *AdapterPattern {
	return &AdapterPattern{
		EndpointPattern:   NewDefaultEndpointPattern(),
		CredentialPattern: NewDefaultCredentialPattern(),
	}
}

// NewDefaultEndpointPattern ...
func NewDefaultEndpointPattern() *EndpointPattern {
	return &EndpointPattern{
		EndpointType: EndpointPatternTypeStandard,
	}
}

// NewDefaultCredentialPattern ...
func NewDefaultCredentialPattern() *CredentialPattern {
	return &CredentialPattern{
		AccessKeyType:    AccessKeyTypeStandard,
		AccessSecretType: AccessSecretTypeStandard,
	}
}
