package repository

import (
	"context"

	"github.com/an3wers/notification-serv/internal/domain/entity"
	"github.com/google/uuid"
)

type EmailRepository interface {
	Create(ctx context.Context, email *entity.Email) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Email, error)
	Update(ctx context.Context, email *entity.Email) error
	CreateAttachment(ctx context.Context, attachment *entity.Attachment) error
	FindAttachmentsByEmailID(ctx context.Context, emailID uuid.UUID) ([]entity.Attachment, error)
}
