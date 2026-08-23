package database

import "database/sql"

type DB interface {
	Connect() error
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row

	BeginTransaction() error
	RollbackTransaction() error
	CommitTransaction() error

	Close()
}
