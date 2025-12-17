package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/an3wers/notification-serv/internal/domain/entity"
	"github.com/an3wers/notification-serv/internal/domain/service"
	"github.com/an3wers/notification-serv/internal/pkg/config"
	"gopkg.in/gomail.v2"
)

type smtpProvider struct {
	cfg    config.SMTPConfig
	dialer *gomail.Dialer
}

func NewSMTPProvider(cfg config.SMTPConfig) service.EmailProvider {
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)

	if cfg.TLS {
		dialer.TLSConfig = &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         cfg.Host,
			MinVersion:         tls.VersionTLS12,
		}

	} else {
		dialer.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	return &smtpProvider{
		cfg:    cfg,
		dialer: dialer,
	}
}

func (p *smtpProvider) Send(ctx context.Context, email *entity.Email) (*service.SendEmailResult, error) {
	m := gomail.NewMessage()

	m.SetAddressHeader("From", email.From, p.cfg.FromDisplayName)
	m.SetHeader("To", email.To...)

	if len(email.CC) > 0 {
		m.SetHeader("Cc", email.CC...)
	}

	if len(email.BCC) > 0 {
		m.SetHeader("Bcc", email.BCC...)
	}

	m.SetHeader("Subject", email.Subject)
	m.SetBody("text/plain", email.Body)

	if email.HTML != nil {
		m.AddAlternative("text/html", *email.HTML)
	}

	// Attach files
	for _, att := range email.Attachments {
		m.Attach(att.Path, gomail.Rename(att.OriginalName)) // att.Path - ./uploads/...
	}

	// Create a channel for timeout
	done := make(chan error, 1)

	go func() {
		done <- p.dialer.DialAndSend(m)
	}()

	// Wait with timeout
	timeout := time.Duration(p.cfg.Timeout) * time.Second

	select {
	case err := <-done:
		if err != nil {
			return &service.SendEmailResult{
				Success: false,
				Error:   err,
			}, nil
		}
		return &service.SendEmailResult{
			Success:   true,
			MessageID: email.ID.String(),
		}, nil
	case <-time.After(timeout):
		return &service.SendEmailResult{
			Success: false,
			Error:   fmt.Errorf("smtp timeout after %v", timeout),
		}, nil
	case <-ctx.Done():
		return &service.SendEmailResult{
			Success: false,
			Error:   ctx.Err(),
		}, nil
	}
}
