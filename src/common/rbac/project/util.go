package project

import "github.com/chenxull/goGridhub/gridhub/src/common/rbac"

var (
	// subresource policies for public project
	publicProjectPolicies = []*rbac.Policy{
		{Resource: rbac.ResourceSelf, Action: rbac.ActionRead},

		{Resource: rbac.ResourceLabel, Action: rbac.ActionRead},
		{Resource: rbac.ResourceLabel, Action: rbac.ActionList},

		{Resource: rbac.ResourceRepository, Action: rbac.ActionList},
		{Resource: rbac.ResourceRepository, Action: rbac.ActionPull},

		{Resource: rbac.ResourceRepositoryLabel, Action: rbac.ActionList},

		{Resource: rbac.ResourceRepositoryTag, Action: rbac.ActionRead},
		{Resource: rbac.ResourceRepositoryTag, Action: rbac.ActionList},

		{Resource: rbac.ResourceRepositoryTagLabel, Action: rbac.ActionList},

		{Resource: rbac.ResourceRepositoryTagVulnerability, Action: rbac.ActionList},

		{Resource: rbac.ResourceRepositoryTagManifest, Action: rbac.ActionRead},

		{Resource: rbac.ResourceHelmChart, Action: rbac.ActionRead},
		{Resource: rbac.ResourceHelmChart, Action: rbac.ActionList},

		{Resource: rbac.ResourceHelmChartVersion, Action: rbac.ActionRead},
		{Resource: rbac.ResourceHelmChartVersion, Action: rbac.ActionList},

		{Resource: rbac.ResourceScan, Action: rbac.ActionRead},
		{Resource: rbac.ResourceScanner, Action: rbac.ActionRead},
	}

	// all policies for the projects
	allPolicies = computeAllPolicies()
)

// PoliciesForPublicProject ...
func PoliciesForPublicProject(namespace rbac.Namespace) []*rbac.Policy {
	policies := []*rbac.Policy{}

	for _, policy := range publicProjectPolicies {
		policies = append(policies, &rbac.Policy{
			Resource: namespace.Resource(policy.Resource),
			Action:   policy.Action,
			Effect:   policy.Effect,
		})
	}

	return policies
}

// GetAllPolicies returns all policies for namespace of the project
func GetAllPolicies(namespace rbac.Namespace) []*rbac.Policy {
	policies := []*rbac.Policy{}

	for _, policy := range allPolicies {
		policies = append(policies, &rbac.Policy{
			Resource: namespace.Resource(policy.Resource),
			Action:   policy.Action,
			Effect:   policy.Effect,
		})
	}

	return policies
}

func computeAllPolicies() []*rbac.Policy {
	var results []*rbac.Policy

	mp := map[string]bool{}
	for _, policies := range rolePoliciesMap {
		for _, policy := range policies {
			if !mp[policy.String()] {
				results = append(results, policy)
				mp[policy.String()] = true
			}
		}
	}

	return results
}
