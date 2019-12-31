package modifier

import "net/http"

type Modifier interface {
	Modify(r *http.Request) error
}
