package sql

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/wego/pkg/errors"
	"gorm.io/gorm/clause"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"gorm.io/gorm"

	// Needed for migration
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"
	"github.com/spf13/viper"

	database "github.com/wego/pkg/database/postgres"
	intTest "github.com/wego/pkg/integration-test"
)

// TestSuite ...
type TestSuite struct {
	dbConn                  *gorm.DB
	dbName                  string
	dbMigrationSourceFolder string
	dbDataSeedFilePath      string
}

// TestSuiteParams ...
type TestSuiteParams struct {
	DbConfigFilePath        string
	DbName                  string
	DbMigrationSourceFolder string
	DbDataSeedFilePath      string
}

// NewTestSuite return a new TestSuite for SQL
func NewTestSuite(p TestSuiteParams) intTest.TestSuite {
	viper.SetDefault(env, "test")
	viper.BindEnv(env, "APP_ENV")

	dbConfigFilePath := p.DbConfigFilePath
	if len(dbConfigFilePath) == 0 {
		dbConfigFilePath = defaultDbConfigFilePath
	}

	dbName := p.DbName
	if len(dbName) == 0 {
		dbName = defaultDbName
	}

	dbMigrationSourceFolder := p.DbMigrationSourceFolder
	if len(dbMigrationSourceFolder) == 0 {
		dbMigrationSourceFolder = defaultDbMigrationSourceFolder
	}

	dbDataSeedFilePath := p.DbDataSeedFilePath
	if len(dbDataSeedFilePath) == 0 {
		dbDataSeedFilePath = defaultDbDataSeedFilePath
	}

	db, err := database.NewConnection(resolveExternalPath(dbConfigFilePath))
	if err != nil {
		log.Fatal(err)
	}

	return &TestSuite{
		dbConn:                  db,
		dbName:                  dbName,
		dbMigrationSourceFolder: dbMigrationSourceFolder,
		dbDataSeedFilePath:      dbDataSeedFilePath,
	}
}

// StartUp runs database migration up scripts
func (s *TestSuite) StartUp() interface{} {
	sqlDB, err := s.dbConn.DB()
	if err != nil {
		log.Fatal(err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	migrationPath := "file://" + resolveExternalPath(s.dbMigrationSourceFolder)
	m, err := migrate.NewWithDatabaseInstance(migrationPath, s.dbName, driver)
	if err != nil {
		log.Fatal(err)
	}

	err = m.Up()
	handleMigrateErr(err)

	return s.dbConn
}

func resolveExternalPath(externalPath string) string {
	// base directory of the running test file
	testDir, _ := os.Getwd()
	// index 1 can be hardcoded since 'pkg' folder is constant in the project
	testDirectoryCount := strings.Count(strings.SplitAfter(testDir, "pkg")[1], string(filepath.Separator))

	slashes := ""
	for i := 0; i <= testDirectoryCount; i++ {
		// filepath.Separator is not used because got error, it seems '/' also works in Windows
		slashes = slashes + "../"
	}

	return slashes + externalPath
}

func handleMigrateErr(err error) {
	// https://github.com/golang-migrate/migrate/blob/v4.11.0/internal/cli/commands.go#L169-L174
	if err != nil {
		if err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		log.Println(err)
	}
}

// CleanUp delete all records and insert data from seed sql file
func (s *TestSuite) CleanUp() {
	tables, err := s.getAllTableNames()
	if err != nil {
		log.Fatal(err)
	}

	// delete all records
	err = s.dbConn.Transaction(func(tx *gorm.DB) error {
		for _, table := range tables {
			query := `TRUNCATE ` + pq.QuoteIdentifier(table) + ` RESTART IDENTITY CASCADE`
			fmt.Println(query)
			if err := tx.Exec(query).Error; err != nil {
				log.Fatal(err)
			}
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	// insert data seed test data for next test suite
	seedSQLPath := resolveExternalPath(s.dbDataSeedFilePath)
	seedSQL, err := ioutil.ReadFile(filepath.Clean(seedSQLPath))
	if err != nil {
		log.Fatal(err)
	}
	if err = s.dbConn.Exec(string(seedSQL)).Error; err != nil {
		log.Fatal(err)
	}
}

// Create creates new record(s)
func (s *TestSuite) Create(ctx context.Context, value interface{}) error {
	const op errors.Op = "sql.TestSuite.Create"
	reflectValue := reflect.ValueOf(value)
	switch reflectValue.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < reflectValue.Len(); i++ {
			if err := s.Create(ctx, reflectValue.Index(i).Interface()); err != nil {
				return err
			}
		}
	default:
		return s.create(ctx, op, value)
	}
	return nil
}

func (s *TestSuite) create(ctx context.Context, op errors.Op, v interface{}) error {
	if err := s.dbConn.WithContext(ctx).
		Clauses(clause.OnConflict{
			UpdateAll: true,
		}).
		Create(v).
		Error; err != nil {
		return errors.WrapGORMError(op, err)
	}
	return nil
}

// FindByID retrieves the entity from database by primary key ID
func (s *TestSuite) FindByID(ctx context.Context, entity interface{}, id uint) error {
	const op errors.Op = "sql.TestSuite.FindByID"
	if err := s.dbConn.WithContext(ctx).
		Preload(clause.Associations).
		First(entity, id).
		Error; err != nil {
		return errors.WrapGORMError(op, err)
	}
	return nil
}

func (s *TestSuite) getAllTableNames() ([]string, error) {
	// select all tables in current schema, exclude schema_migrations
	query := `SELECT table_name FROM information_schema.tables
							WHERE table_schema=(SELECT current_schema()) AND
										table_type='BASE TABLE' AND
										table_name<>'schema_migrations'`
	var tables []string
	err := s.dbConn.Raw(query).Scan(&tables).Error
	if err != nil {
		return nil, err
	}

	return tables, nil
}
