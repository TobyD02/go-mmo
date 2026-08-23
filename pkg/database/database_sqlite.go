package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type DBSqlite struct {
	conn *sql.DB
	tx   *sql.Tx
}

func NewDBSqlite() *DBSqlite {
	return &DBSqlite{
		conn: nil,
	}
}

func (d *DBSqlite) Connect() error {

	if d.conn != nil {
		return nil
	}

	sqlitePath := os.Getenv("SQLITE_PATH")
	if sqlitePath == "" {
		return fmt.Errorf("no sqlite path found")
	}

	db, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		return err
	}

	d.conn = db
	return nil
}

func (d *DBSqlite) BeginTransaction() error {
	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}

	if d.tx != nil {
		err = tx.Rollback()
		return fmt.Errorf("cannot begin new transaction whilst one is already open | %s", err)
	}

	d.tx = tx
	return nil
}

func (d *DBSqlite) RollbackTransaction() error {
	if d.tx == nil {
		return fmt.Errorf("no active transaction")
	}

	err := d.tx.Rollback()
	d.tx = nil

	return err
}

func (d *DBSqlite) CommitTransaction() error {
	if d.tx == nil {
		return fmt.Errorf("no active transaction")
	}

	err := d.tx.Commit()
	d.tx = nil

	return err
}

func (d *DBSqlite) Exec(query string, args ...any) (sql.Result, error) {
	if d.tx != nil {
		return d.tx.Exec(query, args...)
	} else {
		return d.conn.Exec(query, args...)
	}
}

func (d *DBSqlite) Query(query string, args ...any) (*sql.Rows, error) {
	if d.tx != nil {
		return d.tx.Query(query, args...)
	} else {
		return d.conn.Query(query, args...)
	}
}

func (d *DBSqlite) QueryRow(query string, args ...any) *sql.Row {
	if d.tx != nil {
		return d.tx.QueryRow(query, args...)
	} else {
		return d.conn.QueryRow(query, args...)
	}
}

func (d *DBSqlite) Close() {

	if d.tx != nil {
		d.tx.Rollback()
		d.tx = nil
	}

	if d.conn != nil {
		d.conn.Close()
		d.conn = nil
	}
}
