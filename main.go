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

  // Nightly cron endpoint
  r.HandleFunc("/send/", cronHandler)

  // Serve static assets
  r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

  // Views
  r.HandleFunc("/books/", (&views.TplRenderer{Tpl: "book", IsWeb: true}).ServeView).Methods("GET")
  r.HandleFunc("/", newSubscriptionView).Methods("GET")
  //r.HandleFunc("/success", subscriptionSuccessView).Methods("GET")
  //r.HandleFunc("/validate", validateView).Methods("POST")
  //r.HandleFunc("/unsubscribe", unsubscribeView).Methods("POST")

  // Endpoints
  r.HandleFunc("/api/books/new/", newBookHandler).Methods("POST")
  r.HandleFunc("/api/subscriptions/new/", newSubscriptionHandler).Methods("POST")
  //r.HandleFunc("/api/subscriptions/reactivate/{subscription_id}", reactivateSubscriptionHandler).Methods("GET")
  //r.HandleFunc("/api/subscriptions/validate/{subscription_id}", validateSubscriptionHandler).Methods("GET")

  http.Handle("/", r)
  appengine.Main()
}

/*
  Views
*/
func newSubscriptionView(w http.ResponseWriter, r *http.Request) {
  db, err := dbConn()
  if err != nil {
    reportError(err)
  }
  defer db.Close()

  validBooks, err := db.GetBooks()
  if err != nil {
    reportError(err)
  }

  views.NewSubscriptionRenderer(validBooks).ServeView(w, r)
}


/*
  Cron logic
*/

type sendEmailResponse struct {
  subId string
  err error
}

func cronHandler(w http.ResponseWriter, r *http.Request) {
  db, err := dbConn()
  if err != nil {
    reportError(err)
  }
  defer db.Close()

  ctx := appengine.NewContext(r)

  subs, err := db.GetSubscriptionsForSending()
  if err != nil {
    reportError(err)
  }

  ch := make(chan sendEmailResponse)
  for _, sub := range subs {
    go sendEmailForSubscription(sub.SubscriptionId, ctx, ch)
  }

  successes := 0
  tries := 0
  for ; tries < len(subs); tries++ {
    resp := <-ch
    if resp.err == nil {
      successes++
    }
    if err = db.NewEmailAudit(resp.subId, 0, resp.err != nil); err != nil {
      reportError(err)
    }
  }

  fmt.Fprintf(w, "Done, %d successes out of %d tries", successes, tries)
}

func sendEmailForSubscription(subscriptionId string, ctx context.Context, ch chan sendEmailResponse) {
  db, err := dbConn()
  if err != nil {
    reportError(err)
  }
  defer db.Close()

  defer func() {
    ch <- sendEmailResponse{subscriptionId, err}
  }()

  sub, err := db.GetSubscription(subscriptionId)
  if err != nil {
    return
  }

  bookMeta, err := db.GetBook(sub.BookId)
  if err != nil {
    return
  }

  body, err := books.GetChapter(sub.BookId, sub.ChaptersSent + 1, ctx)
  if err != nil {
    return
  }

  content := strings.Replace(body, "\n", "<br/>", -1)
  emailBody, err := views.NewEmailRenderer(bookMeta, sub, content).GetView()
  if err != nil {
    return
  }

  if err = SendMail(sub.Email, bookMeta.Title, emailBody, ctx); err != nil {
    return
  }

  if err = db.IncrementChaptersSent(sub.SubscriptionId); err != nil {
    return
  }
}

/*
  Endpoints
*/
func newBookHandler(w http.ResponseWriter, r *http.Request) {
  db, err := dbConn()
  if err != nil {
    reportError(err)
  }
  defer db.Close()

  r.ParseForm()

  bookId := r.Form["bookId"][0]
  delimiter := r.Form["delim"][0]

  ctx := appengine.NewContext(r)
  meta, err := books.ChapterizeBook(bookId, delimiter, ctx)
  if err != nil {
    reportError(err)
    return
  }

  err = db.NewBook(books.BookMeta{
    bookId,
    meta.Title,
    meta.Author,
    meta.Chapters,
    delimiter,
    0, // BIGF
  })
  if err != nil {
    reportError(err)
    return
  }

  fmt.Fprintf(w, "New book: ", meta)
}

func newSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
  db, err := dbConn()
  if err != nil {
    reportError(err)
  }
  defer db.Close()

  r.ParseForm()

  bookId := r.Form["bookId"][0]
  emailAddr := r.Form["email"][0]

  if err := db.NewSubscription(bookId, emailAddr); err != nil {
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
      panic(err)
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
