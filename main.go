package main

import (
    "fmt"
		"log"
    "net/http"

    "google.golang.org/appengine"
)

func main() {
    http.HandleFunc("/", handler)
    appengine.Main()
}

func handler(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.Redirect(w, r, "/", http.StatusFound)
        return
    }

    fmt.Fprintln(w, "Hello, bigf!")
    log.Println("Hello, world!")
}
