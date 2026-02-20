package mailer

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"text/template"
	"time"

	gomail "gopkg.in/mail.v2"
)

type mailTrapClient struct {
	fromEmail string
	apiKey    string
}

func NewMailTrapClient(apiKey, email string) (*mailTrapClient, error) {
	if apiKey == "" {
		return &mailTrapClient{}, errors.New("api key is requaired")
	}

	return &mailTrapClient{
		apiKey:    apiKey,
		fromEmail: email,
	}, nil
}

func (m *mailTrapClient) Send(templateFile, username, email string,
	data any, isSandbox bool) (int, error) {
	if isSandbox {
		return -1, fmt.Errorf("isSandbox enabled. Do not send email")
	}
	// Template parsing and building
	tmpl, err := template.ParseFS(FS, "template/"+templateFile)
	if err != nil {
		return -1, err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return -1, err
	}

	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return -1, err
	}

	message := gomail.NewMessage()
	message.SetHeader("From", m.fromEmail)
	message.SetHeader("To", email)
	message.SetHeader("Subject", subject.String())

	message.AddAlternative("text/html", body.String())

	dialer := gomail.NewDialer("live.smtp.mailtrap.io", 587, "api", m.apiKey)
	for i := 0; i < maxSendingRetries; i++ {
		err := dialer.DialAndSend(message)
		if err != nil {
			log.Printf("Failed to send email, attempt %d of %d",
				i+1, maxSendingRetries)
			log.Println("Error ", err.Error())

			time.Sleep(time.Second * time.Duration(i+1))
		} else {
			return 200, nil
		}
	}

	return -1, fmt.Errorf("failed to send an email after %d attempts", maxSendingRetries)
}
