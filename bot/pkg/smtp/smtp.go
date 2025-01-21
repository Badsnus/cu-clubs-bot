package smtp

import (
	"fmt"
	"github.com/Badsnus/cu-clubs-bot/internal/adapters/logger"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
	"time"
)

// Client представляет почтовый клиент.
type Client struct {
	dialer *gomail.Dialer
}

// NewClient инициализирует Client.
func NewClient(dialer *gomail.Dialer) *Client {
	return &Client{dialer: dialer}
}

// SendConfirmationEmail отправляет письмо с подтверждением почты.
func (c *Client) SendConfirmationEmail(to string, code string) {
	fmt.Println(1)
	msg := gomail.NewMessage()

	domain := viper.GetString("bot.domain")
	messageID := generateMessageID(domain)

	msg.SetHeader("Message-ID", messageID)
	msg.SetHeader("Date", time.Now().Format(time.RFC1123Z))
	msg.SetHeader("From", viper.GetString("bot.smtp-email"))
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", "Email Confirmation")
	msg.SetBody("text/plain", fmt.Sprintf("Отправьте этот код боту %s", code))
	msg.AddAlternative("text/html", code)
	if err := c.dialer.DialAndSend(msg); err != nil {
		logger.Log.Error(err)
		return
	}

	logger.Log.Info("Email successfully sent")
}

func generateMessageID(domain string) string {
	uniqueID := uuid.New().String()
	return fmt.Sprintf("<%s@%s>", uniqueID, domain)
}
