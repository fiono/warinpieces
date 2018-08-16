package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strings"

	"books"
	"config"
	"views"

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

	r.HandleFunc("/books/", (&views.TplRenderer{Tpl: "book", IsWeb: true}).ServeView).Methods("GET")
	r.HandleFunc("/books/new/", newBookHandler).Methods("POST")

	r.HandleFunc("/", newSubscriptionView).Methods("GET")
	r.HandleFunc("/subscriptions/new/", newSubscriptionHandler).Methods("POST")
	r.HandleFunc("/subscriptions/unsubscribe/", singleUnsubscribeHandler).Methods("GET")
	r.HandleFunc("/subscriptions/confirm/", validateSubscriptionHandler).Methods("GET")

	http.Handle("/", r)
	appengine.Main()
}

/*
  Cron logic
*/

type sendEmailResponse struct {
	subId string
	err   error
}

func cronHandler(w http.ResponseWriter, r *http.Request) {
	db, err := dbConn()
	if err != nil {
		reportError(r, err)
		return
	}
	defer db.Close()

	subs, err := db.GetSubscriptionsForSending()
	if err != nil {
		reportError(r, err)
		return
	}

	ctx := appengine.NewContext(r)
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
			reportError(r, err)
		}
	}

	log.Printf("Done, %d successes out of %d tries", successes, tries)
	fmt.Fprintf(w, "Done, %d successes out of %d tries", successes, tries)
}

func sendEmailForSubscription(subscriptionId string, ctx context.Context, ch chan sendEmailResponse) {
	ch <- sendEmailResponse{subscriptionId, sendEmailForSubscriptionSingle(subscriptionId, ctx)}
}

func sendEmailForSubscriptionSingle(subscriptionId string, ctx context.Context) error {
	db, err := dbConn()
	if err != nil {
		return err
	}
	defer db.Close()

	sub, err := db.GetSubscription(subscriptionId)
	if err != nil {
		return err
	}

	book, err := db.GetBook(sub.BookId)
	if err != nil {
		return err
	}

	body, err := books.GetChapter(sub.BookId, sub.ChaptersSent+1, ctx)
	if err != nil {
		return err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	token, err := getSubscriptionToken(sub.BookId, sub.Email)
	if err != nil {
		return err
	}

	content := strings.Replace(body, "\n\n", "<br/><br/>", -1)
	emailBody, err := views.EmailRenderer(token, content, cfg.Main.UrlBase, book, sub).GetView()
	if err != nil {
		return err
	}

	if err = SendMail(sub.Email, book.Title, emailBody, body, ctx); err != nil {
		return err
	}

	if sub.ChaptersSent+1 == book.Chapters {
		db.IncrementChaptersSent(sub.SubscriptionId)
		return db.DeactivateSingle(sub.Email, sub.BookId)
	}
	return db.IncrementChaptersSent(sub.SubscriptionId)
}

/*
  Endpoints
*/
func newSubscriptionView(w http.ResponseWriter, r *http.Request) {
	db, err := dbConn()
	if err != nil {
		reportAndReturnInternalError(w, r, err)
		return
	}
	defer db.Close()

	validBooks, err := db.GetBooks()
	if err != nil {
		reportError(r, err)
		return
	}

	views.NewSubscriptionRenderer(validBooks).ServeView(w, r)
}

func newBookHandler(w http.ResponseWriter, r *http.Request) {
	db, err := dbConn()
	if err != nil {
		reportAndReturnInternalError(w, r, err)
		return
	}
	defer db.Close()

	r.ParseForm()
	bookId := r.Form["bookId"][0]
	delimiter := r.Form["delim"][0]

	ctx := appengine.NewContext(r)
	meta, err := books.ChapterizeBook(bookId, delimiter, ctx)
	if err != nil {
		reportAndReturnInternalError(w, r, err)
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
		reportAndReturnInternalError(w, r, err)
		return
	}

	fmt.Fprintf(w, "New book: ", meta)
}

func newSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	bookId := r.Form["bookId"][0]
	parsedAddr, err := mail.ParseAddress(r.Form["email"][0])
	if err != nil {
		returnClientError(w, "Invalid email address!")
		return
	}
	emailAddress := parsedAddr.Address

	db, err := dbConn()
	if err != nil {
		reportAndReturnInternalError(w, r, err)
		return
	}
	defer db.Close()

	if err := db.NewSubscription(bookId, emailAddress); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			returnClientError(w, "Already have a subscription for that book & email address!")
		} else {
			reportAndReturnInternalError(w, r, err)
		}
		return
	}

	book, err := db.GetBook(bookId)
	if err != nil {
		reportAndReturnInternalError(w, r, err)
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return
	}

	token, err := getSubscriptionToken(book.BookId, emailAddress)
	if err != nil {
		return
	}

	emailBody, err := views.ConfirmEmailRenderer(emailAddress, token, cfg.Main.UrlBase, book).GetView()
	if err != nil {
		reportAndReturnInternalError(w, r, err)
		return
	}

	ctx := appengine.NewContext(r)
	if err = SendMail(emailAddress, book.Title, emailBody, "Success!", ctx); err != nil {
		reportAndReturnInternalError(w, r, err)
		return
	}

	views.SubscriptionSuccessRenderer(book, emailAddress).ServeView(w, r)
}

