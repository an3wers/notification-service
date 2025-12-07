package entity

import (
	"time"

	"github.com/google/uuid"
)

type Attachment struct {
	ID           uuid.UUID
	EmailID      uuid.UUID
	Filename     string
	OriginalName string
	Mimetype     string
	Size         int64
	Path         string
	URL          *string
	CreatedAt    time.Time
}

func NewAttachment(emailID uuid.UUID, filename, originalName, mimetype string, size int64, path string, url *string) *Attachment {
	return &Attachment{
		ID:           uuid.New(),
		EmailID:      emailID,
		Filename:     filename,
		OriginalName: originalName,
		Mimetype:     mimetype,
		Size:         size,
		Path:         path,
		URL:          url,
		CreatedAt:    time.Now(),
	}
}
