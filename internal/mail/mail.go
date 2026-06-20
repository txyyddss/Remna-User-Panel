// Package mail provides SMTP email delivery for verification codes,
// password resets, magic links, and notification emails.
package mail

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// Config holds SMTP server connection parameters.
type Config struct {
	Host       string
	Port       int
	Username   string
	Password   string
	FromName   string
	FromEmail  string
	Encryption string // "none", "tls", "starttls"
}

// IsConfigured returns true when the minimum required fields are set.
func (c Config) IsConfigured() bool {
	return c.Host != "" && c.Port > 0 && c.FromEmail != ""
}

// Addr returns the host:port address.
func (c Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Mailer sends emails via SMTP.
type Mailer struct {
	config Config
}

// NewMailer creates a Mailer from the given config.
func NewMailer(config Config) *Mailer {
	return &Mailer{config: config}
}

// IsConfigured returns true when the minimum required fields are set.
func (m *Mailer) IsConfigured() bool {
	return m.config.IsConfigured()
}

// Message represents an email to send.
type Message struct {
	To          []string
	Subject     string
	BodyHTML    string
	BodyPlain   string
	ContentType string // defaults to "text/html; charset=UTF-8"
}

// Send delivers an email message.
func (m *Mailer) Send(msg Message) error {
	if !m.config.IsConfigured() {
		return fmt.Errorf("mailer is not configured")
	}
	if len(msg.To) == 0 {
		return fmt.Errorf("no recipients")
	}

	contentType := msg.ContentType
	if contentType == "" {
		contentType = "text/html; charset=UTF-8"
	}

	headers := make(map[string]string)
	headers["From"] = m.formatFrom()
	headers["To"] = strings.Join(msg.To, ", ")
	headers["Subject"] = msg.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = contentType
	headers["Date"] = time.Now().Format(time.RFC1123Z)

	body := msg.BodyHTML
	if body == "" {
		body = msg.BodyPlain
		headers["Content-Type"] = "text/plain; charset=UTF-8"
	}

	var message strings.Builder
	for k, v := range headers {
		fmt.Fprintf(&message, "%s: %s\r\n", k, v)
	}
	message.WriteString("\r\n")
	message.WriteString(body)

	return m.sendMail(msg.To, message.String())
}

func (m *Mailer) formatFrom() string {
	if m.config.FromName != "" {
		return fmt.Sprintf("%s <%s>", m.config.FromName, m.config.FromEmail)
	}
	return m.config.FromEmail
}

func (m *Mailer) sendMail(to []string, msg string) error {
	addr := m.config.Addr()
	enc := strings.ToLower(strings.TrimSpace(m.config.Encryption))

	switch enc {
	case "tls":
		return m.sendWithTLS(addr, to, msg)
	case "starttls":
		return m.sendWithSTARTTLS(addr, to, msg)
	default:
		return m.sendPlain(addr, to, msg)
	}
}

func (m *Mailer) sendWithTLS(addr string, to []string, msg string) error {
	tlsConfig := &tls.Config{
		ServerName: m.config.Host,
		MinVersion: tls.VersionTLS12,
	}
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer func() { _ = conn.Close() }()

	client, err := smtp.NewClient(conn, m.config.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer func() { _ = client.Quit() }()

	return m.authAndSend(client, to, msg)
}

func (m *Mailer) sendWithSTARTTLS(addr string, to []string, msg string) error {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer func() { _ = conn.Close() }()

	client, err := smtp.NewClient(conn, m.config.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer func() { _ = client.Quit() }()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName: m.config.Host,
			MinVersion: tls.VersionTLS12,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	return m.authAndSend(client, to, msg)
}

func (m *Mailer) sendPlain(addr string, to []string, msg string) error {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer func() { _ = conn.Close() }()

	client, err := smtp.NewClient(conn, m.config.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer func() { _ = client.Quit() }()

	return m.authAndSend(client, to, msg)
}

func (m *Mailer) authAndSend(client *smtp.Client, to []string, msg string) error {
	if m.config.Username != "" || m.config.Password != "" {
		auth := smtp.PlainAuth("", m.config.Username, m.config.Password, m.config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("auth: %w", err)
		}
	}

	if err := client.Mail(m.config.FromEmail); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("rcpt to %s: %w", recipient, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return w.Close()
}
