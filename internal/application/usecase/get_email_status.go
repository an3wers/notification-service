package usecase

import (
	"context"

	"github.com/an3wers/notification-serv/internal/domain/entity"
	"github.com/an3wers/notification-serv/internal/domain/repository"
	"github.com/google/uuid"
)

type GetEmailStatusUseCase struct {
	emailRepo repository.EmailRepository
}

func NewGetEmailStatusUseCase(emailRepo repository.EmailRepository) *GetEmailStatusUseCase {
	return &GetEmailStatusUseCase{
		emailRepo: emailRepo,
	}
}

func (uc *GetEmailStatusUseCase) Execute(ctx context.Context, emailID uuid.UUID) (*entity.Email, error) {
	return uc.emailRepo.FindByID(ctx, emailID)
}
