package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/blend/go-sdk/exception"
)

const (
	// GlobalScope is used for things like superuser etc.
	GlobalScope = "global"
)

// GrantScope represents a scope for which a user can be granted.
type GrantScope struct {
	Scope     string    `json:"scope" yaml:"scope"`
	CreatedBy string    `json:"createdBy" yaml:"createdBy"`
	CreatedAt time.Time `json:"createdAt" yaml:"createdAt"`
}

// Grant represents a set of roles granted to a scope for a user.
type Grant struct {
	Scope     string    `json:"scope" yaml:"scope"`
	Target    string    `json:"target" yaml:"target"`
	GrantedBy string    `json:"grantedBy" yaml:"grantedBy"`
	GrantedAt time.Time `json:"grantedAt" yaml:"grantedAt"`
	Role      Role      `json:"roles" yaml:"roles"`
}

// Grants is a map between scope and role.
type Grants map[string][]Role

// Validate validates that the grant scope, target, role are all valid types
func (g Grant) Validate() error {
	scopeType, targetType := TypeOfScope(g.Scope), TypeOfTarget(g.Target)
	if scopeType == ScopeTypeNone {
		return exception.New("Invalid Scope Type").WithMessagef("Scope type `%s`", scopeType)
	}
	if targetType == TargetTypeNone {
		return exception.New("Invalid Target Type").WithMessagef("Target type `%s`", targetType)
	}
	if err := CheckRole(g.Role); err != nil {
		return err
	}
	validRoles, ok := ScopeRoles[scopeType]
	if len(g.Role) == 0 || !ok || len(validRoles) == 0 || !ContainsRole(validRoles, g.Role) {
		return exception.New("Invalid Role").WithMessagef("Role `%s` is not valid for scope `%s`", g.Role, g.Scope)
	}
	return nil
}

// HasAnyTrait returns if a user has a given trait for a given scope.
func (g Grants) HasAnyTrait(scope string, traits ...Trait) bool {
	// check global scope for the trait and superuser.
	if globalRoles, hasGlobal := g[GlobalScope]; hasGlobal {
		for _, globalRole := range globalRoles {
			if g.RoleHasAnyTrait(globalRole, TraitSuperUser) {
				return true
			}
			if g.RoleHasAnyTrait(globalRole, traits...) {
				return true
			}
		}
	}

	if scopedRoles, hasScopedRoles := g[scope]; hasScopedRoles {
		for _, scopedRole := range scopedRoles {
			if g.RoleHasAnyTrait(scopedRole, traits...) {
				return true
			}
		}
	}
	return false
}

// HasTraitForSomeScope returns if a user has any of the given traits in at
// least one of the given scopes
func (g Grants) HasTraitForSomeScope(scopes []string, traits ...Trait) bool {
	for _, scope := range scopes {
		if g.HasAnyTrait(scope, traits...) {
			return true
		}
	}
	return false
}

// RoleHasAnyTrait returns if a given role has a given trait.
func (g Grants) RoleHasAnyTrait(role Role, traits ...Trait) bool {
	if roleTraits, roleExists := RoleTraits[role]; roleExists {
		for _, roleTrait := range roleTraits {
			for _, checkTrait := range traits {
				if roleTrait == checkTrait {
					return true
				}
			}
		}
	}
	return false
}

// IsSuperuser returns if the user is a superuser.
func (g Grants) IsSuperuser() bool {
	// check global scope for the trait and superuser.
	if globalRoles, hasGlobal := g[GlobalScope]; hasGlobal {
		for _, globalRole := range globalRoles {
			if g.RoleHasAnyTrait(globalRole, TraitSuperUser) {
				return true
			}
		}
	}
	return false
}

// GetValidGrantRoles returns the roles that a user has privileges to assign for
// a grant
func (g Grants) GetValidGrantRoles() []Role {
	if g.IsSuperuser() {
		return AllRoles
	}
	return append(ServiceRoles,
		append(ProjectRoles,
			append(DatabaseRoles,
				append(TeamRoles,
					NamespaceRoles...)...)...,
		)...,
	)

}

