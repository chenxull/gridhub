package rbac

import "fmt"

// Namespace the namespace interface
type Namespace interface {
	// Kind returns the kind of namespace
	Kind() string
	// Resource returns new resource for subresources with the namespace
	Resource(subresources ...Resource) Resource
	// Identity returns identity attached with namespace
	Identity() interface{}
	// IsPublic returns true if namespace is public
	IsPublic() bool
}

type projectNamespace struct {
	projectID int64
	isPublic  bool
}

func (ns *projectNamespace) Kind() string {
	return "project"
}

func (ns *projectNamespace) Resource(subresources ...Resource) Resource {
	return Resource(fmt.Sprintf("/project/%d", ns.projectID)).Subresource(subresources...)
}

func (ns *projectNamespace) Identity() interface{} {
	return ns.projectID
}

func (ns *projectNamespace) IsPublic() bool {
	return ns.isPublic
}

func NewProjectNamespace(projectID int64, isPublic ...bool) Namespace {
	isPublicNamespace := false
	if len(isPublic) > 0 {
		isPublicNamespace = isPublic[0]
	}
	return &projectNamespace{projectID: projectID, isPublic: isPublicNamespace}
}
