package sender

import (
	"encoding/json"
	"net/smtp"
	"sync"

	"github.com/Kost0/L3/internal/repository"
	"github.com/wb-go/wbf/zlog"
)

type emailConfig struct {
	auth     smtp.Auth
	host     string
	port     string
	from     string
	password string
}

var DeletedMu sync.RWMutex

func isDeleted(id string) bool {
	DeletedMu.RLock()
	defer DeletedMu.RUnlock()
	_, exist := repository.Deleted[id]
	return exist
}

func SendNotification(ch <-chan []byte) {
	cfg := startSMTP()

	go func() {
		for msg := range ch {
			zlog.Logger.Info().Msg("Notification received from queue")
			notify := &repository.Notify{}

			err := json.Unmarshal(msg, notify)
			if err != nil {
				zlog.Logger.Log().Msg(err.Error())
				continue
			}

			if isDeleted(notify.ID) {
				zlog.Logger.Info().Str("message", notify.ID).Msg("notify deleted")
				continue
			}

			if notify.Email != "" {
				err = sendEmail(notify, cfg)
				if err != nil {
					zlog.Logger.Log().Msg(err.Error())
				}
				zlog.Logger.Info().Str("message", notify.ID).Msg("notify sent on email")
			}

			if notify.TGUser != "" {
				err = sendTG(notify)
				if err != nil {
					zlog.Logger.Log().Msg(err.Error())
				}
				zlog.Logger.Info().Str("message", notify.ID).Msg("notify sent on tg")
			}
		}
	}()
}