// GetAuthorizedScopes filters out a slice of scopes in which a user has
// authorization to manage grants, given a list of scopes
func (g Grants) GetAuthorizedScopes(scopes []string) map[string]bool {
	authorizedScopes := make(map[string]bool)
	for _, scope := range scopes {
		if g.HasAnyTrait(scope, TraitManageGrants, TraitSuperUser) {
			authorizedScopes[scope] = true
		}
	}
	return authorizedScopes
}

// GetAuthorizedGrantsForScopes returns the grants that a user is allowed to see/manage
// given a set of grants
func (g Grants) GetAuthorizedGrantsForScopes(authorizedGrantScopes map[string]bool, allGrants []Grant) []Grant {
	var authorizedGrants []Grant

	for _, grant := range allGrants {
		if _, ok := authorizedGrantScopes[grant.Scope]; ok {
			authorizedGrants = append(authorizedGrants, grant)
		}
	}
	return authorizedGrants
}

// HasTrait returns whether a set of grants has a trait
func (g Grants) HasTrait(traits ...Trait) bool {
	for scope := range g {
		if g.HasAnyTrait(scope, traits...) {
			return true
		}
	}
	return false
}

// String returns a string representation of the grants collection.
func (g Grants) String() string {
	buffer := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent("", "\t")
	_ = encoder.Encode(g)
	return buffer.String()
}

func (g Grant) String() string {
	return fmt.Sprintf("Scope: `%s`, Target: `%s`, Role: `%s`, GrantedBy: `%s`, GrantedAt: `%s`", g.Scope, g.Target, g.Role, g.GrantedBy, g.GrantedAt.UTC())
}

// NamespaceScope returns a fully qualified scope for a namespace.
func NamespaceScope(namespaceName string) string {
	return fmt.Sprintf("namespace:%s", namespaceName)
}

// IsNamespaceScope returns if the scope is a namespace scope.
func IsNamespaceScope(scope string) bool {
	return strings.HasPrefix(scope, "namespace:")
}

// GetNamespaceNameFromScope returns the namespace name from a given scope.
func GetNamespaceNameFromScope(scope string) (string, error) {
	if !IsNamespaceScope(scope) {
		return "", exception.New("scope is not for a namespace")
	}
	return strings.TrimPrefix(scope, "namespace:"), nil
}

// TeamScope returns a fully qualified scope for a team.
func TeamScope(teamName string) string {
	return fmt.Sprintf("team:%s", teamName)
}

// IsTeamScope returns if the scope is a team scope.
func IsTeamScope(scope string) bool {
	return strings.HasPrefix(scope, "team:")
}

// GetTeamNameFromScope returns the team name from a given scope.
func GetTeamNameFromScope(scope string) (string, error) {
	if !IsTeamScope(scope) {
		return "", exception.New("scope is not for a team")
	}
	return strings.TrimPrefix(scope, "team:"), nil
}

// ServiceScope returns a fully qualified scope for a service.
func ServiceScope(serviceName string) string {
	return fmt.Sprintf("service:%s", serviceName)
}

// TaskScope returns the scope of a task for display purposes. For data storage use ServiceScope
func TaskScope(serviceName string) string {
	return fmt.Sprintf("task:%s", serviceName)
}

// IsServiceScope returns if the scope is a service scope.
func IsServiceScope(scope string) bool {
	return strings.HasPrefix(scope, "service:")
}

// IsTaskScope returns if the scope is a task (service) scope
func IsTaskScope(scope string) bool {
	return strings.HasPrefix(scope, "task:")
}

// GetServiceNameFromScope returns the service name from a given scope.
func GetServiceNameFromScope(scope string) (string, error) {
	if !IsServiceScope(scope) {
		return "", exception.New("scope is not for a service")
	}
	return strings.TrimPrefix(scope, "service:"), nil
}

// GetTaskNameFromScope returns the task name from a given scope.
func GetTaskNameFromScope(scope string) (string, error) {
	if !IsTaskScope(scope) {
		return "", exception.New("scope is not for a task")
	}
	return strings.TrimPrefix(scope, "task:"), nil
}

// ProjectScope returns a fully qualified scope for a team.
func ProjectScope(projectName string) string {
	return fmt.Sprintf("project:%s", projectName)
}

// IsProjectScope returns if the scope is a proejct scope.
func IsProjectScope(scope string) bool {
	return strings.HasPrefix(scope, "project:")
}

