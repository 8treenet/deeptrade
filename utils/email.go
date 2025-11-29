package utils

import (
	"gopkg.in/gomail.v2"
)

func SendHtmlMail(subject, body string) error {
	host := "smtp.exmail.qq.com"
	port := 465
	user := ""
	pw := ""

	msg := gomail.NewMessage()
	msg.SetHeader("From", "investment"+"<"+user+">")
	msg.SetHeader("To", "")
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)
	return gomail.NewDialer(host, port, user, pw).DialAndSend(msg)
}
