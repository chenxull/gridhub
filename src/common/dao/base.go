package dao

import (
	"errors"
	"fmt"
	"github.com/chenxull/goGridhub/gridhub/src/common/config/models"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/logger"
	"strconv"
)

const (
	// NonExistUserID : if a user does not exist, the ID of the user will be 0.
	NonExistUserID = 0
	// ClairDBAlias ...
	ClairDBAlias = "clair-db"
)

var ErrDupRows = errors.New("sql: duplicate row in DB")

// Database is an interface of different databases
type Database interface {
	// Name returns the name of database
	Name() string
	// String returns the details of database
	String() string
	// Register registers the database which will be used
	Register(alias ...string) error
	// UpgradeSchema upgrades the DB schema to the latest version
	UpgradeSchema() error
}

// InitDatabase registers the database
func InitDatabase(database *models.Database) error {
	db, err := getDatabase(database)
	if err != nil {
		return err
	}

	logger.Infof("Registering database: %s", db.String())
	if err := db.Register(); err != nil {
		return err
	}

	logger.Info("Register database completed")
	return nil
}
func getDatabase(database *models.Database) (db Database, err error) {

	switch database.Type {
	case "", "postgresql":
		db = NewPGSQL(
			database.PostGreSQL.Host,
			strconv.Itoa(database.PostGreSQL.Port),
			database.PostGreSQL.Username,
			database.PostGreSQL.Password,
			database.PostGreSQL.Database,
			database.PostGreSQL.SSLMode,
			database.PostGreSQL.MaxIdleConns,
			database.PostGreSQL.MaxOpenConns,
		)
	default:
		err = fmt.Errorf("invalid database: %s", database.Type)
	}
	return
}
