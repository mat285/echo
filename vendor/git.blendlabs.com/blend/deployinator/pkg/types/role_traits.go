package types

import (
	"fmt"

	exception "github.com/blend/go-sdk/exception"
)

// Role is something that we are.
type Role string

const (
	// RoleSuperUser is a role.
	RoleSuperUser Role = "superuser"
	// RoleUser is a role
	RoleUser Role = "user"
	// RoleServiceAuditor is a role
	RoleServiceAuditor Role = "auditor"
	// RoleTeam is a role
	RoleTeam Role = "team"
	// RoleServiceOwner is a role.
	RoleServiceOwner Role = "service_owner"
	// RoleServiceUser is a role.
	RoleServiceUser Role = "service_user"
	// RoleProjectOwner is the project owner
	RoleProjectOwner Role = "project_owner"
	// RoleProjectUser is a project user
	RoleProjectUser Role = "project_user"
	// RoleDatabaseOwner is the project owner
	RoleDatabaseOwner Role = "database_owner"
	// RoleTeamAdmin is a role.
	RoleTeamAdmin Role = "team_admin"
	// RoleNamespaceOwner is a role
	RoleNamespaceOwner Role = "namespace_owner"
	// RoleNamespaceUser is a role
	RoleNamespaceUser Role = "namespace_user"
)

var (
	// AllRoles are all roles
	AllRoles = []Role{
		RoleSuperUser,
		RoleUser,
		RoleServiceAuditor,
		RoleTeam,
		RoleTeamAdmin,
		RoleServiceOwner,
		RoleServiceUser,
		RoleProjectOwner,
		RoleProjectUser,
		RoleDatabaseOwner,
		RoleNamespaceOwner,
		RoleNamespaceUser,
	}

	// ServiceRoles are roles that pertain to service usage or management
	ServiceRoles = []Role{
		RoleServiceOwner,
		RoleServiceUser,
		RoleServiceAuditor,
	}

	// ProjectRoles are roles that pertain to project usage or management
	ProjectRoles = []Role{
		RoleProjectOwner,
		RoleProjectUser,
	}

	// DatabaseRoles are roles that pertain to project usage or management
	DatabaseRoles = []Role{
		RoleDatabaseOwner,
	}

	// TeamRoles are the roles that pertain to team management
	TeamRoles = []Role{
		RoleTeamAdmin,
	}

	// NamespaceRoles are the roles that pertain to namespaces
	NamespaceRoles = []Role{
		RoleNamespaceOwner,
		RoleNamespaceUser,
	}

	// GlobalRoles are the roles that pertain to global
	GlobalRoles = []Role{
		RoleUser,
		RoleSuperUser,
		RoleTeam,
		RoleServiceAuditor,
	}

	// RoleTraits map roles to their traits.
	RoleTraits = map[Role][]Trait{
		RoleSuperUser: []Trait{TraitSuperUser},
		RoleUser:      []Trait{TraitServiceCreate, TraitDatabaseCreate, TraitTeamCreate, TraitNamespaceCreate},
		RoleTeam:      []Trait{TraitServiceCreate, TraitDatabaseCreate},
		RoleTeamAdmin: []Trait{
			TraitTeamManageUsers,
			TraitTeamManageTokens,
			TraitTeamViewTokens,
			TraitManageGrants},
		RoleServiceOwner: []Trait{
			TraitServiceDeploy,
			TraitServiceRefresh, TraitServiceDeprecate, TraitServiceDelete, TraitServiceOptions,
			TraitServiceScale, TraitServiceEnvVars, TraitServiceSecrets, TraitServiceFiles,
			TraitServiceCerts, TraitServiceLogs, TraitServiceInstances, TraitServiceDeployHistory,
			TraitManageGrants, TraitServiceSuspend, TraitServiceRollback,
		},
		RoleServiceUser: []Trait{
			TraitServiceDeploy, TraitServiceRefresh, TraitServiceLogs,
			TraitServiceInstances, TraitServiceDeployHistory,
		},
		RoleServiceAuditor: []Trait{
			TraitServiceInstancesRead, TraitServiceLogsRead, TraitServiceOptionsRead, TraitServiceRead,
		},
		RoleProjectOwner: []Trait{
			TraitProjectRun,
			TraitProjectOptions,
			TraitProjectSecrets,
			TraitProjectDelete,
			TraitProjectLogs,
			TraitProjectRollback,
			TraitServiceDeployHistory,
			TraitManageGrants,
		},
		RoleProjectUser: []Trait{TraitProjectRun, TraitProjectLogs, TraitServiceDeployHistory},
		RoleDatabaseOwner: []Trait{
			TraitDatabaseDelete,
			TraitDatabaseLaunch,
			TraitDatabaseLogs,
			TraitDatabaseCreate,
			TraitDatabaseDeprecate,
			TraitDatabaseInfo,
			TraitManageGrants,
			TraitDatabaseOptions,
		},
		RoleNamespaceOwner: []Trait{TraitNamespaceView, TraitNamespaceManage, TraitNamespaceDelete},
		RoleNamespaceUser:  []Trait{TraitNamespaceView},
	}
)

