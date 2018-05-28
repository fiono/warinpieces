package main

import (
	"log"
	"net/http"

	"cloud.google.com/go/errorreporting"
	"google.golang.org/appengine"
)

func reportAndReturnInternalError(w http.ResponseWriter, r *http.Request, err error) {
	reportError(r, err)
	http.Error(w, "Sorry! Something unexpected happened. I'll take a look soon.", 500)
}

func returnClientError(w http.ResponseWriter, msg string) {
	http.Error(w, msg, 409)
}

func getErrorReportingClient(r *http.Request, projectId string) (client *errorreporting.Client, err error) {
	ctx := appengine.NewContext(r)
	return errorreporting.NewClient(ctx, projectId, errorreporting.Config{
		ServiceName: "gutenbits",
		OnError: func(err error) {
			panic(err)
		},
	})
}

func reportError(r *http.Request, err error) {
	errorClient, e := getErrorReportingClient(r, "gutenbits")
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
