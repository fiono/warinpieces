package main

import (
  "database/sql"

  "books"
  "config"

  _ "github.com/lib/pq"
)

func dbConn() (db *sql.DB, err error) {
  cfg := config.LoadConfig()
	db, err = sql.Open("postgres", cfg.Db.ConnectionStr)
	if err != nil {
    return nil, err
	}
  return
}

func NewBook(bookMeta books.BookMeta) (res sql.Result, err error) {
  db, err := dbConn()
  if err != nil {
    return nil, err
  }

  res, err = db.Exec(
    "INSERT INTO books (book_id, title, author, chapter_count, chapter_delim) values ($1, $2, $3, $4, $5)",
    bookMeta.BookId,
    bookMeta.Title,
    bookMeta.Author,
    bookMeta.Chapters,
    bookMeta.Delimiter,
  )
  return
}

func NewSubscription(bookId string, emailAddr string) (res sql.Result, err error) {
  db, err := dbConn()
  if err != nil {
    return nil, err
  }

  res, err = db.Exec(
    "INSERT INTO subscriptions (subscription_id, book_id, email_address, create_datetime) values (DEFAULT, $1, $2, NOW())",
    bookId,
    emailAddr,
  )
  return
}
