package collector

import (
	"database/sql"
	"fmt"
	"net/url"

	_ "github.com/denisenkom/go-mssqldb" // import mssql for database connection
	"github.com/pkg/errors"
)

// OpenSQL opens a mssql database connection by given config
func OpenSQL(cfg DBConfig) (*sql.DB, error) {
	query := url.Values{}
	query.Add("database", cfg.Database)
	query.Add("encrypt", "disable")

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     fmt.Sprintf("%s:%d", cfg.Server, cfg.Port),
		RawQuery: query.Encode(),
	}

	db, err := sql.Open("sqlserver", u.String())
	if err != nil {
		return nil, errors.Wrap(err,"could not open db connection pool")
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, errors.Wrap(err, "could not connect to sqlserver")
	}
	return db, nil
}
