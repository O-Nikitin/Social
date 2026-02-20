package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *sendgrid.Client
}

func NewSendGridMailer(apiKey, email string) *SendGridMailer {
	client := sendgrid.NewSendClient(apiKey)

	return &SendGridMailer{
		fromEmail: email,
		apiKey:    apiKey,
		client:    client,
	}
}

func (m *SendGridMailer) Send(templateFile, username, email string,
	data any, isSandbox bool) error {
	from := mail.NewEmail(FromName, m.fromEmail)
	to := mail.NewEmail(username, email)

	//template parsing and building
	tmpl, err := template.ParseFS(
		FS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}
	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return err
	}
	message := mail.NewSingleEmail(
		from, subject.String(), to, "", body.String())

	//Set test mode for emails
	message.SetMailSettings(&mail.MailSettings{
		SandboxMode: &mail.Setting{
			Enable: &isSandbox,
		}})

	for i := 0; i < maxSendingRetries; i++ {
		response, err := m.client.Send(message)
		if err != nil {
			log.Printf("Failed to send email to %s, attempt %d of %d",
				email, i+1, maxSendingRetries)
			log.Println("Error ", err.Error())

			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		log.Printf("Email sent with status code %v",
			response.StatusCode)
		return nil

	}

	return fmt.Errorf("failed to send email after %d attempts", maxSendingRetries)
}
