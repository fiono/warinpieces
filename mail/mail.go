// using SendGrid's Go Library
// https://github.com/sendgrid/sendgrid-go
package mail

import (
  "log"

  "config"

  "golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
  "gopkg.in/sendgrid/sendgrid-go.v2"
)

func SendMail(to string, subject string, body string, ctx context.Context) error {
  cfg := config.LoadConfig()

  sg := sendgrid.NewSendGridClientWithApiKey(cfg.Email.SendgridApiKey)
  sg.Client = urlfetch.Client(ctx)

  log.Printf("Sending email to %s with subject %s", to, subject)

  message := sendgrid.NewMail()
  message.AddTo(to)
  message.SetFrom("fiona@witches.nyc")
  message.SetSubject(subject)
  message.SetHTML(body)

  err := sg.Send(message)
  return err
}
