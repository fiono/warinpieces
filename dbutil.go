package main

import (
  "crypto/md5"
  "database/sql"
  "fmt"

  "books"
  "config"


  "google.golang.org/appengine"
  _ "github.com/go-sql-driver/mysql"
)

type DbConn struct {
  Conn *sql.DB
  Cfg config.GutenConfig
}

func dbConn() (db *DbConn, err error) {
  cfg := config.LoadConfig()

  var conn *sql.DB
  if appengine.IsDevAppServer() {
    conn, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", cfg.Db.User, cfg.Db.Password, cfg.Db.DbName))
  } else {
    conn, err = sql.Open("mysql", fmt.Sprintf("%s:%s@cloudsql(%s)/%s", cfg.Db.User, cfg.Db.Password, cfg.Db.ConnectionName, cfg.Db.DbName))
  }
  return &DbConn{conn, cfg}, err
}

func (db *DbConn) Close() {
  db.Conn.Close()
}

/*
  Book utils
*/

func (db *DbConn) GetBooks() (b []books.BookMeta, err error) {
  rows, err := db.Conn.Query(
    "SELECT book_id, title, author, chapter_count, chapter_delim, publishing_schedule_type FROM books",
  )
  if err != nil {
    return
  }
  defer rows.Close()

  for rows.Next() {
    var bookId, title, author, chapterDelim string
    var chapters, scheduleType int

    if err = rows.Scan(&bookId, &title, &author, &chapters, &chapterDelim, &scheduleType); err != nil {
      return
    }
    b = append(
      b,
      books.BookMeta{bookId, title, author, chapters, chapterDelim, scheduleType},
    )
  }
  if err = rows.Err(); err != nil {
    return
  }
  return b, nil
}

func (db *DbConn) GetBook(bookId string) (book books.BookMeta, err error) {
  var title, author, chapterDelim string
  var chapters int

  err = db.Conn.QueryRow(
    "SELECT title, author, chapter_count, chapter_delim FROM books WHERE book_id = ?",
    bookId,
  ).Scan(&title, &author, &chapters, &chapterDelim)
  if err != nil {
    return
  }

  return books.BookMeta{bookId, title, author, chapters, chapterDelim, 0}, nil // BIGF
}

func (db *DbConn) NewBook(bookMeta books.BookMeta) error {
  _, err := db.Conn.Exec(
    "INSERT INTO books (book_id, title, author, chapter_count, chapter_delim, publishing_schedule_type) VALUES (?, ?, ?, ?, ?, ?)",
    bookMeta.BookId,
    bookMeta.Title,
    bookMeta.Author,
    bookMeta.Chapters,
    bookMeta.Delimiter,
    bookMeta.ScheduleType,
  )
  return err
}

/*
  Subscription utils
*/

func getSubscriptionToken(bookId, emailAddress string) string {
  cfg := config.LoadConfig()
  h := md5.Sum([]byte(fmt.Sprintf("%s-%s-%s", bookId, emailAddress, cfg.Main.HashSalt)))
  return fmt.Sprintf("%x", h)
}

func (db *DbConn) NewSubscription(bookId, emailAddr string) error {
  _, err := db.Conn.Exec(
    "INSERT INTO subscriptions (subscription_id, book_id, email_address, create_datetime) VALUES (DEFAULT, ?, ?, NOW())",
    bookId,
    emailAddr,
  )
  return err
}

func (db *DbConn) GetSubscription(subscriptionId string) (sub books.SubscriptionMeta, err error) {
  var bookId, emailAddress string
  var chaptersSent int
  var isActive, isValidated bool

  err = db.Conn.QueryRow(
    "SELECT book_id, email_address, chapters_sent, is_active, is_validated FROM subscriptions WHERE subscription_id = ?",
    subscriptionId,
  ).Scan(&bookId, &emailAddress, &chaptersSent, &isActive, &isValidated)
  if err != nil {
    return
  }

  return books.SubscriptionMeta{subscriptionId, bookId, emailAddress, chaptersSent, isActive, isValidated}, nil
}

func (db *DbConn) GetSubscriptionByData(bookId, emailAddress string) (sub books.SubscriptionMeta, err error) {
  var subscriptionId string
  var chaptersSent int
  var isActive, isValidated bool

  err = db.Conn.QueryRow(
    "SELECT subscription_id, chapters_sent, is_active, is_validated FROM subscriptions WHERE book_id = ? AND email_address = ?",
    bookId,
    emailAddress,
  ).Scan(&subscriptionId, &chaptersSent, &isActive, &isValidated)
  if err != nil {
    return
  }

  return books.SubscriptionMeta{subscriptionId, bookId, emailAddress, chaptersSent, isActive, isValidated}, nil
}

func (db *DbConn) GetSubscriptionsForSending() (subs []books.SubscriptionMeta, err error) {
  rows, err := db.Conn.Query(
    `SELECT
      subscription_id, book_id, email_address, chapters_sent, is_active, is_validated
      FROM subscriptions
      WHERE is_active = true
        AND is_validated = true AND
        DAYOFWEEK(DATE_SUB(create_datetime, INTERVAL 1 DAY)) = DAYOFWEEK(NOW())`, // temp hack to avoid resending 1st chapter
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

func (db *DbConn) ActivateSubscription(bookId, emailAddress string) error {
  _, err := db.Conn.Exec(
    "UPDATE subscriptions SET is_active = 1, is_validated = 1 WHERE email_address = ? AND book_id = ?",
    emailAddress,
    bookId,
  )
  return err
}

func (db *DbConn) IncrementChaptersSent(subscriptionId string) error {
  _, err := db.Conn.Exec(
    "UPDATE subscriptions SET chapters_sent = chapters_sent + 1 WHERE subscription_id = ?",
    subscriptionId,
  )
  return err
}

func (db *DbConn) UnsubscribeSingle(emailAddress, bookId string) error {
  _, err := db.Conn.Exec(
    "UPDATE subscriptions SET is_active = 0 WHERE email_address = ? and book_id = ?",
    emailAddress,
    bookId,
  )
  return err
}

func (db *DbConn) UnsubscribeEmail(emailAddress string) error {
  _, err := db.Conn.Exec(
    "UPDATE subscriptions SET is_active = 0 WHERE email_address = ?",
    emailAddress,
  )
  return err
}

/*
  Email audit utils
*/

func (db *DbConn) NewEmailAudit(subscriptionId string, emailLen int, success bool) error {
  _, err := db.Conn.Exec(
    "INSERT INTO email_audit (subscription_id, email_len, send_datetime, is_success) VALUES (?, ?, NOW(), ?)",
    subscriptionId,
    emailLen,
    success,
  )
  return err
}