// CheckRole checks if a role exists.
func CheckRole(role Role) error {
	if _, hasRole := RoleTraits[role]; !hasRole {
		return exception.New(fmt.Sprintf("invalid_role")).WithMessagef("role: %s", role)
	}
	return nil
}

// Trait is something that we secure.
type Trait string

// Traits
const (
	// TraitSuperUser allows the user to do everything.
	TraitSuperUser Trait = "super_user"
	// TraitClusterAdmin allows the user to administer a cluster
	TraitClusterAdmin Trait = "cluster_admin"

	// TraitManageGrants allows the user to manage grants for a scope or target.
	TraitManageGrants Trait = "manage_grants"

	// TraitTeamCreate allows the user to create.
	TraitTeamCreate Trait = "team_create"
	// TraitTeamManageUsers allows the user to add and remove users from a team.
	TraitTeamManageUsers Trait = "team_manage_users"
	// TraitTeamViewTokens is the trait required to view team api tokens
	TraitTeamViewTokens Trait = "team_view_tokens"
	// TraitManageTokens is the trait required to manage team api tokens
	TraitTeamManageTokens Trait = "team_manage_tokens"
	// TraitTeamDelete allows the user to delete a team and all it's grants.
	TraitTeamDelete Trait = "team_delete"

	// TraitServiceRead allows a user to view a service
	TraitServiceRead Trait = "service_read"
	// TraitServiceCreate allows the user to create new services.
	TraitServiceCreate Trait = "service_create"
	// TraitServiceDeploy allows a user to run a deploy.
	TraitServiceDeploy Trait = "service_deploy"
	// TraitServiceRefresh allows a user to run a refresh deploy.
	TraitServiceRefresh Trait = "service_refresh"
	// TraitServiceDeprecate allows a user to run a deprecate deploy.
	TraitServiceDeprecate Trait = "service_deprecate"
	// TraitServiceDelete allows a user to run a delete deploy.
	TraitServiceDelete Trait = "service_delete"
	// TraitServiceSuspend allows a user to suspend and resume a scheduled task.
	TraitServiceSuspend Trait = "service_suspend"
	// TraitServiceRollback allows a user to rollback to a previous deploy.
	TraitServiceRollback Trait = "service_rollback"

	// TraitServiceOptions allows a user to configure service options.
	TraitServiceOptions Trait = "service_options"
	// TraitServiceOptions allows a user to read service options.
	TraitServiceOptionsRead Trait = "service_options_read"
	// TraitServiceScale allows a user to configure the scale of a service.
	TraitServiceScale Trait = "service_scale"
	// TraitServiceEnvVars allows a user to configure service environment variables.
	TraitServiceEnvVars Trait = "service_env_vars"
	// TraitServiceSecrets allows a user to configure service secrets in vault.
	TraitServiceSecrets Trait = "service_secrets"
	// TraitServiceFiles allows a user to configure service files.
	TraitServiceFiles Trait = "service_files"
	// TraitServiceCerts allows a user to configure service certs.
	TraitServiceCerts Trait = "service_certs"

	// TraitServiceLogs allows a user to manage logs.
	TraitServiceLogs Trait = "service_logs"
	// TraitServiceLogsRead allows a user to read logs.
	TraitServiceLogsRead Trait = "service_logs_read"
	// TraitServiceDeployHistory allows a user to view deploy history.
	TraitServiceDeployHistory Trait = "service_deploy_history"
	// TraitServiceInstances allows a user to manage service instances
	TraitServiceInstances Trait = "service_instances"
	// TraitServiceInstancesRead allows a user to read service info
	TraitServiceInstancesRead Trait = "service_instances_read"

	// TraitProjectRun allows a user to run a project
	TraitProjectRun Trait = "project_run"
	// TraitProjectOptions allows a user to configure a project
	TraitProjectOptions Trait = "project_options"
	// TraitProjectDelete allows a user to delete a project
	TraitProjectDelete Trait = "project_delete"
	// TraitProjectLogs allows a user to view logs for a project
	TraitProjectLogs Trait = "project_logs"
	// TraitProjectRollback allows a user to rollback a project
	TraitProjectRollback Trait = "project_rollback"
	// TraitProjectSecrets allows a user to view and edit project secrets
	TraitProjectSecrets Trait = "project_secrets"
	// TraitProjectEnvVars allows a user to configure project environment variables.
	TraitProjectEnvVars Trait = "project_env_vars"
	// TraitProjectFiles allows a user to configure project files.
	TraitProjectFiles Trait = "project_files"

	// TraitDatabaseLaunch allows a user to run a database
	TraitDatabaseLaunch Trait = "database_launch"
	// TraitDatabaseCreate allows a user to configure a database
	TraitDatabaseCreate Trait = "database_create"
	// TraitDatabaseDeprecate allows a user to deprecate a database
	TraitDatabaseDeprecate Trait = "database_deprecate"
	// TraitDatabaseDelete allows a user to delete a database
	TraitDatabaseDelete Trait = "database_delete"
	// TraitDatabaseLogs allows a user to view logs for a database
	TraitDatabaseLogs Trait = "database_logs"
	// TraitDatabaseInfo allows a user to view database information
	TraitDatabaseInfo Trait = "database_info"
	// TraitDatabaseOptions allows a user to configure database options
	TraitDatabaseOptions Trait = "database_options"

	// TraitNamespaceCreate allows a user to create a namespace
	TraitNamespaceCreate Trait = "namespace_create"
	// TraitNamespaceView allows a user to view a namespace and corresponding resources
	TraitNamespaceView Trait = "namespace_view"
	// TraitNamespaceManage allows a user to manage a namespace with full access to resources
	TraitNamespaceManage Trait = "namespace_manage"
	// TraitNamespaceDelete allows a user to delete a namespace
	TraitNamespaceDelete Trait = "namespace_delete"
)

