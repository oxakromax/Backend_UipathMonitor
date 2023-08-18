package functions

import (
	"os"
	"strconv"
	"strings"

	"gopkg.in/gomail.v2"
)

func SendMail(to []string, subject string, body string) error {
	if len(to) == 0 {
		return nil
	}
	// Set up authentication information.
	from := os.Getenv("MAIL_ADRESS")
	password := os.Getenv("MAIL_PASSWORD")
	smtpServer := os.Getenv("MAIL_SMTP_SERVER")
	smtpPort := os.Getenv("MAIL_SMTP_PORT")
	smtpPortInt, err := strconv.Atoi(smtpPort)
	if err != nil {
		return err
	}
	// uses GoMail
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", removeDuplicates(to)...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	d := gomail.NewDialer(smtpServer, smtpPortInt, from, password)
	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

func removeDuplicates(to []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range to {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, strings.TrimSpace(entry))
		}
	}
	return list
}
