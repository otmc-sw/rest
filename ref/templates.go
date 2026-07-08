/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
**/

package handlers

import (
	"database/sql"
	"encoding/json"
	db "otmc/app/db/sqlc"

	"github.com/gofiber/fiber/v2"
)

type TemplateResponse struct {
	ID          int64       `json:"id"`
	Title       string      `json:"title"`
	Content     interface{} `json:"content,omitempty"`
	DocIcon     interface{} `json:"doc_icon,omitempty"`
	Description interface{} `json:"description,omitempty"`
	IsSystem    bool        `json:"is_system"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
}

func toTemplateResponse(tpl db.Template) TemplateResponse {
	var content interface{}
	contentStr := nullStringToString(tpl.Content)
	if contentStr != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(contentStr), &parsed); err == nil {
			content = parsed
		} else {
			content = contentStr
		}
	}

	var docIcon interface{}
	docIconStr := nullStringToString(tpl.DocIcon)
	if docIconStr != "" {
		var parsedIcon interface{}
		if err := json.Unmarshal([]byte(docIconStr), &parsedIcon); err == nil {
			docIcon = parsedIcon
		} else {
			docIcon = docIconStr
		}
	}

	var description interface{}
	descStr := nullStringToString(tpl.Description)
	if descStr != "" {
		description = descStr
	}

	return TemplateResponse{
		ID:          tpl.ID,
		Title:       tpl.Title,
		Content:     content,
		DocIcon:     docIcon,
		Description: description,
		IsSystem:    tpl.IsSystem.Valid && tpl.IsSystem.Bool,
		CreatedAt:   nullTimeToString(tpl.CreatedAt),
		UpdatedAt:   nullTimeToString(tpl.UpdatedAt),
	}
}

func toTemplateResponses(templates []db.Template) []TemplateResponse {
	responses := make([]TemplateResponse, len(templates))
	for i, tpl := range templates {
		responses[i] = toTemplateResponse(tpl)
	}
	return responses
}

func (h *Handler) GetTemplatesHandler(c *fiber.Ctx) error {
	templates, err := h.db.ListTemplates(c.Context())
	if err != nil {
		return errJSON(c, "list templates failed", err)
	}
	return okJSON(c, toTemplateResponses(templates))
}

func (h *Handler) GetTemplateByIDHandler(c *fiber.Ctx) error {
	id, err := validateAndExtractInt64ID(c, "id")
	if err != nil {
		return err
	}

	template, err := h.db.GetTemplate(c.Context(), id)
	if err != nil {
		return notFoundJSON(c, "Template not found")
	}

	return okJSON(c, toTemplateResponse(template))
}

func (h *Handler) CreateTemplateHandler(c *fiber.Ctx) error {
	var req struct {
		Title       string          `json:"title"`
		Content     json.RawMessage `json:"content"`
		DocIcon     *string         `json:"doc_icon"`
		Description *string         `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return badRequestJSON(c, "Invalid request body", err)
	}

	if req.Title == "" {
		return badRequestJSON(c, "Validation failed", "title is required")
	}

	var content sql.NullString
	if len(req.Content) > 0 {
		if json.Valid(req.Content) {
			content = sql.NullString{String: string(req.Content), Valid: true}
		} else {
			contentStr := string(req.Content)
			if contentStr != "" {
				content = sql.NullString{String: contentStr, Valid: true}
			}
		}
	}

	docIcon := stringPtrOrNull(req.DocIcon)
	description := stringPtrOrNull(req.Description)

	result, err := h.db.CreateTemplate(c.Context(), db.CreateTemplateParams{
		Title:       req.Title,
		Content:     content,
		DocIcon:     docIcon,
		Description: description,
		IsSystem:    sql.NullBool{Bool: false, Valid: true},
	})
	if err != nil {
		return errJSON(c, "create template failed", err)
	}

	lastID, _ := result.LastInsertId()
	template, err := h.db.GetTemplate(c.Context(), lastID)
	if err != nil {
		return errJSON(c, "get created template failed", err)
	}

	return createdJSON(c, toTemplateResponse(template))
}

