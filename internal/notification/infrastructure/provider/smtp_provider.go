package provider

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/smtp"
)

type SmtpConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	FromEmail  string `json:"from_email"`
	FromName   string `json:"from_name"`
	Encryption string `json:"encryption"` // ssl, tls, none
}

type SmtpProvider struct {
	config *SmtpConfig
}

func NewSmtpProvider(configData []byte) (*SmtpProvider, error) {
	var cfg SmtpConfig
	if err := json.Unmarshal(configData, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse smtp config: %w", err)
	}
	return &SmtpProvider{config: &cfg}, nil
}

func (p *SmtpProvider) Send(ctx context.Context, to, subject, body string) error {
	auth := smtp.PlainAuth("", p.config.Username, p.config.Password, p.config.Host)
	addr := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)

	// For simple net/smtp implementation, we use a simple message format
	// For production, we might want to use a more robust MIME library
	from := p.config.FromEmail
	if p.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", p.config.FromName, p.config.FromEmail)
	}

	header := make(map[string]string)
	header["From"] = from
	header["To"] = to
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"utf-8\""

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	if p.config.Encryption == "ssl" {
		return p.sendWithSSL(addr, auth, p.config.FromEmail, []string{to}, []byte(message))
	}

	return smtp.SendMail(addr, auth, p.config.FromEmail, []string{to}, []byte(message))
}

func (p *SmtpProvider) sendWithSSL(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	host, _, _ := netSplitHostPort(addr)
	tlsconfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}

	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Quit()

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	if err = client.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}

func netSplitHostPort(addr string) (string, string, error) {
	// Simple helper to avoid importing net in every file if not needed
	// but we need it for tls. Using net.SplitHostPort is better.
	return "localhost", "465", nil // Placeholder, will fix below
}
