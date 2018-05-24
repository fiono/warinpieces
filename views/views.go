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

type emailView struct {
  Title string
  Author string
  Chapter int
  Body string
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

func NewEmailRenderer(book books.BookMeta, sub books.SubscriptionMeta, content string) *TplRenderer {
  return &TplRenderer{
    "email",
    emailView{book.Title, book.Author, sub.ChaptersSent + 1, content},
    false,
  }
}
