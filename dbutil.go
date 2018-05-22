package main

import (
  "database/sql"

  "books"
  "config"

  _ "github.com/lib/pq"
)

func dbConn() (db *sql.DB, err error) {
  cfg := config.LoadConfig()
  return sql.Open("postgres", cfg.Db.ConnectionStr)
}

func NewBook(bookMeta books.BookMeta) (res sql.Result, err error) {
  db, err := dbConn()
  defer db.Close()
  if err != nil {
    return
  }

  return db.Exec(
    "INSERT INTO books (book_id, title, author, chapter_count, chapter_delim) VALUES ($1, $2, $3, $4, $5)",
    bookMeta.BookId,
    bookMeta.Title,
    bookMeta.Author,
    bookMeta.Chapters,
    bookMeta.Delimiter,
  )
}

func NewEmailAudit(subscriptionId string, emailLen int, success bool) (res sql.Result, err error) {
  db, err := dbConn()
  defer db.Close()
  if err != nil {
    return
  }

  return db.Exec(
    "INSERT INTO email_audit (subscription_id, email_len, send_datetime, is_success) VALUES ($1, $2, NOW(), $3)",
    subscriptionId,
    emailLen,
    success,
  )
}

func NewSubscription(bookId, emailAddr string) (res sql.Result, err error) {
  db, err := dbConn()
  defer db.Close()
  if err != nil {
    return nil, err
  }

  res, err = db.Exec(
    "INSERT INTO subscriptions (subscription_id, book_id, email_address, create_datetime) VALUES (DEFAULT, $1, $2, NOW())",
    bookId,
    emailAddr,
  )
  return
}

func GetBook(book_id string) (book books.BookMeta, err error) {
  db, err := dbConn()
  defer db.Close()
  if err != nil {
    return
  }

  var title, author, chapter_delim string
  var chapters int

  err = db.QueryRow(
    "SELECT title, author, chapter_count, chapter_delim FROM books WHERE book_id = $1",
    book_id,
  ).Scan(&title, &author, &chapters, &chapter_delim)
  if err != nil {
    return
  }

  return books.BookMeta{book_id, title, author, chapters, chapter_delim}, nil
}

func GetSubscription(subscription_id string) (sub books.SubscriptionMeta, err error) {
  db, err := dbConn()
  defer db.Close()
  if err != nil {
    return
  }

  var book_id, email_address string
  var chapters_sent int
  var is_active, is_validated bool

  err = db.QueryRow(
    "SELECT book_id, email_address, chapters_sent, is_active, is_validated FROM subscriptions WHERE subscription_id = $1",
    subscription_id,
  ).Scan(&book_id, &email_address, &chapters_sent, &is_active, &is_validated)
  if err != nil {
    return
  }

  return books.SubscriptionMeta{subscription_id, book_id, email_address, chapters_sent, is_active, is_validated}, nil
}

func IncrementChaptersSent(subscription_id string) error {
  db, err := dbConn()
  defer db.Close()
  if err != nil {
    return err
  }

  _, err = db.Exec(
    "UPDATE subscriptions SET chapters_sent = chapters_sent + 1 WHERE subscription_id = $1",
    subscription_id,
  )
  return err
}
