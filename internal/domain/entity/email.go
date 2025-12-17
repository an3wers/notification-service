package entity

import (
	"time"

	"github.com/google/uuid"
)

type EmailStatus string

const (
	StatusPending EmailStatus = "PENDING"
	StatusQueued  EmailStatus = "QUEUED"
	StatusSent    EmailStatus = "SENT"
	StatusFailed  EmailStatus = "FAILED"
)

type Email struct {
	ID          uuid.UUID
	From        string
	DisplayName string
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	Body        string
	HTML        *string
	Status      EmailStatus
	Error       *string
	SentAt      *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	Attachments []Attachment
}

func NewEmail(from string, to []string, displayName, subject, body string) *Email {
	now := time.Now().UTC()
	return &Email{
		ID:          uuid.New(),
		From:        from,
		To:          to,
		DisplayName: displayName,
		CC:          []string{},
		BCC:         []string{},
		Subject:     subject,
		Body:        body,
		Status:      StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
		Attachments: []Attachment{},
	}
}

func (e *Email) MarkAsSent() {
	now := time.Now().UTC()
	e.Status = StatusSent
	e.SentAt = &now
	e.UpdatedAt = now
}

func (e *Email) MarkAsFailed(errMsg string) {
	e.Status = StatusFailed
	e.Error = &errMsg
	e.UpdatedAt = time.Now().UTC()
}

func (e *Email) MarkAsQueued() {
	e.Status = StatusQueued
	e.UpdatedAt = time.Now().UTC()
}
