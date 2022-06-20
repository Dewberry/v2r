package processing

import (
	"context"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

type dbConfig struct {
	dbUser string
	dbPass string
	dbHost string
	dbPort string
	dbName string
}

func DBInit() *sqlx.DB {
	conf := dbConfig{"docker", "docker", "v2r-db", "5432", "dev"}
	creds := fmt.Sprintf("user=%s password=%s host=%s port=%s database=%s sslmode=disable",
		conf.dbUser, conf.dbPass, conf.dbHost, conf.dbPort, conf.dbName)
	return sqlx.MustOpen("pgx", creds)
}

// PingWithTimeout ...
func PingWithTimeout(db *sqlx.DB) error {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}
