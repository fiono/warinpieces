package main

import (
  "fmt"
  "log"
  "net/http"

  "books"
  "mail"
  "views"

  "github.com/gorilla/mux"
  "google.golang.org/appengine"
)

func main() {
    r := mux.NewRouter()

    // BIGF: temp testing endpoint
    r.HandleFunc("/send", emailHandler)

    // Serve static assets
    r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    /*
     Views
    */
    r.HandleFunc("/books/", (&views.TplRenderer{Tpl: "book"}).RenderView).Methods("GET")

    r.HandleFunc("/", (&views.TplRenderer{
      "subscription_form",
      views.SubscriptionFormView{"new subscription", "/api/subscriptions/new/"},
    }).RenderView).Methods("GET")

    r.HandleFunc("/deactivate/", (&views.TplRenderer{
      "subscription_form",
      views.SubscriptionFormView{"pause subscription", "/api/subscriptions/deactivate/"},
    }).RenderView).Methods("GET")

    r.HandleFunc("/reactivate/", (&views.TplRenderer{
      "subscription_form",
      views.SubscriptionFormView{"reactivate subscription", "/api/subscriptions/reactivate/"},
    }).RenderView).Methods("GET")

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

func emailHandler(w http.ResponseWriter, r *http.Request) {
  ctx := appengine.NewContext(r)
  err := mail.SendMail("fiona@witches.nyc", "hey fiona", "<strong>whats good</strong>", ctx)
  if err != nil {
    logAndPrintError(w, err)
    return
  }

  fmt.Fprintln(w, "Sent mail")
}

func newBookHandler(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()

  bookId := r.Form["bookId"][0]
  delimiter := r.Form["delim"][0]

  ctx := appengine.NewContext(r)
  meta, err := books.ChapterizeBook(bookId, delimiter, ctx)
  if err != nil {
    logAndPrintError(w, err)
    return
  }

  _, err = NewBook(meta)
  if err != nil {
    logAndPrintError(w, err)
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
    logAndPrintError(w, err)
    return
  }

  fmt.Fprintf(w, "New subscription with ID %s and address %s", bookId, emailAddr)
}

func logAndPrintError(w http.ResponseWriter, err error) {
  fmt.Fprintln(w, err)
  log.Println(err)
}
