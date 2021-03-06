// using SendGrid's Go Library
// https://github.com/sendgrid/sendgrid-go
package main

import (
	"log"

	"config"

	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
	"gopkg.in/sendgrid/sendgrid-go.v2"
)

func SendMail(to, subject, html, plaintext string, ctx context.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	sg := sendgrid.NewSendGridClientWithApiKey(cfg.Email.SendgridApiKey)
	sg.Client = urlfetch.Client(ctx)

	log.Printf("Sending email to %s with subject %s", to, subject)

	message := sendgrid.NewMail()
	message.AddTo(to)
	message.SetFrom(cfg.Email.FromEmail)
	message.SetSubject(subject)
	message.SetHTML(html)
	message.SetText(plaintext)

	return sg.Send(message)
}
