package sender

import (
	"encoding/json"
	"fmt"
	"net/smtp"

	"internal/repository"
)

type emailConfig struct {
	auth     smtp.Auth
	host     string
	port     string
	from     string
	password string
}

func SendNotification(ch <-chan []byte) {
	cfg := startSMTP()

	go func() {
		for msg := range ch {
			notify := &repository.Notify{}

			err := json.Unmarshal(msg, notify)
			if err != nil {
				fmt.Println(err)
			}

			if notify.Email != "" {
				err = sendEmail(notify, cfg)
				if err != nil {
					fmt.Println(err)
				}
			}

			if notify.TGUser != "" {
				err = sendTG(notify)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}()
}
