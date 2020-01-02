package rbac

// const action variables
const (
	ActionAll = Action("*") // action match any other actions

	ActionPull = Action("pull") // pull repository tag
	ActionPush = Action("push") // push repository tag

	// create, read, update, delete, list actions compatible with restful api methods
	ActionCreate = Action("create")
	ActionRead   = Action("read")
	ActionUpdate = Action("update")
	ActionDelete = Action("delete")
	ActionList   = Action("list")

	ActionOperate     = Action("operate")
	ActionScannerPull = Action("scanner-pull") // for robot account created by scanner to pull image, bypass the policy check
)

// const resource variables
const (
	ResourceAll                        = Resource("*")             // resource match any other resources
	ResourceConfiguration              = Resource("configuration") // project configuration compatible for portal only
	ResourceHelmChart                  = Resource("helm-chart")
	ResourceHelmChartVersion           = Resource("helm-chart-version")
	ResourceHelmChartVersionLabel      = Resource("helm-chart-version-label")
	ResourceLabel                      = Resource("label")
	ResourceLabelResource              = Resource("label-resource")
	ResourceLog                        = Resource("log")
	ResourceMember                     = Resource("member")
	ResourceMetadata                   = Resource("metadata")
	ResourceQuota                      = Resource("quota")
	ResourceReplication                = Resource("replication")     // TODO remove
	ResourceReplicationJob             = Resource("replication-job") // TODO remove
	ResourceReplicationExecution       = Resource("replication-execution")
	ResourceReplicationTask            = Resource("replication-task")
	ResourceRepository                 = Resource("repository")
	ResourceTagRetention               = Resource("tag-retention")
	ResourceImmutableTag               = Resource("immutable-tag")
	ResourceRepositoryLabel            = Resource("repository-label")
	ResourceRepositoryTag              = Resource("repository-tag")
	ResourceRepositoryTagLabel         = Resource("repository-tag-label")
	ResourceRepositoryTagManifest      = Resource("repository-tag-manifest")
	ResourceRepositoryTagScanJob       = Resource("repository-tag-scan-job")      // TODO: remove
	ResourceRepositoryTagVulnerability = Resource("repository-tag-vulnerability") // TODO: remove
	ResourceRobot                      = Resource("robot")
	ResourceNotificationPolicy         = Resource("notification-policy")
	ResourceScan                       = Resource("scan")
	ResourceScanner                    = Resource("scanner")
	ResourceSelf                       = Resource("") // subresource for self
)
