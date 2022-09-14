package postgres

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/stdlib"
	"github.com/spf13/viper"
	"github.com/wego/pkg/common"
	sqlTrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
	gormTrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gorm.io/gorm.v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type config struct {
	Host                   string
	Port                   int
	Username               string
	Password               string
	Database               string
	MaxOpenConns           int `mapstructure:"max_open_conns"`
	MaxIdleConns           int `mapstructure:"max_idle_conns"`
	ConnMaxLifeTimeMinutes int `mapstructure:"conn_max_life_time_minutes"`
}

const (
	// ConfigEnvName is the name of the env that contains the config value
	ConfigEnvName = "DB_CONFIG_SECRET"
)

// NewConnection create new db instance from env and fallback to config file if env is empty
func NewConnection(dbConfigEnvName, dbConfigFilePath, dbConfigFormat string) (*gorm.DB, error) {
	if len(os.Getenv(dbConfigEnvName)) > 0 {
		return NewConnectionFromEnv(dbConfigEnvName, dbConfigFormat)
	}

	return NewConnectionFromFile(dbConfigFilePath, dbConfigFormat)
}

// NewConnectionFromFile create new db instance from config file
func NewConnectionFromFile(dbConfigFilePath, dbConfigFormat string) (*gorm.DB, error) {
	config, err := readConfigFromFile(dbConfigFilePath, dbConfigFormat)
	if err != nil {
		return nil, fmt.Errorf("cannot load config for DB: %w", err)
	}

	return connectDB(config)
}

// NewConnectionFromEnv create new db instance from env
func NewConnectionFromEnv(envName, configType string) (*gorm.DB, error) {
	config, err := readConfigFromEnv(envName, configType)
	if err != nil {
		return nil, fmt.Errorf("cannot load config for DB: %w", err)
	}

	return connectDB(config)
}

func connectDB(c *config) (*gorm.DB, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		url.QueryEscape(c.Username), url.QueryEscape(c.Password), c.Host, c.Port, c.Database)
	sqlTrace.Register("pgx", &stdlib.Driver{}, sqlTrace.WithServiceName(viper.GetString("service_name")))
	sqlDB, err := sqlTrace.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(c.ConnMaxLifeTimeMinutes) * time.Minute)

	db, err := gormTrace.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger:  logger.Default.LogMode(logger.Silent),
		NowFunc: func() time.Time { return common.CurrentUTCTime() },
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func readConfigFromFile(dbConfigFilePath, dbConfigFormat string) (*config, error) {
	configReader := viper.New()
	configReader.SetConfigFile(dbConfigFilePath)
	configReader.SetConfigType(dbConfigFormat)

	if err := configReader.ReadInConfig(); err != nil {
		return nil, err
	}

	return unmarshalConfig(configReader)
}

func readConfigFromEnv(envName, configType string) (*config, error) {
	configReader := viper.New()
	configReader.SetConfigType(configType)

	if err := configReader.ReadConfig(strings.NewReader(os.Getenv(envName))); err != nil {
		return nil, err
	}

	return unmarshalConfig(configReader)
}

func unmarshalConfig(configReader *viper.Viper) (*config, error) {
	var c config
	env := viper.GetString("env")

	envConfig := configReader.Sub(env)
	if envConfig == nil {
		return nil, fmt.Errorf("env[%s] not found in config", env)
	}

	if err := envConfig.Unmarshal(&c); err != nil {
		return nil, err
	}

	return &c, nil
}
