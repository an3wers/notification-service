package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/an3wers/notification-serv/internal/application/dto"
	"github.com/an3wers/notification-serv/internal/application/usecase"
	"github.com/an3wers/notification-serv/internal/domain/entity"
	"github.com/an3wers/notification-serv/internal/pkg/config"
	apperrors "github.com/an3wers/notification-serv/internal/pkg/errors"
	"github.com/an3wers/notification-serv/internal/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type EmailHandler struct {
	sendEmailUC      *usecase.SendEmailUseCase
	getEmailStatusUC *usecase.GetEmailStatusUseCase
	validator        *validator.Validate
	storageCfg       config.StorageConfig
	serverCfg        config.ServerConfig
	logger           *logger.Logger
}

func NewEmailHandler(
	sendEmailUC *usecase.SendEmailUseCase,
	getEmailStatusUC *usecase.GetEmailStatusUseCase,
	storageCfg config.StorageConfig,
	serverCfg config.ServerConfig,
	logger *logger.Logger,
) *EmailHandler {
	return &EmailHandler{
		sendEmailUC:      sendEmailUC,
		getEmailStatusUC: getEmailStatusUC,
		validator:        validator.New(),
		storageCfg:       storageCfg,
		serverCfg:        serverCfg,
		logger:           logger,
	}
}

func (h *EmailHandler) SendEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	isValidKey := h.checkSecretKey(r)

	if !isValidKey {
		h.respondError(w, http.StatusUnauthorized, "invalid secret key", errors.New("invalid secret key"))
		return
	}

	// Extract email data
	// var req dto.SendEmailRequest

	// normalize request
	var normalizedReq dto.SendEmailNormalizedRequest

	// Parse JSON from form field or directly from body
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {

		data, err := h.normalizeRequestFromJson(r)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "invalid request body", err)
			return
		}

		normalizedReq = *data

	} else {
		// Parse multipart form
		if err := r.ParseMultipartForm(h.storageCfg.MaxFileSize); err != nil {
			h.respondError(w, http.StatusBadRequest, "failed to parse form", err)
			return
		}

		data, err := h.normalizeRequestFromFormData(r)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "invalid request body", err)
			return
		}

		normalizedReq = *data
	}

	// Validate normalized request
	if err := h.validator.Struct(normalizedReq); err != nil {
		h.respondError(w, http.StatusBadRequest, "validation failed", err)
		return
	}

	// Handle file uploads
	var attachments []dto.AttachmentDTO

	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		files := r.MultipartForm.File["files"]

		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				h.respondError(w, http.StatusBadRequest, "failed to open file", err)
				return
			}

			defer file.Close()

			// Generate unique filename
			filename := uuid.New().String() + filepath.Ext(fileHeader.Filename)
			filepath := filepath.Join(h.storageCfg.LocalPath, filename)

			// Ensure upload directory exists
			if err := os.MkdirAll(h.storageCfg.LocalPath, 0755); err != nil {
				h.respondError(w, http.StatusInternalServerError, "failed to create upload directory", err)
				return
			}

			// Save file
			dst, err := os.Create(filepath)

			if err != nil {
				h.respondError(w, http.StatusInternalServerError, "failed to create file", err)
				return
			}

			defer dst.Close()

			if _, err := io.Copy(dst, file); err != nil {
				h.respondError(w, http.StatusInternalServerError, "failed to save file", err)
				return
			}

			attachments = append(attachments, dto.AttachmentDTO{
				Filename:     filename,
				OriginalName: fileHeader.Filename,
				Mimetype:     fileHeader.Header.Get("Content-Type"),
				Size:         fileHeader.Size,
				Path:         filepath,
			})
		}
	}

	// Execute use case
	email, err := h.sendEmailUC.Execute(ctx, &normalizedReq, attachments)

	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to send email", err)
		return
	}

	// Build response
	response := h.buildEmailResponse(email)
	h.respondJSON(w, http.StatusCreated, response)
}

func (h *EmailHandler) GetEmailStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	emailID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid email ID", err)
		return
	}

	email, err := h.getEmailStatusUC.Execute(ctx, emailID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "email not found", err)
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to get email", err)
		return
	}

	response := h.buildEmailResponse(email)
	h.respondJSON(w, http.StatusOK, response)
}