func (h *Handler) UpdateTemplateHandler(c *fiber.Ctx) error {
	id, err := validateAndExtractInt64ID(c, "id")
	if err != nil {
		return err
	}

	existingTpl, err := h.db.GetTemplate(c.Context(), id)
	if err != nil {
		return notFoundJSON(c, "Template not found")
	}

	var req struct {
		Title       *string         `json:"title"`
		Content     json.RawMessage `json:"content"`
		DocIcon     *string         `json:"doc_icon"`
		Description *string         `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return badRequestJSON(c, "Invalid request body", err)
	}

	title := req.Title
	if title == nil || *title == "" {
		title = &existingTpl.Title
	}

	var content sql.NullString
	if len(req.Content) > 0 {
		if json.Valid(req.Content) {
			content = sql.NullString{String: string(req.Content), Valid: true}
		} else {
			contentStr := string(req.Content)
			if contentStr != "" {
				content = sql.NullString{String: contentStr, Valid: true}
			}
		}
	}
	if !content.Valid {
		content = existingTpl.Content
	}

	docIcon := stringPtrOrNull(req.DocIcon)
	if !docIcon.Valid {
		docIcon = existingTpl.DocIcon
	}

	description := stringPtrOrNull(req.Description)
	if !description.Valid {
		description = existingTpl.Description
	}

	if err := h.db.UpdateTemplate(c.Context(), db.UpdateTemplateParams{
		ID:          id,
		Title:       *title,
		Content:     content,
		DocIcon:     docIcon,
		Description: description,
	}); err != nil {
		return errJSON(c, "update template failed", err)
	}

	template, err := h.db.GetTemplate(c.Context(), id)
	if err != nil {
		return errJSON(c, "get updated template failed", err)
	}

	return okJSON(c, toTemplateResponse(template))
}

func (h *Handler) DeleteTemplateHandler(c *fiber.Ctx) error {
	id, err := validateAndExtractInt64ID(c, "id")
	if err != nil {
		return err
	}

	template, err := h.db.GetTemplate(c.Context(), id)
	if err != nil {
		return notFoundJSON(c, "Template not found")
	}

	if template.IsSystem.Valid && template.IsSystem.Bool {
		return badRequestJSON(c, "System templates cannot be deleted")
	}

	if err := h.db.DeleteTemplate(c.Context(), id); err != nil {
		return errJSON(c, "delete template failed", err)
	}

	return noContentJSON(c)
}

func (h *Handler) CreateDocumentFromTemplateHandler(c *fiber.Ctx) error {
	id, err := validateAndExtractInt64ID(c, "id")
	if err != nil {
		return err
	}

	template, err := h.db.GetTemplate(c.Context(), id)
	if err != nil {
		return notFoundJSON(c, "Template not found")
	}

	var req struct {
		ParentID  *int64  `json:"parent_id"`
		DocType   *string `json:"doc_type"`
		DocStatus *string `json:"doc_status"`
		Title     *string `json:"title"`
	}

	if err := c.BodyParser(&req); err != nil {
		return badRequestJSON(c, "Invalid request body", err)
	}

	title := template.Title
	if req.Title != nil && *req.Title != "" {
		title = *req.Title
	} else {
		title = template.Title
	}

	parentID := int64OrNull(req.ParentID)
	docType := stringPtrOrNull(req.DocType)
	if !docType.Valid {
		docType = sql.NullString{String: "document", Valid: true}
	}
	docStatus := stringPtrOrNull(req.DocStatus)
	if !docStatus.Valid {
		docStatus = sql.NullString{String: "draft", Valid: true}
	}

	docIcon := template.DocIcon
	if !docIcon.Valid {
		docIcon = sql.NullString{String: `{"type":"emoji","name":"📚"}`, Valid: true}
	}

	result, err := h.db.CreateDocumentFromTemplate(c.Context(), db.CreateDocumentFromTemplateParams{
		ParentID:  parentID,
		DocType:   docType,
		DocStatus: docStatus,
		Title:     title,
		Content:   template.Content,
		DocIcon:   docIcon,
	})
	if err != nil {
		return errJSON(c, "create document from template failed", err)
	}

	lastID, _ := result.LastInsertId()
	document, err := h.db.GetDocument(c.Context(), lastID)
	if err != nil {
		return errJSON(c, "get created document failed", err)
	}

	return createdJSON(c, toDocumentResponse(document))
}
