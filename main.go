package main

import (
    "fmt"
		"log"
    "net/http"

    "mail"

    "google.golang.org/appengine"
)

func main() {
    http.HandleFunc("/", rootHandler)
    http.HandleFunc("/send", emailHandler)
    appengine.Main()
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.Redirect(w, r, "/", http.StatusFound)
        return
    }

    fmt.Fprintln(w, "Hello, bigf!")
    log.Println("Hello, world!")
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
