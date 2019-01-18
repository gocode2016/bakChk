package server

import (
	"net/smtp"
	"registerServer/common"
	"strings"
	"fmt"
)


func SendEmail(body string) {
	auth := smtp.PlainAuth("", G_config.username, G_config.password, G_config.smtpServer)
	to := []string{G_config.receiver}
	sender := G_config.sender
	user := G_config.username
	subject := common.EMAIL_SUBJECT
	contentType := "Content-Type: text/plain; charset=UTF-8"
	msg := []byte("To: " + strings.Join(to, ",") + "\r\nFrom: " + sender +
		"<" + user + ">\r\nSubject: " + subject + "\r\n" + contentType + "\r\n\r\n" + body)
	emailServer := G_config.smtpServer + ":25"
	err := smtp.SendMail(emailServer, auth, user, to, msg)
	if err != nil {
		fmt.Println("Email send failed,", err)
	}
}