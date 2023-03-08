package functions

import (
	"gopkg.in/gomail.v2"
	"strings"
)

func SendMail(to []string, subject string, body string) error {
	// Set up authentication information.
	from := "monitordeprocesos@outlook.com"
	password := "Monitor123!"
	smtpServer := "smtp-mail.outlook.com"
	smtpPort := 587
	// uses GoMail
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", strings.Join(removeDuplicates(to), ","))
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
			list = append(list, entry)
		}
	}
	return list
}
