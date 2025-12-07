package service

import (
	"context"

	"github.com/an3wers/notification-serv/internal/domain/entity"
)

type SendEmailResult struct {
	Success   bool
	MessageID string
	Error     error
}

type EmailProvider interface {
	Send(ctx context.Context, email *entity.Email) (*SendEmailResult, error)
}
