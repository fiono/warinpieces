// using SendGrid's Go Library
// https://github.com/sendgrid/sendgrid-go
package mail

import (
  "log"
  "os"

  "golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
  "gopkg.in/sendgrid/sendgrid-go.v2"
)

func SendMail(to string, subject string, body string, ctx context.Context) error {
  key := os.Getenv("sendgrid_api_key")
  sg := sendgrid.NewSendGridClientWithApiKey(key)
  sg.Client = urlfetch.Client(ctx)

  message := sendgrid.NewMail()
  message.AddTo(to)
  message.SetFrom("fiona@witches.nyc")
  message.SetSubject(subject)
  message.SetHTML(body)

  err := sg.Send(message)
  return err
}
