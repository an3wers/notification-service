package usecase

import (
	"context"
	"fmt"

	"github.com/an3wers/notification-serv/internal/application/dto"
	"github.com/an3wers/notification-serv/internal/domain/entity"
	"github.com/an3wers/notification-serv/internal/domain/repository"
	"github.com/an3wers/notification-serv/internal/domain/service"
	"github.com/an3wers/notification-serv/internal/pkg/config"
	"github.com/an3wers/notification-serv/internal/pkg/logger"
	"go.uber.org/zap"
)

type SendEmailUseCase struct {
	emailRepo     repository.EmailRepository
	emailProvider service.EmailProvider
	cfg           config.SMTPConfig
	logger        *logger.Logger
}

func NewSendEmailUseCase(
	emailRepo repository.EmailRepository,
	emailProvider service.EmailProvider,
	cfg config.SMTPConfig,
	logger *logger.Logger,
) *SendEmailUseCase {
	return &SendEmailUseCase{
		emailRepo:     emailRepo,
		emailProvider: emailProvider,
		cfg:           cfg,
		logger:        logger,
	}
}

func (uc *SendEmailUseCase) Execute(
	ctx context.Context,
	req *dto.SendEmailNormalizedRequest,
	attachments []dto.AttachmentDTO,
) (*entity.Email, error) {

	// Create email entity
	var subject string

	if req.Subject != nil {
		subject = *req.Subject
	} else {
		subject = ""
	}

	var body string
	if req.Body != nil {
		body = *req.Body
	} else {
		body = ""
	}

	// uc.cfg.From
	var from string
	if req.From != nil {
		from = *req.From
	} else {
		from = uc.cfg.From
	}

	var displayName string
	if req.DisplayName != nil {
		displayName = *req.DisplayName
	} else {
		displayName = uc.cfg.FromDisplayName
	}

	email := entity.NewEmail(from, req.To, displayName, subject, body)
	email.CC = req.CC
	email.BCC = req.BCC
	email.HTML = req.HTML

	// Add attachments
	for _, att := range attachments {
		attachment := entity.NewAttachment(
			email.ID,
			att.Filename,
			att.OriginalName,
			att.Mimetype,
			att.Size,
			att.Path,
			nil,
		)

		email.Attachments = append(email.Attachments, *attachment)
	}

	// Save to database
	if err := uc.emailRepo.Create(ctx, email); err != nil {
		uc.logger.Error("Failed to save email", zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to save email: %w", err)
	}

	uc.logger.Info("Email saved to database", zap.Any("email_id", email.ID))

	// Send email
	result, err := uc.emailProvider.Send(ctx, email)

	if err != nil {
		uc.logger.Error("Failed to send email", zap.String("error", err.Error()), zap.Any("email_id", email.ID))
		email.MarkAsFailed(err.Error())
		uc.emailRepo.Update(ctx, email)
		return email, err
	}

	if !result.Success {
		errMsg := "unknown error"

		if result.Error != nil {
			errMsg = result.Error.Error()
		}

		uc.logger.Error("Email send failed", zap.String("error", errMsg), zap.Any("email_id", email.ID))
		email.MarkAsFailed(errMsg)
		uc.emailRepo.Update(ctx, email)
		return email, fmt.Errorf("email send failed: %s", errMsg)
	}

	// Mark as sent
	email.MarkAsSent()
	if err := uc.emailRepo.Update(ctx, email); err != nil {
		uc.logger.Error("Failed to update email status", zap.String("error", err.Error()), zap.Any("email_id", email.ID))
		return email, err
	}

	uc.logger.Info("Email sent successfully", zap.Any("email_id", email.ID), zap.String("message_id", result.MessageID))
	return email, nil
}
