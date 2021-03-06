package dao

import (
	"fmt"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/logger"
	"github.com/golang-migrate/migrate"
	_ "github.com/lib/pq" // register pgsql driver

	"net/url"
	"os"
)

const defaultMigrationPath = "migrations/postgresql/"

type pgsql struct {
	host         string
	port         string
	usr          string
	pwd          string
	database     string
	sslmode      string
	maxIdleConns int
	maxOpenConns int
}

// Name returns the name of PostgreSQL
func (p *pgsql) Name() string {
	return "PostgreSQL"
}

// String ...
func (p *pgsql) String() string {
	return fmt.Sprintf("type-%s host-%s port-%s databse-%s sslmode-%q",
		p.Name(), p.host, p.port, p.database, p.sslmode)
}

func (p pgsql) Register(alias ...string) error {
	//todo
	panic("implement me")
}

// UpgradeSchema calls migrate tool to upgrade schema to the latest based on the SQL scripts.
func (p *pgsql) UpgradeSchema() error {
	dbURL := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(p.usr, p.pwd),
		Host:     fmt.Sprintf("%s:%s", p.host, p.port),
		Path:     p.database,
		RawQuery: fmt.Sprintf("sslmode=%s", p.sslmode),
	}

	// For UT
	path := os.Getenv("POSTGRES_MIGRATION_SCRIPTS_PATH")
	if len(path) == 0 {
		path = defaultMigrationPath
	}
	srcURL := fmt.Sprintf("file://%s", path)
	m, err := migrate.New(srcURL, dbURL.String())
	if err != nil {
		return err
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil || dbErr != nil {
			logger.Warningf("Failed to close migrator, source error: %v, db error: %v", srcErr, dbErr)
		}
	}()
	logger.Infof("Upgrading schema for pgsql ...")
	err = m.Up()
	if err == migrate.ErrNoChange {
		logger.Infof("No change in schema, skip.")
	} else if err != nil { // migrate.ErrLockTimeout will be thrown when another process is doing migration and timeout.
		logger.Errorf("Failed to upgrade schema, error: %q", err)
		return err
	}
	return nil
}

// NewPGSQL returns an instance of postgres
func NewPGSQL(host string, port string, usr string, pwd string, database string, sslmode string, maxIdleConns int, maxOpenConns int) Database {
	if len(sslmode) == 0 {
		sslmode = "disable"
	}
	return &pgsql{
		host:         host,
		port:         port,
		usr:          usr,
		pwd:          pwd,
		database:     database,
		sslmode:      sslmode,
		maxIdleConns: maxIdleConns,
		maxOpenConns: maxOpenConns,
	}
}