// GetProjectNameFromScope returns the project name from a given scope.
func GetProjectNameFromScope(scope string) (string, error) {
	if !IsProjectScope(scope) {
		return "", exception.New("scope is not for a project")
	}
	return strings.TrimPrefix(scope, "project:"), nil
}

// DatabaseScope returns a fully qualified scope for a database.
func DatabaseScope(databaseName string) string {
	return fmt.Sprintf("database:%s", databaseName)
}

// IsDatabaseScope returns if the scope is a database scope.
func IsDatabaseScope(scope string) bool {
	return strings.HasPrefix(scope, "database:")
}

// GetDatabaseNameFromScope returns the database name from a given scope.
func GetDatabaseNameFromScope(scope string) (string, error) {
	if !IsDatabaseScope(scope) {
		return "", exception.New("scope is not for a database")
	}
	return strings.TrimPrefix(scope, "database:"), nil
}

// UserTarget creates a new target for a team.
func UserTarget(username string) string {
	return fmt.Sprintf("user:%s", username)
}

// IsUserTarget returns if the target is a user target.
func IsUserTarget(target string) bool {
	return strings.HasPrefix(target, "user:")
}

// GetUsernameFromTarget gets a username from a given target.
func GetUsernameFromTarget(target string) (string, error) {
	if !IsUserTarget(target) {
		return "", exception.New("target is not a user")
	}
	return strings.TrimPrefix(target, "user:"), nil
}

// TeamTarget creates a new target for a team.
func TeamTarget(teamName string) string {
	return fmt.Sprintf("team:%s", teamName)
}

// IsTeamTarget returns if the target is a team target.
func IsTeamTarget(target string) bool {
	return strings.HasPrefix(target, "team:")
}

// GetTeamNameFromTarget gets a team name from a given target.
func GetTeamNameFromTarget(target string) (string, error) {
	if !IsTeamTarget(target) {
		return "", exception.New("target is not a team")
	}
	return strings.TrimPrefix(target, "team:"), nil
}

// KubeUserTarget is the escaped target for a user
func KubeUserTarget(username string) string {
	return "user_" + username
}

// IsKubeUserTarget checks if the target is a kube user target
func IsKubeUserTarget(target string) bool {
	return strings.HasPrefix(target, "user_")
}

// GetUsernameFromKubeTarget returns the username from the target
func GetUsernameFromKubeTarget(target string) (string, error) {
	if !IsKubeUserTarget(target) {
		return "", exception.New("Target is not a user")
	}
	return strings.TrimPrefix(target, "user_"), nil
}

// KubeTeamTarget returns the escaped team target
func KubeTeamTarget(team string) string {
	return "team_" + team
}

// IsKubeTeamTarget returns if this is an escaped team target
func IsKubeTeamTarget(target string) bool {
	return strings.HasPrefix(target, "team_")
}

// GetTeamNameFromKubeTarget returns the team name from the target
func GetTeamNameFromKubeTarget(target string) (string, error) {
	if !IsKubeTeamTarget(target) {
		return "", exception.New("Target is not a team")
	}
	return strings.TrimPrefix(target, "team_"), nil
}

// KubeEscapeTarget fixes the target string to be valid for kube use
func KubeEscapeTarget(target string) (string, error) {
	if target == "" {
		return "", nil
	} else if IsUserTarget(target) {
		username, err := GetUsernameFromTarget(target)
		if err != nil {
			return "", err
		}
		return KubeUserTarget(username), nil
	} else if IsTeamTarget(target) {
		team, err := GetTeamNameFromTarget(target)
		if err != nil {
			return "", err
		}
		return KubeTeamTarget(team), nil
	}
	return "", exception.New("InvalidTarget").WithMessagef("Target: `%s`", target)
}

// KubeUnescapeTarget returns the escaped target to its original form
func KubeUnescapeTarget(target string) (string, error) {
	if target == "" {
		return "", nil
	} else if IsKubeUserTarget(target) {
		username, err := GetUsernameFromKubeTarget(target)
		if err != nil {
			return "", err
		}
		return UserTarget(username), nil
	} else if IsKubeTeamTarget(target) {
		team, err := GetTeamNameFromKubeTarget(target)
		if err != nil {
			return "", err
		}
		return TeamTarget(team), nil
	}
	return "", exception.New("InvalidTarget").WithMessagef("Target: `%s`", target)
}
