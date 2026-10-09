package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strings"

	"github.com/wzantopulos/snipe-po/config"
)

func SendEmail(to []string, subject, body, attachmentPath string) error {
	return SendEmailHTML(to, subject, body, "", attachmentPath)
}

func SendEmailHTML(to []string, subject, htmlBody, textBody, attachmentPath string) error {
	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("config not loaded")
	}

	from := cfg.SMTP.From
	host := cfg.SMTP.Host
	port := cfg.SMTP.Port

	addr := fmt.Sprintf("%s:%d", host, port)

	var auth smtp.Auth
	if cfg.SMTP.Username != "" && cfg.SMTP.Password != "" {
		auth = smtp.PlainAuth("", cfg.SMTP.Username, cfg.SMTP.Password, host)
	}

	msg := buildHTMLMessage(from, to, subject, htmlBody, textBody, attachmentPath)

	// Port 465 uses implicit TLS (direct TLS connection)
	// Port 587 uses explicit TLS (STARTTLS after connection)
	// Port 25 uses no encryption (plaintext)
	if port == 465 {
		return sendWithImplicitTLS(addr, host, from, to, auth, msg)
	} else if port == 587 {
		return sendWithSTARTTLS(addr, host, from, to, auth, msg)
	}
	return sendWithPlaintext(addr, host, from, to, auth, msg)
}

// sendWithImplicitTLS connects directly with TLS (port 465)
func sendWithImplicitTLS(addr, host, from string, to []string, auth smtp.Auth, msg string) error {
	tlsConfig := &tls.Config{
		ServerName: host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server with TLS: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}
	}

	return sendWithClient(client, from, to, msg)
}

// sendWithSTARTTLS connects and upgrades with STARTTLS (port 587)
func sendWithPlaintext(addr, host, from string, to []string, auth smtp.Auth, msg string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}
	}

	return sendWithClient(client, from, to, msg)
}

func sendWithSTARTTLS(addr, host, from string, to []string, auth smtp.Auth, msg string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Send EHLO
	if err := client.Hello("localhost"); err != nil {
		return fmt.Errorf("SMTP EHLO failed: %w", err)
	}

	// Start TLS
	tlsConfig := &tls.Config{
		ServerName: host,
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("SMTP STARTTLS failed: %w", err)
	}

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}
	}

	return sendWithClient(client, from, to, msg)
}

func sendWithClient(client *smtp.Client, from string, to []string, msg string) error {
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL FROM failed: %w", err)
	}

	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("SMTP RCPT TO failed: %w", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA failed: %w", err)
	}
	defer w.Close()

	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	return nil
}

func buildMessage(from string, to []string, subject, body, attachmentPath string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("From: %s\r\n", from))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	sb.WriteString("MIME-Version: 1.0\r\n")

	if attachmentPath != "" {
		sb.WriteString("Content-Type: multipart/mixed; boundary=\"boundary\"\r\n\r\n")
		sb.WriteString("--boundary\r\n")
	}

	sb.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(body)
	sb.WriteString("\r\n")

	if attachmentPath != "" {
		filename := getFilename(attachmentPath)
		pdfData, err := os.ReadFile(attachmentPath)
		if err == nil {
			sb.WriteString("\r\n--boundary\r\n")
			sb.WriteString(fmt.Sprintf("Content-Type: application/pdf; name=\"%s\"\r\n", filename))
			sb.WriteString("Content-Transfer-Encoding: base64\r\n")
			sb.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", filename))
			sb.WriteString(encodeBase64(pdfData))
			sb.WriteString("\r\n--boundary--\r\n")
		}
	}

	return sb.String()
}

func buildHTMLMessage(from string, to []string, subject, htmlBody, textBody, attachmentPath string) string {
	var sb strings.Builder
	mimeBoundary := "boundary"

	sb.WriteString(fmt.Sprintf("From: %s\r\n", from))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", mimeBoundary))

	// Plain text part
	sb.WriteString("\r\n--" + mimeBoundary + "\r\n")
	sb.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	sb.WriteString("\r\n")
	if textBody != "" {
		sb.WriteString(textBody)
	} else {
		sb.WriteString(htmlBody) // fallback
	}
	sb.WriteString("\r\n")

	// HTML part
	sb.WriteString("\r\n--" + mimeBoundary + "\r\n")
	sb.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(htmlBody)
	sb.WriteString("\r\n")

	// Attachment part
	if attachmentPath != "" {
		sb.WriteString("\r\n--" + mimeBoundary + "\r\n")
		filename := getFilename(attachmentPath)
		pdfData, err := os.ReadFile(attachmentPath)
		if err == nil {
			sb.WriteString(fmt.Sprintf("Content-Type: application/pdf; name=\"%s\"\r\n", filename))
			sb.WriteString("Content-Transfer-Encoding: base64\r\n")
			sb.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", filename))
			sb.WriteString(encodeBase64(pdfData))
			sb.WriteString("\r\n")
		}
	}

	sb.WriteString("--" + mimeBoundary + "--\r\n")
	return sb.String()
}

func getFilename(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return path
}

func encodeBase64(data []byte) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var sb strings.Builder
	sb.Grow(len(data)*2)

	for i := 0; i < len(data); i += 3 {
		var n uint32
		switch len(data) - i {
		case 1:
			n = uint32(data[i]) << 16
		case 2:
			n = uint32(data[i])<<16 | uint32(data[i+1])<<8
		default:
			n = uint32(data[i])<<16 | uint32(data[i+1])<<8 | uint32(data[i+2])
		}

		sb.WriteByte(chars[(n>>18)&0x3F])
		sb.WriteByte(chars[(n>>12)&0x3F])
		if len(data)-i > 1 {
			sb.WriteByte(chars[(n>>6)&0x3F])
		}
		if len(data)-i > 2 {
			sb.WriteByte(chars[n&0x3F])
		}
	}

	padding := (3 - len(data)%3) % 3
	for i := 0; i < padding; i++ {
		sb.WriteByte('=')
	}

	return sb.String()
}
