// Package env consists of environment variables
package env

import "github.com/spf13/viper"

// Constants for environment values
const (
	Production = "production"
	Staging    = "staging"
)

const (
	appConfig       = "app_config"
	appConfigSecret = "app_config_secret"
	dbConfig        = "db_config"
	env             = "env"
	version         = "version"
	serviceName     = "service_name"
	authModel       = "auth_model"
)

// AuthModelFilePath returns the path to the auth model file
func AuthModelFilePath() string {
	return viper.GetString(authModel)
}

// ConfigFilePath returns the path to the config file
func ConfigFilePath() string {
	return viper.GetString(appConfig)
}

// ConfigValue returns the value of the config file
func ConfigValue() string {
	return viper.GetString(appConfigSecret)
}

// DBConfigFilePath returns the path to the db config file
func DBConfigFilePath() string {
	return viper.GetString(dbConfig)
}

// Env returns the environment
func Env() string {
	return viper.GetString(env)
}

// IsProduction returns true if the environment is production
func IsProduction() bool {
	return Env() == Production
}

// IsStaging returns true if the environment is staging
func IsStaging() bool {
	return Env() == Staging
}

// ServiceName returns the service name
func ServiceName() string {
	return viper.GetString(serviceName)
}

// Version returns the version
func Version() string {
	return viper.GetString(version)
}

func init() {
	viper.SetDefault(env, "development")
	viper.SetDefault(serviceName, "unknown")
	viper.SetDefault(version, "unknown")
	viper.SetDefault(appConfig, "./config/config.toml")
	viper.SetDefault(dbConfig, "./config/database.yml")
	viper.SetDefault(authModel, "./config/auth_model.conf")

	viper.BindEnv(env, "APP_ENV")
	viper.BindEnv(serviceName, "APP_NAME")
	viper.BindEnv(version, "APP_VERSION")
	viper.BindEnv(appConfig, "APP_CONFIG")
	viper.BindEnv(appConfigSecret, "APP_CONFIG_SECRET")
	viper.BindEnv(dbConfig, "DB_CONFIG")
	viper.BindEnv(authModel, "AUTH_MODEL")
}