func (h *EmailHandler) buildEmailResponse(email *entity.Email) *dto.EmailResponse {
	resp := &dto.EmailResponse{
		ID:        email.ID.String(),
		Status:    string(email.Status),
		To:        email.To,
		Subject:   email.Subject,
		CreatedAt: email.CreatedAt.Format(time.RFC3339),
		Error:     email.Error,
	}

	if email.SentAt != nil {
		sentAt := email.SentAt.Format(time.RFC3339)
		resp.SentAt = &sentAt
	}

	return resp
}

func (h *EmailHandler) checkSecretKey(r *http.Request) bool {
	return r.Header.Get("ssy") == h.serverCfg.SecretKey
}

func (h *EmailHandler) normalizeRequestFromJson(r *http.Request) (*dto.SendEmailNormalizedRequest, error) {
	var req dto.SendEmailRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	// Parse To field - handle semicolon-separated strings
	to := h.parseEmailList(req.To)

	// Parse CC and BCC fields - handle semicolon-separated strings
	cc := h.parseEmailList(req.CC)
	bcc := h.parseEmailList(req.BCC)

	normalized := &dto.SendEmailNormalizedRequest{
		To:          to,
		From:        nil,
		DisplayName: nil,
		CC:          cc,
		BCC:         bcc,
		Subject:     nil,
		Body:        nil,
		HTML:        req.HTML,
	}

	if req.FromEmail != "" {
		normalized.From = &req.FromEmail
	}
	if req.FromDisplayName != "" {
		normalized.DisplayName = &req.FromDisplayName
	}
	if req.Subject != "" {
		normalized.Subject = &req.Subject
	} else if req.Title != "" {
		normalized.Subject = &req.Title
	}
	if req.Body != "" {
		normalized.Body = &req.Body
	} else if req.Message != "" {
		normalized.Body = &req.Message
	}

	return normalized, nil
}

func (h *EmailHandler) normalizeRequestFromFormData(r *http.Request) (*dto.SendEmailNormalizedRequest, error) {

	// Helper for getting string slice (required or optional)
	getStrings := func(key string, required bool) ([]string, error) {
		values := r.MultipartForm.Value[key]
		if required && (len(values) == 0 || (len(values) == 1 && values[0] == "")) {
			return nil, errors.New("missing required field: " + key)
		}
		// Remove empty strings
		filtered := make([]string, 0, len(values))
		for _, v := range values {
			if v != "" {
				filtered = append(filtered, v)
			}
		}
		if required && len(filtered) == 0 {
			return nil, errors.New("missing required field: " + key)
		}
		return filtered, nil
	}

	getStringPtr := func(key string) *string {
		if vals, ok := r.MultipartForm.Value[key]; ok && len(vals) > 0 && vals[0] != "" {
			return &vals[0]
		}
		return nil
	}

	// "to" required
	toRaw, err := getStrings("to", true)
	if err != nil {
		return nil, err
	}
	// Parse To field - handle semicolon-separated strings
	to := h.parseEmailList(toRaw)

	// Optional fields
	ccRaw, _ := getStrings("cc", false)
	cc := h.parseEmailList(ccRaw)
	bccRaw, _ := getStrings("bcc", false)
	bcc := h.parseEmailList(bccRaw)

	from := getStringPtr("fromEmail")
	if from == nil {
		from = getStringPtr("from")
	}
	displayName := getStringPtr("fromDisplayName")
	if displayName == nil {
		displayName = getStringPtr("displayName")
	}

	// subject/title (prefer subject)
	subject := getStringPtr("subject")
	if subject == nil {
		subject = getStringPtr("title")
	}

	// body/message (prefer body)
	body := getStringPtr("body")
	if body == nil {
		body = getStringPtr("message")
	}

	html := getStringPtr("html")

	return &dto.SendEmailNormalizedRequest{
		To:          to,
		From:        from,
		DisplayName: displayName,
		CC:          cc,
		BCC:         bcc,
		Subject:     subject,
		Body:        body,
		HTML:        html,
	}, nil
}

func (h *EmailHandler) respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *EmailHandler) parseEmailList(emails []string) []string {
	if len(emails) == 0 {
		return emails
	}

	var result []string
	for _, emailStr := range emails {
		// Split by semicolon and trim whitespace
		parts := strings.Split(emailStr, ";")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
	}
	return result
}

func (h *EmailHandler) respondError(w http.ResponseWriter, status int, message string, err error) {
	h.logger.Error(message, zap.String("error", err.Error()))

	errorResponse := map[string]any{
		"error":   message,
		"details": err.Error(),
	}

	h.respondJSON(w, status, errorResponse)
}
