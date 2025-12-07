package dto

type SendEmailRequest struct {
	To      []string `json:"to" validate:"required,dive,email"`
	CC      []string `json:"cc,omitempty" validate:"omitempty,dive,email"`
	BCC     []string `json:"bcc,omitempty" validate:"omitempty,dive,email"`
	Subject string   `json:"subject" validate:"required,min=1,max=255"`
	Body    string   `json:"body" validate:"required,min=1"`
	HTML    *string  `json:"html,omitempty"`
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
