package notifiers

import (
	"fmt"
	"net/smtp"

	"notification_system/config"
)

type GmailNotifier struct {
	From string
}

func (notifier *GmailNotifier) Notify(to, message string) error {
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	auth := smtp.PlainAuth("", notifier.From, config.Cfg.GmailAppPassword, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, notifier.From, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("notifiers.gmail error: %w", err)
	}
	return nil
}
