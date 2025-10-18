package secondary

import "bytes"

type SMTPClient interface {
	Send(to string, body, message string, subject string, file *bytes.Buffer) error
	GenerateEmailConfirmationMessage(filename string, data map[string]string) (string, error)
}
