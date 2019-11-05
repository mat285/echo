package logging

import (
	"strings"
	"sync"

	"git.blendlabs.com/blend/deployinator/pkg/core"
	"git.blendlabs.com/blend/deployinator/pkg/types"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/sentry"
)

var (
	defaultLogger *logger.Logger
	defaultLock   sync.Mutex
)

// SetDefault sets the default logger
func SetDefault(log *logger.Logger) {
	defaultLock.Lock()
	defaultLogger = log
	defaultLock.Unlock()
}

// Default returns the default logger
func Default() *logger.Logger {
	if defaultLogger == nil {
		defaultLock.Lock()
		defer defaultLock.Unlock()
		if defaultLogger == nil {
			defaultLogger = NewFromEnvOrAll()
		}
	}
	return defaultLogger
}

// NewFromEnvOrAll creates a new logger from the env or if an error occurs then logger.All()
func NewFromEnvOrAll() *logger.Logger {
	return core.NewLoggerFromEnvOrAll()
}

// LogUserLogin logs a user creation
func LogUserLogin(username string) {
	Default().Trigger(logger.NewAuditEvent(username, VerbLogin))
}

// LogGrant logs the grant
func LogGrant(grant types.Grant) {
	Default().Trigger(logger.NewAuditEvent(
		grant.GrantedBy,
		VerbGrant,
	).WithNoun(string(grant.Role)).WithSubject(grant.Target).WithProperty(grant.Scope))
}

// LogRevoke logs the grant revoke
func LogRevoke(username, scope, target, role string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbRevoke,
	).WithNoun(role).WithSubject(target).WithProperty(scope))
}

// LogTeamCreate logs the creation of the team
func LogTeamCreate(team types.Team) {
	Default().Trigger(logger.NewAuditEvent(
		team.CreatedBy,
		VerbCreate,
	).WithNoun(NounTeam).WithSubject(team.String()))
}

// LogTeamChange logs a team change
func LogTeamChange(username, team string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbEdit,
	).WithNoun(NounTeam).WithSubject(team))
}

// LogUserDelete logs the deletion of the user
func LogUserDelete(username, targetUser string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbDelete,
	).WithNoun(NounUser).WithSubject(targetUser))
}

// LogServiceEvent logs the service event
func LogServiceEvent(service string, buildMode types.BuildMode, username string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		string(buildMode),
	).WithNoun(NounService).WithSubject(service))
}

// LogProjectEvent logs the service event
func LogProjectEvent(project string, buildMode types.BuildMode, username string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		string(buildMode),
	).WithNoun(NounProject).WithSubject(project))
}

// LogProjectSet logs the metadata about the project when its set to etcd
func LogProjectSet(project string, services []string, triggers []string, stacktrace string) {
	e := map[string]string{
		"services":   strings.Join(services, ","),
		"triggers":   strings.Join(triggers, ","),
		"stacktrace": stacktrace,
	}
	Default().Trigger(logger.NewAuditEvent(
		types.UserSystem,
		"",
	).WithNoun(NounProject).WithSubject(project).WithExtra(e))
}

// LogDatabaseEvent logs the service event
func LogDatabaseEvent(db string, buildMode types.BuildMode, username string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		string(buildMode),
	).WithNoun(NounDatabase).WithSubject(db))
}

// LogEnvVarCreate logs the creation of a EnvVar
func LogEnvVarCreate(username, service, newKey string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbCreate,
	).WithNoun(NounEnvVar).WithSubject(newKey).WithProperty(service))
}

// LogEnvVarEdit logs the editing of a EnvVar
func LogEnvVarEdit(username, service, oldKey, newKey string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbEdit,
	).WithSubject(newKey).WithProperty(service))
}

// LogEnvVarDelete logs the deletion of a EnvVar
func LogEnvVarDelete(username, service, oldKey string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbDelete,
	).WithNoun(NounEnvVar).WithSubject(oldKey).WithProperty(service))
}

// LogSecretCreate logs the creation of a secret
func LogSecretCreate(username, service, secret string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbCreate,
	).WithNoun(NounService).WithSubject(secret).WithProperty(service))
}

