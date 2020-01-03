package model

import (
	"github.com/chenxull/goGridhub/gridhub/src/common/models"
	"time"
)

//const definition
const (
	FilterTypeResource FilterType = "resource"
	FilterTypeName     FilterType = "name"
	FilterTypeTag      FilterType = "tag"
	FilterTypeLabel    FilterType = "label"

	TriggerTypeManual     TriggerType = "manual"
	TriggerTypeScheduled  TriggerType = "scheduled"
	TriggerTypeEventBased TriggerType = "event_based"
)

type Policy struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Creator     string `json:"creator"`

	//source
	SrcRegistry *Registry `json:"src_registry"`
	//destination
	DestRegistry *Registry `json:"dest_registry"`
	// Only support two dest namespace modes:
	// Put all the src resources to the one single dest namespace
	// or keep namespaces same with the source ones (under this case,
	// the DestNamespace should be set to empty)
	DestNamespace string `json:"dest_namespace"`
	// Filters
	Filters []*Filter `json:"filters"`
	// Trigger
	Trigger *Trigger `json:"trigger"`
	// Settings
	// TODO: rename the property name
	Deletion bool `json:"deletion"`
	// If override the image tag
	Override bool `json:"override"`

	// Operations
	Enabled      bool      `json:"enabled"`
	CreationTime time.Time `json:"creation_time"`
	UpdateTime   time.Time `json:"update_time"`
}

// FilterType represents the type info of the filter.
type FilterType string

// Filter holds the info of the filter
type Filter struct {
	Type  FilterType  `json:"type"`
	Value interface{} `json:"value"`
}

// TriggerType represents the type of trigger.
type TriggerType string

// Trigger holds info for a trigger
type Trigger struct {
	Type     TriggerType      `json:"type"`
	Settings *TriggerSettings `json:"trigger_settings"`
}

// TriggerSettings is the setting about the trigger
type TriggerSettings struct {
	Cron string `json:"cron"`
}

// PolicyQuery defines the query conditions for listing policies
type PolicyQuery struct {
	Name string
	// TODO: need to consider how to support listing the policies
	// of one namespace in both pull and push modes
	Namespace    string
	SrcRegistry  int64
	DestRegistry int64
	models.Pagination
}
