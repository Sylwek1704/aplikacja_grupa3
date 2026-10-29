package main

import (
	"mime"
	"net"
	"net/smtp"
	"strings"
)

func wyslijMaila(do, temat, tresc string) error {
	serwer := env("SMTP_ADRES", "localhost:1025")
	od := env("SMTP_OD", "czat@localhost")

	wiadomosc := "From: " + od + "\r\n" +
		"To: " + do + "\r\n" +
		"Subject: " + mime.QEncoding.Encode("utf-8", temat) + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"Content-Transfer-Encoding: 8bit\r\n" +
		"\r\n" +
		strings.ReplaceAll(tresc, "\n", "\r\n")

	var auth smtp.Auth
	if login := env("SMTP_LOGIN", ""); login != "" {
		host, _, _ := net.SplitHostPort(serwer)
		auth = smtp.PlainAuth("", login, env("SMTP_HASLO", ""), host)
	}
	return smtp.SendMail(serwer, auth, od, []string{do}, []byte(wiadomosc))
}
