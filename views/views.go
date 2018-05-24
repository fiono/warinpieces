package views

import (
  "books"
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
  SubscriptionId string
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

func UnsubscriptionSuccessRenderer(book books.BookMeta, sub books.SubscriptionMeta) *TplRenderer {
  return &TplRenderer{
    "single_unsub_success",
    unsubscriptionSuccessView{book, sub.Email},
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

func NewEmailRenderer(book books.BookMeta, sub books.SubscriptionMeta, content string) *TplRenderer {
  return &TplRenderer{
    "email",
    emailView{book.Title, book.Author, sub.ChaptersSent + 1, content, sub.SubscriptionId},
    false,
  }
}
