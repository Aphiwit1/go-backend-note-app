package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/aphiwit1/notes-app/ent"
	"github.com/aphiwit1/notes-app/ent/note"
	"github.com/aphiwit1/notes-app/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// NoteHandlers เป็น struct ที่เก็บ Ent client และ context
type NoteHandlers struct {
	client *ent.Client
	ctx    context.Context
}

// NewNoteHandlers สร้าง instance ใหม่ของ NoteHandlers
func NewNoteHandlers(client *ent.Client, ctx context.Context) *NoteHandlers {
	return &NoteHandlers{
		client: client,
		ctx:    ctx,
	}
}

// GetNotes handler สำหรับ GET /notes
func (h *NoteHandlers) GetNotes(c *gin.Context) {
	title := c.Query("title")
	content := c.Query("content")

	query := h.client.Note.Query()

	if title != "" {
		query = query.Where(note.TitleContains(title))
	}
	if content != "" {
		query = query.Where(note.ContentContains(content))
	}

	notes, err := query.Order(ent.Desc("created_at")).All(h.ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, notes)
}

// CreateNote handler สำหรับ POST /notes
func (h *NoteHandlers) CreateNote(c *gin.Context) {
	var body models.CreateNoteInput

	if err := c.ShouldBindJSON(&body); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			out := make([]models.ErrorMsg, len(ve))
			for i, fe := range ve {
				out[i] = models.ErrorMsg{Field: fe.Field(), Message: models.GetErrorMsg(fe)}
			}
			c.JSON(http.StatusBadRequest, gin.H{"errors": out})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	n, err := h.client.Note.Create().SetTitle(body.Title).SetContent(body.Content).Save(h.ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, n)
}

// GetNoteByID handler สำหรับ GET /notes/:id
func (h *NoteHandlers) GetNoteByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	n, err := h.client.Note.Get(h.ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, n)
}

// UpdateNote handler สำหรับ PUT /notes/:id
func (h *NoteHandlers) UpdateNote(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var body struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n, err := h.client.Note.UpdateOneID(id).
		SetTitle(body.Title).
		SetContent(body.Content).
		Save(h.ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, n)
}

// DeleteNote handler สำหรับ DELETE /notes/:id
func (h *NoteHandlers) DeleteNote(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.client.Note.DeleteOneID(id).Exec(h.ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
