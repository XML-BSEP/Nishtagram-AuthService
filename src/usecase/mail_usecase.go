package usecase

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
)

type MailForActivation struct {
	Email	string `json:"email"`
	Token	string `json:"token"`
	Expires	string `json::"expires"`
}

const (
	emailHost = "smtp.gmail.com"
	emailFrom = "duke.strategic@gmail.com"
	emailPassword = "Duke123456"
	emailPort = "587"
	adr = emailHost + ":" + emailPort
)

func SendMail(subjectMail string, subjectName string, verCode string) error {

	emailAuth := smtp.PlainAuth("", emailFrom, emailPassword, emailHost)

	t, _  := template.ParseFiles("assets/mail_template/template.html")
	var body bytes.Buffer
	headers := "MIME-version: 1.0;\nContent-Type: text/html;"
	body.Write([]byte(fmt.Sprintf("Subject: Email verification by Duke Strategic Techologies\n%s\n\n", headers)))

	t.Execute(&body, struct {
		UserName string
		Code string
	}{
		UserName: subjectName,
		Code: verCode,
	})

	to := []string{subjectMail}


	if err := smtp.SendMail(adr, emailAuth, emailFrom, to, body.Bytes()); err != nil {
		return  err
	}



	return nil
}

func SendRestartPasswordMail(subjectMail string, verCode string) error {

	emailAuth := smtp.PlainAuth("", emailFrom, emailPassword, emailHost)

	t, _  := template.ParseFiles("assets/mail_template/template_reset.html")
	var body bytes.Buffer
	headers := "MIME-version: 1.0;\nContent-Type: text/html;"
	body.Write([]byte(fmt.Sprintf("Subject: Code for password reset by Duke Strategic Techologies\n%s\n\n", headers)))

	t.Execute(&body, struct {
		Code string
	}{
		Code: verCode,
	})

	to := []string{subjectMail}


	if err := smtp.SendMail(adr, emailAuth, emailFrom, to, body.Bytes()); err != nil {
		return  err
	}



	return nil
}




func ParseTemplate(templateFileName string, data interface{}) (string, error) {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return "", err
	}
	body := buf.String()
	return body, nil
}
