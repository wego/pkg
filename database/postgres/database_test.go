package postgres

import (
	"os"
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

func Test_readConfigFromEnv(t *testing.T) {
	expectedConfig := &config{
		Host:                   "127.0.0.1",
		Port:                   15432,
		Database:               "payments-test",
		Username:               "root",
		Password:               "password",
		MaxOpenConns:           10,
		MaxIdleConns:           5,
		ConnMaxLifeTimeMinutes: 60,
	}
	type args struct {
		envValue   string
		configType string
	}
	tests := []struct {
		name    string
		args    args
		want    *config
		wantErr bool
	}{
		{
			name: "readConfigFromEnv_toml_Invalid_Error",
			args: args{
				envValue: `
[staging]
invalid = [toml]
`,
				configType: "toml",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "readConfigFromEnv_toml_WrongEnv_Error",
			args: args{
				envValue: `
[production]
host = "127.0.0.1"
port = 15432
database = "payments-test"
username = "root"
password = "password"
max_open_conns = 10
max_idle_conns = 5
conn_max_life_time_minutes = 60
`,
				configType: "toml",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "readConfigFromEnv_toml_OK",
			args: args{
				envValue: `
[test]
host = "127.0.0.1"
port = 15432
database = "payments-test"
username = "root"
password = "password"
max_open_conns = 10
max_idle_conns = 5
conn_max_life_time_minutes = 60
`,
				configType: "toml",
			},
			want:    expectedConfig,
			wantErr: false,
		},
		{
			name: "readConfigFromEnv_yaml_Invalid_Error",
			args: args{
				envValue: `
[staging]
invalid = format
`,
				configType: "yaml",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "readConfigFromEnv_yaml_OK",
			args: args{
				envValue: `
test:
  host: 127.0.0.1
  port: 15432
  database: payments-test
  username: root
  password: password
  max_open_conns: 10
  max_idle_conns: 5
  conn_max_life_time_minutes: 60
`,
				configType: "yaml",
			},
			want:    expectedConfig,
			wantErr: false,
		},
	}

	testEnvName := "TEST_ENV_NAME"
	viper.Set("env", "test")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(testEnvName, tt.args.envValue)
			got, err := readConfigFromEnv(testEnvName, tt.args.configType)
			if (err != nil) != tt.wantErr {
				t.Errorf("readConfigFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readConfigFromEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
