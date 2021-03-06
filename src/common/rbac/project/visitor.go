package project

import "github.com/chenxull/goGridhub/gridhub/src/common/rbac"

// visitorContext the context interface for the project visitor
type visitorContext interface {
	IsAuthenticated() bool
	// GetUsername returns the username of user related to the context
	GetUsername() string
	// IsSysAdmin returns whether the user is system admin
	IsSysAdmin() bool
}

// visitor implement the rbac.User interface for project visitor
type visitor struct {
	ctx          visitorContext
	namespace    rbac.Namespace
	projectRoles []int
}

// GetUserName returns username of the visitor
func (v *visitor) GetUserName() string {
	// anonymous username for unauthenticated Visitor
	if !v.ctx.IsAuthenticated() {
		return "anonymous"
	}

	return v.ctx.GetUsername()
}

// GetPolicies returns policies of the visitor
func (v *visitor) GetPolicies() []*rbac.Policy {
	if v.ctx.IsSysAdmin() {
		return GetAllPolicies(v.namespace)
	}

	if v.namespace.IsPublic() {
		return PoliciesForPublicProject(v.namespace)
	}

	return nil
}

// GetRoles returns roles of the visitor
func (v *visitor) GetRoles() []rbac.Role {
	// Ignore roles when visitor is anonymous or system admin
	if !v.ctx.IsAuthenticated() || v.ctx.IsSysAdmin() {
		return nil
	}

	roles := []rbac.Role{}

	for _, roleID := range v.projectRoles {
		roles = append(roles, &visitorRole{roleID: roleID, namespace: v.namespace})
	}

	return roles
}

// NewUser returns rbac.User interface for the project visitor
func NewUser(ctx visitorContext, namespace rbac.Namespace, projectRoles ...int) rbac.User {
	return &visitor{
		ctx:          ctx,
		namespace:    namespace,
		projectRoles: projectRoles,
	}
}
