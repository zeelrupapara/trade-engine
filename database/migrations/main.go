package database

import (
	"database/sql"
	"strconv"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/zeelrupapara/trade-engine/config"
)

var db *sql.DB
var dbURL string
var err error

const POSTGRES = "postgres"

// Connect with database
func Connect(cfg config.DBConfig) (*goqu.Database, error) {
	return postgresDBConnection(cfg)
}

func postgresDBConnection(cfg config.DBConfig) (*goqu.Database, error) {
	dbURL = "postgres://" + cfg.Username + ":" + cfg.Password + "@" + cfg.Host + ":" + strconv.Itoa(cfg.Port) + "/" + cfg.Db + "?" + cfg.QueryString
	if db == nil {
		db, err = sql.Open(POSTGRES, dbURL)
		if err != nil {
			return nil, err
		}
		return goqu.New(POSTGRES, db), err
	}
	return goqu.New(POSTGRES, db), err
}
