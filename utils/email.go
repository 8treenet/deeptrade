package utils

import (
	"gopkg.in/gomail.v2"
)

func SendHtmlMail(subject, body string) error {
	host := "smtp.exmail.qq.com"
	port := 465
	user := "8tree@8tree.net"
	pw := "Bachashu7944"

	msg := gomail.NewMessage()
	msg.SetHeader("From", "investment"+"<"+user+">")
	msg.SetHeader("To", "4932004@qq.com", "136753545@qq.com")
	// if len(conf.Get().System.ToMailList) > 0 {
	// 	msg.SetHeader("Cc", conf.Get().System.ToMailList...)
	// }
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)
	return gomail.NewDialer(host, port, user, pw).DialAndSend(msg)
}
