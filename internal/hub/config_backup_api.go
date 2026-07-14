package hub

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

func (h *Hub) exportConfigBackup(e *core.RequestEvent) error {
	var request ConfigBackupExportRequest
	if err := e.BindBody(&request); err != nil {
		return e.BadRequestError("invalid backup export request", err)
	}
	document, warnings, err := h.buildConfigBackupDocument(configBackupExportOptions{
		IncludeSecrets: request.IncludeSecrets,
		Credential:     request.EncryptionCredential,
		Sections:       request.Sections,
	})
	if err != nil {
		return e.BadRequestError(err.Error(), err)
	}
	content, err := marshalConfigBackupDocument(document)
	if err != nil {
		return err
	}
	filename := "beszel-config-" + time.Now().UTC().Format("20060102-150405") + ".yml"
	return e.JSON(http.StatusOK, ConfigBackupExportResponse{
		Filename:      filename,
		ContentType:   "application/x-yaml",
		BackupVersion: ConfigBackupVersion,
		Warnings:      warnings,
		Content:       content,
	})
}

func (h *Hub) validateConfigBackup(e *core.RequestEvent) error {
	var request ConfigBackupValidationRequest
	if err := e.BindBody(&request); err != nil {
		return e.BadRequestError("invalid backup validation request", err)
	}
	document, err := parseConfigBackupDocument(request.Content)
	if err != nil {
		return e.BadRequestError(err.Error(), err)
	}
	preview, err := h.previewConfigBackup(document, request.Content, request.DecryptionCredential)
	if err != nil {
		return e.BadRequestError(err.Error(), err)
	}
	if preview.Summary.Conflict > 0 || preview.Summary.Error > 0 {
		return e.JSON(http.StatusConflict, preview)
	}
	return e.JSON(http.StatusOK, preview)
}

func (h *Hub) restoreConfigBackup(e *core.RequestEvent) error {
	var request ConfigBackupRestoreRequest
	if err := e.BindBody(&request); err != nil {
		return e.BadRequestError("invalid backup restore request", err)
	}
	if request.Mode != "" && request.Mode != ConfigBackupMode {
		return e.BadRequestError("only merge restore mode is supported", nil)
	}
	document, err := parseConfigBackupDocument(request.Content)
	if err != nil {
		return e.BadRequestError(err.Error(), err)
	}
	expectedPreviewID := configBackupPreviewID(request.Content, request.DecryptionCredential)
	if strings.TrimSpace(request.PreviewID) != expectedPreviewID {
		return e.BadRequestError("previewId does not match backup content", nil)
	}
	preview, err := h.previewConfigBackup(document, request.Content, request.DecryptionCredential)
	if err != nil {
		return e.BadRequestError(err.Error(), err)
	}
	if preview.Summary.Conflict > 0 || preview.Summary.Error > 0 {
		return e.JSON(http.StatusConflict, preview)
	}
	applied, warnings, err := h.applyConfigBackup(document, request.DecryptionCredential)
	if err != nil {
		var sectionErr *configBackupSectionRestoreError
		if errors.As(err, &sectionErr) {
			return e.JSON(http.StatusUnprocessableEntity, ConfigBackupRestoreFailureResponse{
				Mode: ConfigBackupMode, Applied: sectionErr.Applied, CompletedSections: sectionErr.CompletedSections,
				FailedSection: sectionErr.Section, Warnings: append(preview.Warnings, warnings...), Error: err.Error(),
			})
		}
		return e.JSON(http.StatusUnprocessableEntity, ConfigBackupRestoreFailureResponse{
			Mode: ConfigBackupMode, Applied: applied, Warnings: append(preview.Warnings, warnings...), Error: err.Error(),
		})
	}
	return e.JSON(http.StatusOK, ConfigBackupRestoreResponse{
		Mode:     ConfigBackupMode,
		Applied:  applied,
		Warnings: append(preview.Warnings, warnings...),
	})
}
