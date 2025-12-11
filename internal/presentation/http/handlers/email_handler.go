package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	logger           *logger.Logger
}

func NewEmailHandler(
	sendEmailUC *usecase.SendEmailUseCase,
	getEmailStatusUC *usecase.GetEmailStatusUseCase,
	storageCfg config.StorageConfig,
	logger *logger.Logger,
) *EmailHandler {
	return &EmailHandler{
		sendEmailUC:      sendEmailUC,
		getEmailStatusUC: getEmailStatusUC,
		validator:        validator.New(),
		storageCfg:       storageCfg,
		logger:           logger,
	}
}

func (h *EmailHandler) SendEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse multipart form
	if err := r.ParseMultipartForm(h.storageCfg.MaxFileSize); err != nil {
		h.respondError(w, http.StatusBadRequest, "failed to parse form", err)
		return
	}

	// Extract email data
	var req dto.SendEmailRequest

	// Parse JSON from form field or directly from body
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.respondError(w, http.StatusBadRequest, "invalid request body", err)
			return
		}
	} else {
		// Parse from form fields
		req.To = r.Form["to"]
		req.CC = r.Form["cc"]
		req.BCC = r.Form["bcc"]
		req.Subject = r.FormValue("subject")
		req.Body = r.FormValue("body")

		if htmlValue := r.FormValue("html"); htmlValue != "" {
			req.HTML = &htmlValue
		}
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
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
	email, err := h.sendEmailUC.Execute(ctx, &req, attachments)

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

func (h *EmailHandler) respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *EmailHandler) respondError(w http.ResponseWriter, status int, message string, err error) {
	h.logger.Error(message, zap.String("error", err.Error()))

	errorResponse := map[string]any{
		"error":   message,
		"details": err.Error(),
	}

	h.respondJSON(w, status, errorResponse)
}
