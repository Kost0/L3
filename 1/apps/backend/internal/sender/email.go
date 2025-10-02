package sender

import (
	"net/smtp"
	"os"

	"internal/repository"
)

func startSMTP() *emailConfig {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	from := os.Getenv("SMTP_FROM")
	password := os.Getenv("SMTP_PASSWORD")

	auth := smtp.PlainAuth("", from, password, host)

	return &emailConfig{
		auth:     auth,
		host:     host,
		port:     port,
		from:     from,
		password: password,
	}
}

func sendEmail(notify *repository.Notify, config *emailConfig) error {
	err := smtp.SendMail(config.host+":"+config.port, config.auth, config.from, []string{notify.Email}, []byte(notify.Text))
	if err != nil {
		return err
	}

	return nil
}
