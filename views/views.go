package views

import (
  "fmt"
  "net/url"

  "books"
  "config"
)

type subscriptionFormView struct {
  Title string
  Endpoint string
  BookOptions []books.BookMeta
}

type subscriptionSuccessView struct {
  Book books.BookMeta
  EmailAddress string
}

type unsubscriptionSuccessView struct {
  Book books.BookMeta
  EmailAddress string
}

type emailView struct {
  Title string
  Author string
  Chapter int
  Body string
  HomeUrl string
  UnsubUrl string
}

func NewSubscriptionRenderer(bookOptions []books.BookMeta) *TplRenderer {
  return &TplRenderer{
    "subscription_form",
    subscriptionFormView{"new subscription", "/api/subscriptions/new/", bookOptions},
    true,
  }
}

func SubscriptionSuccessRenderer(book books.BookMeta, email string) *TplRenderer {
  return &TplRenderer{
    "subscription_success",
    subscriptionSuccessView{book, email},
    true,
  }
}

func UnsubscriptionSuccessRenderer(book books.BookMeta, emailAddress string) *TplRenderer {
  return &TplRenderer{
    "single_unsub_success",
    unsubscriptionSuccessView{book, emailAddress},
    true,
  }
}

func EmailUnsubscriptionSuccessRenderer(emailAddress string) *TplRenderer {
  return &TplRenderer{
    "email_unsub_success",
    struct {
      EmailAddress string
    } { emailAddress },
    true,
  }
}

func NewEmailRenderer(book books.BookMeta, sub books.SubscriptionMeta, token string, content string) *TplRenderer {
  cfg := config.LoadConfig()
  urlBase := cfg.Main.UrlBase

  params := url.Values{"email_address": {sub.Email}, "book_id": {book.BookId}, "token": {token}}
  unsubUrl := fmt.Sprintf("%s/unsubscribe/?%s", cfg.Main.UrlBase, params.Encode())
  return &TplRenderer{
    "email",
    emailView{book.Title, book.Author, sub.ChaptersSent + 1, content, urlBase, unsubUrl},
    false,
  }
}
