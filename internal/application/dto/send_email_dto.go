package dto

type SendEmailRequest struct {
	To              []string `json:"to" validate:"required,dive,email"`
	FromEmail       string   `json:"fromEmail,omitempty" validate:"omitempty,email"`
	FromDisplayName string   `json:"fromDisplayName,omitempty" validate:"omitempty,min=1,max=255"`
	CC              []string `json:"cc,omitempty" validate:"omitempty,dive,email"`
	BCC             []string `json:"bcc,omitempty" validate:"omitempty,dive,email"`
	Subject         string   `json:"subject,omitempty" validate:"omitempty,min=1,max=255"`
	Title           string   `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Body            string   `json:"body,omitempty" validate:"omitempty,min=1"`
	Message         string   `json:"message,omitempty" validate:"omitempty,min=1"`
	HTML            *string  `json:"html,omitempty"`
}

type SendEmailNormalizedRequest struct {
	To          []string `validate:"required,dive,email"`
	From        *string  `validate:"omitempty,email"`
	DisplayName *string  `validate:"omitempty,min=1,max=255"`
	CC          []string `validate:"omitempty,dive,email"`
	BCC         []string `validate:"omitempty,dive,email"`
	Subject     *string  `validate:"omitempty,min=1,max=255"`
	Body        *string  `validate:"omitempty,min=1"`
	HTML        *string  `validate:"omitempty"`
}

type AttachmentDTO struct {
	Filename     string
	OriginalName string
	Mimetype     string
	Size         int64
	Path         string
}

type EmailResponse struct {
	ID        string   `json:"id"`
	Status    string   `json:"status"`
	To        []string `json:"to"`
	Subject   string   `json:"subject"`
	CreatedAt string   `json:"createdAt"`
	SentAt    *string  `json:"sentAt,omitempty"`
	Error     *string  `json:"error,omitempty"`
}
