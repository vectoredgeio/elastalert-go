package alerts

import (
	"fmt"
	"net/smtp"
)

type EmailAlert struct {
	Recipients []string
	SMTPServer string
	SMTPPort   int
	Username   string
	Password   string
}

func (e *EmailAlert) Send(message string) error {
	auth := smtp.PlainAuth("", e.Username, e.Password, e.SMTPServer)
	msg := []byte("To: " + e.Recipients[0] + "\r\n" +
		"Subject: ElastAlert Notification\r\n" +
		"\r\n" +
		message + "\r\n")
	err := smtp.SendMail(fmt.Sprintf("%s:%d", e.SMTPServer, e.SMTPPort), auth, e.Username, e.Recipients, msg)
	if err != nil {
		return err
	}
	return nil
}
