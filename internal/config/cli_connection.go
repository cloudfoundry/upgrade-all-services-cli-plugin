package config

// CLIConnection is the interface that the CF Plugin infrastructure must
// implement in order to work with this plugin.
//
//go:generate go tool counterfeiter -generate
//counterfeiter:generate . CLIConnection
type CLIConnection interface {
	IsLoggedIn() (bool, error)
	AccessToken() (string, error)
	ApiVersion() (string, error)
	ApiEndpoint() (string, error)
	IsSSLDisabled() (bool, error)
}