var (
	// Traits is a list of all traits.
	Traits = []Trait{
		TraitSuperUser,
		TraitTeamCreate,
		TraitTeamManageUsers,
		TraitTeamDelete,
		TraitServiceCreate,
		TraitServiceRead,
		TraitServiceDeploy,
		TraitServiceRefresh,
		TraitServiceDeprecate,
		TraitServiceDelete,
		TraitServiceSuspend,
		TraitServiceRollback,
		TraitServiceOptions,
		TraitServiceScale,
		TraitServiceEnvVars,
		TraitServiceSecrets,
		TraitServiceFiles,
		TraitServiceCerts,
		TraitServiceOptionsRead,
		TraitServiceLogs,
		TraitServiceLogsRead,
		TraitServiceInstances,
		TraitServiceInstancesRead,
		TraitManageGrants,
		TraitServiceDeployHistory,
		TraitProjectRun,
		TraitProjectOptions,
		TraitProjectDelete,
		TraitProjectLogs,
		TraitProjectRollback,
		TraitDatabaseLaunch,
		TraitDatabaseCreate,
		TraitDatabaseDelete,
		TraitDatabaseLogs,
		TraitDatabaseDeprecate,
		TraitDatabaseInfo,
		TraitDatabaseOptions,
	}

	// ServiceTraits are traits that allow access to a service in any part.
	ServiceTraits = []Trait{
		TraitServiceCreate,
		TraitServiceRead,
		TraitServiceDeploy,
		TraitServiceRefresh,
		TraitServiceDeprecate,
		TraitServiceDelete,
		TraitServiceSuspend,
		TraitServiceRollback,
		TraitServiceOptions,
		TraitServiceScale,
		TraitServiceEnvVars,
		TraitServiceSecrets,
		TraitServiceFiles,
		TraitServiceCerts,
		TraitServiceLogs,
		TraitServiceLogsRead,
		TraitServiceInstances,
		TraitServiceInstancesRead,
		TraitServiceOptionsRead,
	}

	// DatabaseTraits are traits that allow access to a service in any part.
	DatabaseTraits = []Trait{
		TraitDatabaseLaunch,
		TraitDatabaseDelete,
		TraitDatabaseLogs,
		TraitDatabaseDeprecate,
		TraitDatabaseCreate,
		TraitDatabaseInfo,
		TraitDatabaseOptions,
	}
)

