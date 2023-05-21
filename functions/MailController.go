package functions

import (
	"strings"

	"gopkg.in/gomail.v2"
)

func SendMail(to []string, subject string, body string) error {
	if len(to) == 0 {
		return nil
	}
	// Set up authentication information.
	from := "monitordeprocesos@outlook.com"
	password := "Monitor123!"
	smtpServer := "smtp-mail.outlook.com"
	smtpPort := 587
	// uses GoMail
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", removeDuplicates(to)...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	d := gomail.NewDialer(smtpServer, smtpPort, from, password)
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
