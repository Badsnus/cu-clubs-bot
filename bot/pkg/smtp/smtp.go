package smtp

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"gopkg.in/gomail.v2"

	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger"
)

// Client представляет почтовый клиент.
type Client struct {
	dialer *gomail.Dialer

	domain string
	from   string
}

// NewClient инициализирует Client.
func NewClient(dialer *gomail.Dialer, domain string, from string) *Client {
	return &Client{
		dialer: dialer,
		domain: domain,
		from:   from,
	}
}

// Send отправляет письмо.
func (c *Client) Send(to string, body, message string, subject string, file *bytes.Buffer) error {
	msg := gomail.NewMessage()

	msg.SetHeader("Message-ID", generateMessageID(c.domain))
	msg.SetHeader("Date", time.Now().Format(time.RFC1123Z))
	msg.SetHeader("From", c.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", body)
	msg.AddAlternative("text/html", message)

	if file != nil {
		msg.Attach("passes.xlsx", gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(file.Bytes())
			if err != nil {
				logger.Log.Error(err)
			}
			return err
		}))
	}

	if err := c.dialer.DialAndSend(msg); err != nil {
		logger.Log.Error(err)
		return err
	}

	logger.Log.Info("Email successfully sent")
	return nil
}

func generateMessageID(domain string) string {
	uniqueID := uuid.New().String()
	return fmt.Sprintf("<%s@%s>", uniqueID, domain)
}

// GenerateEmailConfirmationMessage загружает HTML-шаблон для отправки письма с подтверждением аккаунта и подставляет в него переменные.
func GenerateEmailConfirmationMessage(filename string, data map[string]string) (string, error) {
	templateBytes, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	templateContent := string(templateBytes)

	tmpl, err := template.New("email").Parse(templateContent)
	if err != nil {
		return "", err
	}

	var body bytes.Buffer
	err = tmpl.Execute(&body, data)
	if err != nil {
		return "", err
	}

	return body.String(), nil
}