// LogSecretEdit logs the editing of a secret
func LogSecretEdit(username, service, oldSecret, newSecret string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbEdit,
	).WithNoun(NounService).WithSubject(newSecret).WithProperty(service))
}

// LogSecretView logs the viewing of a secret
func LogSecretView(username, service, secret string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbView,
	).WithNoun(NounService).WithSubject(secret).WithProperty(service))
}

// LogSecretDelete logs the deletion of a secret
func LogSecretDelete(username, service, secret string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbDelete,
	).WithNoun(NounService).WithSubject(secret).WithProperty(service))
}

// LogVaultRequest logs the vault request
func LogVaultRequest(path, verb string) {
	Default().Trigger(logger.NewAuditEvent(
		"",
		VerbGet,
	).WithNoun(NounVault).WithSubject(path).WithProperty(verb))
}

// LogFileCreate logs the creation of a file
func LogFileCreate(username, service, file string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbCreate,
	).WithNoun(NounFile).WithProperty(file).WithSubject(service))
}

// LogFileEdit logs the editing of a file
func LogFileEdit(username, service, oldFile, newFile string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbEdit,
	).WithNoun(NounFile).WithProperty(newFile).WithSubject(service).WithExtra(map[string]string{"oldFile": oldFile}))
}

// LogFileDelete logs the deletion of a file
func LogFileDelete(username, service, file string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbDelete,
	).WithNoun(NounFile).WithProperty(file).WithSubject(service))
}

// LogFileView logs the viewing of a file
func LogFileView(username, service, file string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbView,
	).WithNoun(NounFile).WithProperty(file).WithSubject(service))
}

// LogCertCreate logs the creation of a cert
func LogCertCreate(username, service, file string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbCreate,
	).WithNoun(NounCert).WithProperty(file).WithSubject(service))
}

// LogCertEdit logs the editing of a cert
func LogCertEdit(username, service, file string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbEdit,
	).WithNoun(NounCert).WithProperty(file).WithSubject(service))
}

// LogCertDelete logs the deletion of a cert
func LogCertDelete(username, service, file string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbDelete,
	).WithNoun(NounCert).WithProperty(file).WithSubject(service))
}

// LogCertView logs the viewing of a cert
func LogCertView(username, service, file string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbView,
	).WithNoun(NounCert).WithProperty(file).WithSubject(service))
}

// LogAPITokenCreate logs the creation of an api token
func LogAPITokenCreate(username, token string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbCreate,
	).WithNoun(NounAPIToken).WithSubject(token))
}

// LogAPITokenExpire logs the expiration of an api token
func LogAPITokenExpire(username, token string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbExpire,
	).WithNoun(NounAPIToken).WithSubject(token))
}

// LogAPITokenDelete logs the deletion of an api token
func LogAPITokenDelete(username, token string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbDelete,
	).WithNoun(NounAPIToken).WithSubject(token))
}

// LogAPITokenView logs the viewing of api tokens
func LogAPITokenView(username string, optionalTeam ...string) {
	team := ""
	if len(optionalTeam) > 0 {
		team = optionalTeam[0]
	}
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbView,
	).WithNoun(NounAPIToken).WithSubject(team))
}

// LogVaultTokenCreate logs the creation of an api token
func LogVaultTokenCreate(username, token string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbCreate,
	).WithNoun(NounVaultToken).WithSubject(token))
}

// LogOptionsEdit logs the editing of service options
func LogOptionsEdit(username, service string) {
	Default().Trigger(logger.NewAuditEvent(
		username,
		VerbEdit,
	).WithNoun(NounOptions).WithSubject(service))
}

// LogAwsRequest logs the aws request
func LogAwsRequest(url, verb string) {
	Default().Trigger(logger.NewAuditEvent(
		"",
		VerbGet,
	).WithNoun(NounAWS).WithSubject(url).WithProperty(verb))
}

// LogAddSentryListener adds a listener to the client for sentry logging
//
// This is a convenience method that adds a listener to the logger and also
// sets sentry to only listen for fatal flags (since Deployinator lots a *lot*
// of errors).
func LogAddSentryListener(client *sentry.Client) {
	listener := logger.NewErrorEventListener(client.Notify)
	Default().Listen(logger.Fatal, "sentry", listener)
}
