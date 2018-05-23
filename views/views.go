package views

type BookOption struct {
  BookId string
  Title string
  Author string
}

type SubscriptionFormView struct {
  Title string
  Endpoint string
  Options []BookOption
}

type EmailView struct {
  Title string
  Author string
  Chapter int
  Body string
}
