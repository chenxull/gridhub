package filter

import (
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/logger"
	"github.com/chenxull/goGridhub/gridhub/src/replication/util"
)

// const definitions
const (
	FilterableTypeRepository = "repository"
	FilterableTypeVTag       = "vtag"
)

// FilterableType specifies the type of the filterable
type FilterableType string

// Filterable defines the methods that a filterable object must implement
type Filterable interface {
	// return what the type of the filterable object is(repository or vtag)
	GetFilterableType() FilterableType
	// return the resource type of the filterable object(image, chart, ...)
	GetResourceType() string
	GetName() string
	GetLabels() []string
}

// Filter defines the methods that a filter must implement
type Filter interface {
	ApplyTo(Filterable) bool
	Filter(...Filterable) ([]Filterable, error)
}

// NewResourceTypeFilter return a Filter to filter candidates according to the resource type
func NewResourceTypeFilter(resourceType string) Filter {
	return &resourceTypeFilter{
		resourceType: resourceType,
	}
}

// NewRepositoryNameFilter return a Filter to filter the repositories according to the name
func NewRepositoryNameFilter(pattern string) Filter {
	return &nameFilter{
		filterableType: FilterableTypeRepository,
		pattern:        pattern,
	}
}

// NewVTagNameFilter return a Filter to filter the vtags according to the name
func NewVTagNameFilter(pattern string) Filter {
	return &nameFilter{
		filterableType: FilterableTypeVTag,
		pattern:        pattern,
	}
}

// NewVTagLabelFilter return a Filter to filter vtags according to the label
func NewVTagLabelFilter(labels []string) Filter {
	return &labelFilter{
		labels: labels,
	}
}

type resourceTypeFilter struct {
	resourceType string
}

func (r *resourceTypeFilter) ApplyTo(filterable Filterable) bool {
	if filterable == nil {
		return false
	}
	switch filterable.GetFilterableType() {
	case FilterableTypeRepository, FilterableTypeVTag:
		return true
	default:
		return false
	}
}

func (r *resourceTypeFilter) Filter(filterables ...Filterable) ([]Filterable, error) {
	result := []Filterable{}
	for _, filterable := range filterables {
		if filterable.GetResourceType() == r.resourceType {
			result = append(result, filterable)
		}
	}
	return result, nil

}

type nameFilter struct {
	filterableType FilterableType
	pattern        string
}

func (n *nameFilter) ApplyTo(filterable Filterable) bool {
	if filterable == nil {
		return false
	}
	if filterable.GetFilterableType() == n.filterableType {
		return true
	}
	return false
}

func (n *nameFilter) Filter(filterables ...Filterable) ([]Filterable, error) {
	result := []Filterable{}
	for _, filterable := range filterables {
		name := filterable.GetName()
		match, err := util.Match(n.pattern, name)
		if err != nil {
			return nil, err
		}
		if match {
			logger.Debugf("%q matches the pattern %q of name filter", name, n.pattern)
			result = append(result, filterable)
			continue
		}
		logger.Debugf("%q doesn't match the pattern %q of name filter, skip", name, n.pattern)
	}
	return result, nil
}

type labelFilter struct {
	labels []string
}

func (l *labelFilter) ApplyTo(filterable Filterable) bool {
	if filterable == nil {
		return false
	}
	if filterable.GetFilterableType() == FilterableTypeVTag {
		return true
	}
	return false
}
func (l *labelFilter) Filter(filterables ...Filterable) ([]Filterable, error) {
	// if no specified label in the filter, just returns the input filterable
	// candidate as the result
	if len(l.labels) == 0 {
		return filterables, nil
	}
	result := []Filterable{}
	for _, filterable := range filterables {
		labels := map[string]struct{}{}
		for _, label := range filterable.GetLabels() {
			labels[label] = struct{}{}
		}
		match := true
		for _, label := range l.labels {
			if _, exist := labels[label]; !exist {
				match = false
				break
			}
		}
		// add the filterable to the result list if it contains
		// all labels defined for the filter
		if match {
			result = append(result, filterable)
		}
	}
	return result, nil
}

// DoFilter is a util function to help filter filterables easily.
// The parameter "filterables" must be a pointer points to a slice
// whose elements must be Filterable. After applying all the "filters"
// to the "filterables", the result is put back into the variable
// "filterables"
