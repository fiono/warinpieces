package views

import (
  "fmt"
  "net/url"

  "books"
  "config"
)

type emailView struct {
  Book books.BookMeta
  Chapter int
  Body string
  HomeUrl string
  UnsubUrl string
}


func NewSubscriptionRenderer(bookOptions []books.BookMeta) *TplRenderer {
  return &TplRenderer{
    "subscription_form",
    struct {
      Endpoint string
      BookOptions []books.BookMeta
    } { "/subscriptions/new/", bookOptions },
    true,
  }
}

func SubscriptionSuccessRenderer(book books.BookMeta, email string) *TplRenderer {
  return &TplRenderer{
    "subscription_success",
    struct {
      Book books.BookMeta
      EmailAddress string
    } { book, email },
    true,
  }
}

func UnsubscriptionSuccessRenderer(emailAddress string, book books.BookMeta) *TplRenderer {
  return &TplRenderer{
    "single_unsub_success",
    struct {
      Book books.BookMeta
      EmailAddress string
    } { book, emailAddress },
    true,
  }
}

func EmailUnsubscriptionSuccessRenderer(emailAddress string, book books.BookMeta) *TplRenderer {
  return &TplRenderer{
    "email_unsub_success",
    struct {
      EmailAddress string
    } { emailAddress },
    true,
  }
}

func ConfirmationSuccessRenderer(emailAddress string, book books.BookMeta) *TplRenderer {
  return &TplRenderer{
    "confirm_success",
    struct {
      Book books.BookMeta
      EmailAddress string
    } { book, emailAddress },
    true,
  }
}

func EmailRenderer(token, content string, book books.BookMeta, sub books.SubscriptionMeta) *TplRenderer {
  cfg := config.LoadConfig()

  urlBase := cfg.Main.UrlBase
  params := url.Values{"email_address": {sub.Email}, "book_id": {book.BookId}, "token": {token}}
  unsubUrl := fmt.Sprintf("%s/subscriptions/unsubscribe/?%s", urlBase, params.Encode())

  return &TplRenderer{
    "email",
    emailView{book, sub.ChaptersSent + 1, content, urlBase, unsubUrl},
    false,
  }
}

func ConfirmEmailRenderer(emailAddress, token string, book books.BookMeta) *TplRenderer {
  cfg := config.LoadConfig()

  params := url.Values{"email_address": {emailAddress}, "book_id": {book.BookId}, "token": {token}}
  confirmUrl := fmt.Sprintf("%s/subscriptions/confirm/?%s", cfg.Main.UrlBase, params.Encode())

  return &TplRenderer{
    "confirm_email",
    struct {
      Book books.BookMeta
      ConfirmUrl string
    } { book, confirmUrl },
    false,
  }
}