// ScopeType is the type of the scope
type ScopeType string

// TargetType is the type of the target
type TargetType string

const (
	// ScopeTypeService is the service scope
	ScopeTypeService = "service"
	// ScopeTypeTask is the UNUSED task scope reserved for future use
	ScopeTypeTask = "task" // not actually used atm
	// ScopeTypeProject is the project scope
	ScopeTypeProject = "project"
	// ScopeTypeDatabase is the database scope
	ScopeTypeDatabase = "database"
	// ScopeTypeNamespace is the namespace scope
	ScopeTypeNamespace = "namespace"
	// ScopeTypeTeam is the team scope
	ScopeTypeTeam = TargetTypeTeam
	// ScopeTypeGlobal is the global scope
	ScopeTypeGlobal = GlobalScope
	// ScopeTypeNone is the none scope
	ScopeTypeNone = ""

	// TargetTypeUser is the user target
	TargetTypeUser = "user"
	// TargetTypeTeam is the team target
	TargetTypeTeam = "team"
	// TargetTypeNone is the none scope
	TargetTypeNone = ""
)

var (
	// ScopeRoles is the mapping of scopes to allowed roles
	ScopeRoles = map[ScopeType][]Role{
		ScopeTypeService:   ServiceRoles,
		ScopeTypeTask:      []Role{},
		ScopeTypeProject:   ProjectRoles,
		ScopeTypeDatabase:  DatabaseRoles,
		ScopeTypeNamespace: NamespaceRoles,
		ScopeTypeTeam:      TeamRoles,
		ScopeTypeGlobal:    GlobalRoles,
	}
)

// TypeOfScope returns the type of the scope
func TypeOfScope(scope string) ScopeType {
	if IsServiceScope(scope) {
		return ScopeTypeService
	} else if IsTaskScope(scope) {
		return ScopeTypeTask
	} else if IsProjectScope(scope) {
		return ScopeTypeProject
	} else if IsDatabaseScope(scope) {
		return ScopeTypeDatabase
	} else if IsNamespaceScope(scope) {
		return ScopeTypeNamespace
	} else if IsTeamScope(scope) {
		return ScopeTypeTeam
	} else if scope == GlobalScope {
		return ScopeTypeGlobal
	}
	return ScopeTypeNone
}

// TypeOfTarget returns the type of the target
func TypeOfTarget(target string) TargetType {
	if IsUserTarget(target) {
		return TargetTypeUser
	} else if IsTeamTarget(target) {
		return TargetTypeTeam
	}
	return TargetTypeNone
}

// ContainsRole returns if the slice contains the role
func ContainsRole(rs []Role, r Role) bool {
	for _, rr := range rs {
		if rr == r {
			return true
		}
	}
	return false
}

// RequiredTraitForBuildMode returns the required traits for a given build mode.
func RequiredTraitForBuildMode(buildMode BuildMode) Trait {
	switch buildMode {
	case BuildModeCreate:
		return TraitServiceCreate
	case BuildModeDeploy,
		BuildModeRunTask,
		BuildModeDeploySkipBuild,
		BuildModeRunTaskSkipBuild,
		BuildModeScheduleTask:
		return TraitServiceDeploy
	case BuildModeRefresh:
		return TraitServiceRefresh
	case BuildModeDown:
		return TraitServiceDeprecate
	case BuildModeDelete:
		return TraitServiceDelete
	case BuildModeCreateSecrets:
		return TraitServiceEnvVars
	case BuildModeDeleteSecrets:
		return TraitServiceEnvVars
	case BuildModeRunProject:
		return TraitProjectRun
	case BuildModeLaunchDatabase:
		return TraitDatabaseLaunch
	case BuildModeDeprecateDatabase:
		return TraitDatabaseDeprecate
	}
	return TraitSuperUser
}
