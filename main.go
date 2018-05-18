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
    r.HandleFunc("/", rootHandler)
    r.HandleFunc("/send", emailHandler) // BIGF

    r.HandleFunc("/books/new/", newBookHandler)

    //r.HandleFunc("/subscriptions/new/", newSubscriptionHandler).Methods("POST")
    //r.HandleFunc("/subscriptions/validate/{subscription_id}", validateSubscriptionHandler).Methods("GET")
    //r.HandleFunc("/subscriptions/deactivate/{subscription_id}", deactivateSubscriptionHandler).Methods("GET")
    //r.HandleFunc("/subscriptions/reactivate/{subscription_id}", reactivateSubscriptionHandler).Methods("GET")

    http.Handle("/", r)

    appengine.Main()
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.Redirect(w, r, "/", http.StatusFound)
        return
    }

    var t = template.Must(template.New("index.html").ParseFiles("static/html/index.html"))

    err := t.Execute(w, "nil")
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
  } else {
    fmt.Fprintln(w, "Sent mail")
  }
}

func newBookHandler(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  book_id := vars["book_id"]
  fmt.Fprintln(w, "New book with ID", book_id)

  ctx := appengine.NewContext(r)
  err := books.ChapterizeBook(book_id, ctx)
  if err != nil {
    fmt.Fprintln(w, err)
    log.Println(err)
  }
}
