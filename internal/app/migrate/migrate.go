package appmigrate

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // register pgx driver for database/sql
	"github.com/pressly/goose/v3"
)

func Run(dsn, dir, command string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.RunContext(context.Background(), command, db, dir)
}
