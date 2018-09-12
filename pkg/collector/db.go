package collector

import (
	"database/sql"
	"fmt"
	"net/url"

	_ "github.com/denisenkom/go-mssqldb" // import mssql for database connection
	"github.com/pkg/errors"
)

func OpenSQL(cfg Config) (*sql.DB, error) {
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

	return db, errors.Wrap(err, "could not connect to sqlserver")
}
