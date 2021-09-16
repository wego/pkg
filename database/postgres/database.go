package postgres

import (
	"fmt"
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

// NewConnection : create new db instance
func NewConnection(dbConfigFilePath string) (*gorm.DB, error) {
	config, err := readConfig(dbConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf("cannot load config for DB: %w", err)
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.Username, config.Password, config.Host, config.Port, config.Database)
	sqlTrace.Register("pgx", &stdlib.Driver{}, sqlTrace.WithServiceName(viper.GetString("service_name")))
	sqlDB, err := sqlTrace.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.ConnMaxLifeTimeMinutes) * time.Minute)

	db, err := gormTrace.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger:  logger.Default.LogMode(logger.Silent),
		NowFunc: func() time.Time { return common.CurrentUTCTime() },
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func readConfig(dbConfigFilePath string) (*config, error) {
	configReader := viper.New()
	configReader.SetConfigFile(dbConfigFilePath)

	if err := configReader.ReadInConfig(); err != nil {
		return nil, err
	}

	env := viper.GetString("env")
	var c config

	if err := configReader.Sub(env).Unmarshal(&c); err != nil {
		return nil, err
	}

	return &c, nil
}
