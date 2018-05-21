package main

import (
  "database/sql"
  "config"
  _ "github.com/lib/pq"
)

func DbConn() (db *sql.DB, err error) {
  cfg := config.LoadConfig()
	db, err = sql.Open("postgres", cfg.Db.ConnectionStr)
	if err != nil {
    return nil, err
	}
  return
}
