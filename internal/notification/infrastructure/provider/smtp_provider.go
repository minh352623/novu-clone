package provider

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/smtp"
	"strings"
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

	var builder strings.Builder
	for k, v := range header {
		builder.WriteString(k)
		builder.WriteString(": ")
		builder.WriteString(v)
		builder.WriteString("\r\n")
	}
	builder.WriteString("\r\n")
	builder.WriteString(body)
	message := builder.String()

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
		return fmt.Errorf("smtp ssl: failed to dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp ssl: failed to create client: %w", err)
	}
	defer client.Quit()

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("smtp ssl: auth failed: %w", err)
		}
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("smtp ssl: MAIL FROM failed: %w", err)
	}
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("smtp ssl: RCPT TO failed: %w", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp ssl: DATA command failed: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("smtp ssl: failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("smtp ssl: failed to close writer: %w", err)
	}

	return client.Quit()
}

func netSplitHostPort(addr string) (string, string, error) {
	return net.SplitHostPort(addr)
}
