package mailer

import (
	"bytes"
	"errors"
	"text/template"

	gomail "gopkg.in/mail.v2"
)

type mailTrapClient struct {
	fromEmail string
	apiKey    string
}

func NewMailTrapClient(apiKey, fromEmail string) (mailTrapClient, error) {
	if apiKey == "" {
		return mailTrapClient{}, errors.New("api key is required")
	}

	return mailTrapClient{
		fromEmail: fromEmail,
		apiKey:    apiKey,
	}, nil
}

func (m mailTrapClient) Send(templateFile, username, email string, data any) (int, error) {
	tmpl, err := template.ParseFS(FS, "templates/"+UserWelcomeTemplate)
	if err != nil {
		return -1, err
	}
	subject := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(subject, "subject", data); err != nil {
		return -1, err
	}
	body := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(body, "body", data); err != nil {
		return -1, err
	}

	message := gomail.NewMessage()
	message.SetHeader("From", m.fromEmail)
	message.SetHeader("To", "email")
	message.SetHeader("Subject", subject.String())

	// Set email body
	message.AddAlternative("text/html", body.String())

	// Set up the SMTP dialer
	dialer := gomail.NewDialer("live.smtp.mailtrap.io", 587, "api", m.apiKey)

	// Send the email
	if err := dialer.DialAndSend(message); err != nil {
		return -1, err
	}
	return 200, nil

}
