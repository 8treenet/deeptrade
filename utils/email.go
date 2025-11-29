package utils

import (
	"gopkg.in/gomail.v2"
)

/*
#to_mail_list = ["4932004@qq.com","136753545@qq.com","115826097@qq.com","liuhuaqing@bachashu.com","597660243@qq.com","wjfmmy@126.com"]
to_mail_list = ["4932004@qq.com"]
#邮件发送者
from_mail = "8tree@8tree.net"
#邮件发送者密码
from_mail_password = "Bachashu7944"
#邮件服务器地址
mail_server = "smtp.exmail.qq.com"
#邮件服务器端口
mail_server_port = 465
*/

func SendHtmlMail(subject, body string) error {
	host := "smtp.exmail.qq.com"
	port := 465
	user := "8tree@8tree.net"
	pw := "Bachashu7944"

	msg := gomail.NewMessage()
	msg.SetHeader("From", "investment"+"<"+user+">")
	msg.SetHeader("To", "4932004@qq.com")
	// if len(conf.Get().System.ToMailList) > 0 {
	// 	msg.SetHeader("Cc", conf.Get().System.ToMailList...)
	// }
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)
	return gomail.NewDialer(host, port, user, pw).DialAndSend(msg)
}