func singleUnsubscribeHandler(w http.ResponseWriter, r *http.Request) {
	db, err := dbConn()
	if err != nil {
		reportAndReturnInternalError(w, r, err)
		return
	}
	defer db.Close()

	token := r.URL.Query().Get("token")
	emailAddress := r.URL.Query().Get("email_address")
	bookId := r.URL.Query().Get("book_id")

	cmpToken, err := getSubscriptionToken(bookId, emailAddress)
	if err != nil {
		return
	}

	if token == cmpToken {
		if err != nil {
			reportAndReturnInternalError(w, r, err)
			return
		}

		err := db.DeactivateSingle(emailAddress, bookId)
		if err != nil {
			reportAndReturnInternalError(w, r, err)
			return
		}

		book, err := db.GetBook(bookId)
		if err != nil {
			reportAndReturnInternalError(w, r, err)
			return
		}

		views.UnsubscriptionSuccessRenderer(emailAddress, book).ServeView(w, r)
	} else {
		reportError(r, errors.New(fmt.Sprintf("Unsubscribe fail for user %s", emailAddress)))
		returnClientError(w, "Token mismatch--failed to unsubscribe.")
	}
}

func validateSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	db, err := dbConn()
	if err != nil {
		reportAndReturnInternalError(w, r, err)
		return
	}
	defer db.Close()

	token := r.URL.Query().Get("token")
	emailAddress := r.URL.Query().Get("email_address")
	bookId := r.URL.Query().Get("book_id")

	cmpToken, err := getSubscriptionToken(bookId, emailAddress)
	if err != nil {
		return
	}

	if token == cmpToken { // this should actually be time-sensitive ¯\_(ツ)_/¯
		sub, err := db.GetSubscriptionByData(bookId, emailAddress)
		if err != nil {
			reportAndReturnInternalError(w, r, err)
			return
		}

		book, err := db.GetBook(bookId)
		if err != nil {
			reportAndReturnInternalError(w, r, err)
			return
		}

		if !sub.Validated {
			if err = db.ActivateSubscription(bookId, emailAddress); err != nil {
				reportAndReturnInternalError(w, r, err)
				return
			}

			ctx := appengine.NewContext(r)
			if err := sendEmailForSubscriptionSingle(sub.SubscriptionId, ctx); err != nil {
				reportAndReturnInternalError(w, r, err)
				return
			}
		}

		views.ConfirmationSuccessRenderer(emailAddress, book).ServeView(w, r)
	} else {
		returnClientError(w, "Token mismatch--failed to activate your subscription")
	}
}
