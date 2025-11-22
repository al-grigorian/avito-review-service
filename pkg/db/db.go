package db

import (
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
)

func NewDB(connStr string) (*sqlx.DB, error) {
    return sqlx.Open("postgres", connStr)
}