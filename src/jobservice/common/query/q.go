package query

import "encoding/json"

const (
	// DefaultPageSize defines the default page size
	DefaultPageSize uint = 25
	// ParamKeyPage defines query param key of page number
	ParamKeyPage = "page_number"
	// ParamKeyPageSize defines query param key of page size
	ParamKeyPageSize = "page_size"
	// ParamKeyNonStoppedOnly defines query param key of querying non stopped periodic executions
	ParamKeyNonStoppedOnly = "non_dead_only"
	// ParamKeyCursor defines query param of cursor for fetching job stats with batches
	ParamKeyCursor = "cursor"
	// ParamKeyJobKind defines query param of job kind
	ParamKeyJobKind = "kind"
	// ExtraParamKeyNonStoppedOnly defines extra parameter key for querying non stopped periodic executions
	ExtraParamKeyNonStoppedOnly = "NonDeadOnly"
	// ExtraParamKeyCursor defines extra parameter key for the cursor of fetching job stats with batches
	ExtraParamKeyCursor = "Cursor"
	// ExtraParamKeyKind defines extra parameter key for the job kind
	ExtraParamKeyKind = "Kind"
)

// ExtraParameters to keep non pagination query parameters
type ExtraParameters map[string]interface{}

// Set extra parameters
func (ep ExtraParameters) Set(key string, v interface{}) {
	if len(key) > 0 {
		ep[key] = v
	}
}

// Get the extra parameter by key
func (ep ExtraParameters) Get(key string) (interface{}, bool) {
	v, ok := ep[key]

	return v, ok
}

// String returns the json string of ExtraParameters
func (ep ExtraParameters) String() string {
	bytes, err := json.Marshal(&ep)
	if err != nil {
		return ""
	}

	return string(bytes)
}

// Parameter for getting executions
type Parameter struct {
	PageNumber uint
	PageSize   uint
	Extras     ExtraParameters
}
