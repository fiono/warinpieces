package main

import (
  "fmt"
  "log"
  "net/http"
  "strings"

  "books"
  "views"

  "cloud.google.com/go/errorreporting"
  "github.com/gorilla/mux"
  "golang.org/x/net/context"
  "google.golang.org/appengine"
)

func main() {
  r := mux.NewRouter()

  /*
    Nightly cron endpoint
  */
  r.HandleFunc("/send/", cronHandler)

  /*
   Serve static assets
  */
  r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

  /*
   Views
  */
  r.HandleFunc("/books/", (&views.TplRenderer{Tpl: "book"}).ServeView).Methods("GET")

  r.HandleFunc("/", (&views.TplRenderer{
    "subscription_form",
    views.SubscriptionFormView{"new subscription", "/api/subscriptions/new/"},
    true,
  }).ServeView).Methods("GET")

  r.HandleFunc("/deactivate/", (&views.TplRenderer{
    "subscription_form",
    views.SubscriptionFormView{"pause subscription", "/api/subscriptions/deactivate/"},
    true,
  }).ServeView).Methods("GET")

  r.HandleFunc("/reactivate/", (&views.TplRenderer{
    "subscription_form",
    views.SubscriptionFormView{"reactivate subscription", "/api/subscriptions/reactivate/"},
    true,
  }).ServeView).Methods("GET")

  //r.HandleFunc("/validate", validateView).Methods("POST")

  /*
   Endpoints
  */
  r.HandleFunc("/api/books/new/", newBookHandler).Methods("POST")
  r.HandleFunc("/api/subscriptions/new/", newSubscriptionHandler).Methods("POST")
  //r.HandleFunc("/api/subscriptions/validate/{subscription_id}", validateSubscriptionHandler).Methods("GET")
  //r.HandleFunc("/api/subscriptions/deactivate/{subscription_id}", deactivateSubscriptionHandler).Methods("GET")
  //r.HandleFunc("/api/subscriptions/reactivate/{subscription_id}", reactivateSubscriptionHandler).Methods("GET")

  http.Handle("/", r)
  appengine.Main()
}

func sendEmailForSubscription(subscriptionId string, ctx context.Context) error {
  sub, err := GetSubscription(subscriptionId)
  if err != nil {
    return err
  }

  bookMeta, err := GetBook(sub.BookId)
  if err != nil {
    return err
  }

  body, err := books.GetChapter(sub.BookId, sub.ChaptersSent + 1, ctx)
  if err != nil {
    return err
  }

  content := strings.Replace(body, "\n", "<br/>", -1)
  emailBody, err := (&views.TplRenderer{
    "email",
    views.EmailView{bookMeta.Title, bookMeta.Author, sub.ChaptersSent + 1, content},
    false,
  }).GetView()
  if err != nil {
    return err
  }

  if err = SendMail(sub.Email, bookMeta.Title, emailBody, ctx); err != nil {
    return err
  }

  return IncrementChaptersSent(sub.SubscriptionId)
}

func cronHandler(w http.ResponseWriter, r *http.Request) {
  ctx := appengine.NewContext(r)

  //ch := make(chan string)
  //ec := make(chan int)
  subs, err := GetActiveSubscriptions()
  if err != nil {
    reportError(err)
  }
  for _, sub := range subs {
    err := sendEmailForSubscription(sub.SubscriptionId, ctx)
    if err != nil {
      reportError(err)
      NewEmailAudit(sub.SubscriptionId, 0, false)
    } else {
      NewEmailAudit(sub.SubscriptionId, 0, true)
    }
  }

  fmt.Fprintln(w, "Done")
}

func newBookHandler(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()

  bookId := r.Form["bookId"][0]
  delimiter := r.Form["delim"][0]

  ctx := appengine.NewContext(r)
  meta, err := books.ChapterizeBook(bookId, delimiter, ctx)
  if err != nil {
    reportError(err)
    return
  }

  _, err = NewBook(meta)
  if err != nil {
    reportError(err)
    return
  }

  fmt.Fprintf(w, "New book: ", meta)
}

func newSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()

  bookId := r.Form["bookId"][0]
  emailAddr := r.Form["email"][0]

  _, err := NewSubscription(bookId, emailAddr)
  if err != nil {
    reportError(err)
    return
  }

  fmt.Fprintf(w, "New subscription with ID %s and address %s", bookId, emailAddr)
}

func getErrorReportingClient(projectId string) (client *errorreporting.Client, err error) {
  ctx := context.Background()
  return errorreporting.NewClient(ctx, projectId, errorreporting.Config{
    ServiceName: "gutenbits",
      OnError: func(err error) {
      log.Printf("Could not log error: %v", err)
    },
  })
}

func reportError(err error) {
  errorClient, e := getErrorReportingClient("gutenbits")
  if e != nil {
    panic(e)
  }
  defer errorClient.Close()
  defer errorClient.Flush()

  log.Println(err)
  errorClient.Report(errorreporting.Entry{
    Error: err,
  })
}
