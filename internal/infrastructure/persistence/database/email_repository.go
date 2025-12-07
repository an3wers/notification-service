package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/an3wers/notification-serv/internal/domain/entity"
	"github.com/an3wers/notification-serv/internal/domain/repository"
	apperrors "github.com/an3wers/notification-serv/internal/pkg/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type emailRepository struct {
	db *DB
}

func NewEmailRepository(db *DB) repository.EmailRepository {
	return &emailRepository{db: db}
}

func (r *emailRepository) Create(ctx context.Context, email *entity.Email) error {
	query := `
		INSERT INTO emails (
			id, "from", "to", cc, bcc, subject, body, html,
			status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		email.ID,
		email.From,
		email.To,
		email.CC,
		email.BCC,
		email.Subject,
		email.Body,
		email.HTML,
		email.Status,
		email.CreatedAt,
		email.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create email: %w", err)
	}

	// Create attachments
	for _, att := range email.Attachments {
		if err := r.CreateAttachment(ctx, &att); err != nil {
			return err
		}
	}

	return nil
}

func (r *emailRepository) Update(ctx context.Context, email *entity.Email) error {
	query := `
		UPDATE emails
		SET status = $2, error = $3, sent_at = $4, updated_at = $5
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query,
		email.ID,
		email.Status,
		email.Error,
		email.SentAt,
		email.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update email: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}

func (r *emailRepository) CreateAttachment(ctx context.Context, attachment *entity.Attachment) error {
	query := `
		INSERT INTO attachments (
			id, email_id, filename, original_name, mimetype,
			size, path, url, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		attachment.ID,
		attachment.EmailID,
		attachment.Filename,
		attachment.OriginalName,
		attachment.Mimetype,
		attachment.Size,
		attachment.Path,
		attachment.URL,
		attachment.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create attachment: %w", err)
	}

	return nil
}

func (r *emailRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Email, error) {
	query := `
		SELECT
			id, "from", "to", cc, bcc, subject, body, html,
			status, error, sent_at, created_at, updated_at
		FROM emails
		WHERE id = $1
	`

	var email entity.Email
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&email.ID,
		&email.From,
		&email.To,
		&email.CC,
		&email.BCC,
		&email.Subject,
		&email.Body,
		&email.HTML,
		&email.Status,
		&email.Error,
		&email.SentAt,
		&email.CreatedAt,
		&email.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find email: %w", err)
	}

	// Load attachments
	attachments, err := r.FindAttachmentsByEmailID(ctx, email.ID)
	if err != nil {
		return nil, err
	}
	email.Attachments = attachments

	return &email, nil
}

func (r *emailRepository) FindAttachmentsByEmailID(ctx context.Context, emailID uuid.UUID) ([]entity.Attachment, error) {
	query := `
		SELECT
			id, email_id, filename, original_name, mimetype,
			size, path, url, created_at
		FROM attachments
		WHERE email_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.Pool.Query(ctx, query, emailID)

	if err != nil {
		return nil, fmt.Errorf("failed to find attachments: %w", err)
	}

	defer rows.Close()

	var attachments []entity.Attachment

	for rows.Next() {
		var att entity.Attachment
		err := rows.Scan(
			&att.ID,
			&att.EmailID,
			&att.Filename,
			&att.OriginalName,
			&att.Mimetype,
			&att.Size,
			&att.Path,
			&att.URL,
			&att.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan attachment: %w", err)
		}

		attachments = append(attachments, att)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating attachments: %w", err)
	}

	return attachments, nil
}
