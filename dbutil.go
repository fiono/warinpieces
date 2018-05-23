package main

import (
  "database/sql"

  "books"
  "config"

  "google.golang.org/appengine"
  _ "github.com/lib/pq"
)

func dbConn() (db *sql.DB, err error) {
  cfg := config.LoadConfig()
  if appengine.IsDevAppServer() {
    return sql.Open("postgres", cfg.Db.ConnectionStrDev)
  } else {
    return sql.Open("postgres", cfg.Db.ConnectionStr)
  }
}

/*
  Book utils
*/

func GetBook(bookId string) (book books.BookMeta, err error) {
  db, err := dbConn()
  if err != nil {
    return
  }
  defer db.Close()

  var title, author, chapterDelim string
  var chapters int

  err = db.QueryRow(
    "SELECT title, author, chapter_count, chapter_delim FROM books WHERE book_id = $1",
    bookId,
  ).Scan(&title, &author, &chapters, &chapterDelim)
  if err != nil {
    return
  }

  return books.BookMeta{bookId, title, author, chapters, chapterDelim}, nil
}

func NewBook(bookMeta books.BookMeta) (res sql.Result, err error) {
  db, err := dbConn()
  if err != nil {
    return
  }
  defer db.Close()

  return db.Exec(
    "INSERT INTO books (book_id, title, author, chapter_count, chapter_delim) VALUES ($1, $2, $3, $4, $5)",
    bookMeta.BookId,
    bookMeta.Title,
    bookMeta.Author,
    bookMeta.Chapters,
    bookMeta.Delimiter,
  )
}

/*
  Subscription utils
*/

func NewSubscription(bookId, emailAddr string) (res sql.Result, err error) {
  db, err := dbConn()
  if err != nil {
    return
  }
  defer db.Close()

  return db.Exec(
    "INSERT INTO subscriptions (subscription_id, book_id, email_address, create_datetime) VALUES (DEFAULT, $1, $2, NOW())",
    bookId,
    emailAddr,
  )
}

func GetSubscription(subscriptionId string) (sub books.SubscriptionMeta, err error) {
  db, err := dbConn()
  if err != nil {
    return
  }
  defer db.Close()

  var bookId, emailAddress string
  var chaptersSent int
  var isActive, isValidated bool

  err = db.QueryRow(
    "SELECT book_id, email_address, chapters_sent, is_active, is_validated FROM subscriptions WHERE subscription_id = $1",
    subscriptionId,
  ).Scan(&bookId, &emailAddress, &chaptersSent, &isActive, &isValidated)
  if err != nil {
    return
  }

  return books.SubscriptionMeta{subscriptionId, bookId, emailAddress, chaptersSent, isActive, isValidated}, nil
}

func GetActiveSubscriptions() (subs []books.SubscriptionMeta, err error) {
  db, err := dbConn()
  if err != nil {
    return
  }
  defer db.Close()

  rows, err := db.Query(
    "SELECT subscription_id, book_id, email_address, chapters_sent, is_active, is_validated FROM subscriptions WHERE is_active=true",
  )
  if err != nil {
    return
  }
  defer rows.Close()

  for rows.Next() {
    var subscriptionId, bookId, emailAddress string
    var chaptersSent int
    var isActive, isValidated bool

    if err = rows.Scan(&subscriptionId, &bookId, &emailAddress, &chaptersSent, &isActive, &isValidated); err != nil {
      return
    }

    subs = append(
      subs,
      books.SubscriptionMeta{subscriptionId, bookId, emailAddress, chaptersSent, isActive, isValidated},
    )
  }
  if err = rows.Err(); err != nil {
    return
  }
  return subs, nil
}

func IncrementChaptersSent(subscriptionId string) error {
  db, err := dbConn()
  if err != nil {
    return err
  }
  defer db.Close()

  _, err = db.Exec(
    "UPDATE subscriptions SET chapters_sent = chapters_sent + 1 WHERE subscription_id = $1",
    subscriptionId,
  )
  return err
}

/*
  Email audit utils
*/

func NewEmailAudit(subscriptionId string, emailLen int, success bool) (res sql.Result, err error) {
  db, err := dbConn()
  if err != nil {
    return
  }
  defer db.Close()

  return db.Exec(
    "INSERT INTO email_audit (subscription_id, email_len, send_datetime, is_success) VALUES ($1, $2, NOW(), $3)",
    subscriptionId,
    emailLen,
    success,
  )
}

