package main

import (
    "fmt"
    "log"
    "net/http"
    "text/template"

    "books"
    "mail"

    "github.com/gorilla/mux"
    "google.golang.org/appengine"
)

func main() {
    r := mux.NewRouter()

    // BIGF: temp testing endpoint
    r.HandleFunc("/send", emailHandler)

    // Serve static assets
    r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    // Views
    r.HandleFunc("/", (&tplRenderer{"index"}).renderView).Methods("GET")
    r.HandleFunc("/books/", (&tplRenderer{"book"}).renderView).Methods("GET")
    //r.HandleFunc("/deactivate", deactivateView).Methods("GET")
    //r.HandleFunc("/reactivate", reactivateView).Methods("GET")
    //r.HandleFunc("/validate", validateView).Methods("POST")

    // Endpoints
    r.HandleFunc("/api/books/new/", newBookHandler).Methods("POST")
    //r.HandleFunc("/api/subscriptions/new/", newSubscriptionHandler).Methods("POST")
    //r.HandleFunc("/api/subscriptions/validate/{subscription_id}", validateSubscriptionHandler).Methods("GET")
    //r.HandleFunc("/api/subscriptions/deactivate/{subscription_id}", deactivateSubscriptionHandler).Methods("GET")
    //r.HandleFunc("/api/subscriptions/reactivate/{subscription_id}", reactivateSubscriptionHandler).Methods("GET")

    http.Handle("/", r)

    appengine.Main()
}

type tplRenderer struct {
  tplName string
}

func (tr *tplRenderer) renderView(w http.ResponseWriter, r *http.Request) {
  filename := fmt.Sprintf("%s.tmpl", tr.tplName)
  filepath := fmt.Sprintf("static/tpl/%s", filename)

  t := template.Must(template.New(filename).ParseFiles(filepath))
  err := t.Execute(w, nil)
  if err != nil {
    log.Println(err)
    fmt.Println(err)
  }
}

func emailHandler(w http.ResponseWriter, r *http.Request) {
  ctx := appengine.NewContext(r)
  err := mail.SendMail("fiona@witches.nyc", "hey fiona", "<strong>whats good</strong>", ctx)
  if err != nil {
    fmt.Fprintln(w, err)
    return
  }

  fmt.Fprintln(w, "Sent mail")
}

func newBookHandler(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()

  bookId := r.Form["bookId"][0]
  delimiter := r.Form["delim"][0]

  ctx := appengine.NewContext(r)
  err := books.ChapterizeBook(bookId, delimiter, ctx)
  if err != nil {
    fmt.Fprintln(w, err)
    log.Println(err)
    return
  }

  fmt.Fprintf(w, "New book with ID %s and delimiter %s", bookId, delimiter)
}
