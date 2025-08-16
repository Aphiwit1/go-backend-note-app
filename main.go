package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/aphiwit1/notes-app/ent"
	"github.com/aphiwit1/notes-app/ent/note"
	"github.com/aphiwit1/notes-app/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10" // ใช้ v10 ตามที่ติดตั้ง
	_ "github.com/mattn/go-sqlite3"
)

// CreateNoteInput เป็น struct สำหรับรับข้อมูลจาก request
type CreateNoteInput struct {
	Title   string `json:"title" binding:"required,min=3,max=10"`
	Content string `json:"content" binding:"required,min=3,max=1000"`
}

// ErrorMsg เป็น struct สำหรับรูปแบบ response ของ error
type ErrorMsg struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// getErrorMsg แปลง validator.FieldError ให้เป็นข้อความที่อ่านง่าย
func getErrorMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "Must be at least " + fe.Param() + " characters long"
	case "max":
		return "Must not be more than " + fe.Param() + " characters long"
	}
	return "Unknown error"
}

func main() {
	// เชื่อมต่อฐานข้อมูล
	client, err := ent.Open("sqlite3", "file:notes.db?cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("เปิดฐานข้อมูลไม่ได้: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Auto migrate
	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("สร้าง schema ไม่ได้: %v", err)
	}

	r := gin.Default()
	r.Use(cors.Default())
	r.Use(middleware.TimerMiddleware())

	api := r.Group("/api")
	{
		// GET /notes?title=xxx&content=yyy
		api.GET("/notes", func(c *gin.Context) {
			title := c.Query("title")
			content := c.Query("content")

			query := client.Note.Query()

			if title != "" {
				query = query.Where(note.TitleContains(title))
			}
			if content != "" {
				query = query.Where(note.ContentContains(content))
			}

			notes, err := query.Order(ent.Desc("created_at")).All(ctx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, notes)
		})

		// POST /notes - เพิ่ม Validation
		api.POST("/notes", func(c *gin.Context) {
			var body CreateNoteInput

			if err := c.ShouldBindJSON(&body); err != nil {
				var ve validator.ValidationErrors
				if errors.As(err, &ve) {
					out := make([]ErrorMsg, len(ve))
					for i, fe := range ve {
						out[i] = ErrorMsg{Field: fe.Field(), Message: getErrorMsg(fe)}
					}
					c.JSON(http.StatusBadRequest, gin.H{"errors": out})
					return
				}

				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			n, err := client.Note.Create().SetTitle(body.Title).SetContent(body.Content).Save(ctx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, n)
		})

		// GET /notes/:id
		api.GET("/notes/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))
			n, err := client.Note.Get(ctx, id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.JSON(http.StatusOK, n)
		})

		// PUT /notes/:id
		api.PUT("/notes/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))
			var body struct {
				Title   string `json:"title"`
				Content string `json:"content"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			n, err := client.Note.UpdateOneID(id).
				SetTitle(body.Title).
				SetContent(body.Content).
				Save(ctx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, n)
		})

		// DELETE /notes/:id
		api.DELETE("/notes/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))
			if err := client.Note.DeleteOneID(id).Exec(ctx); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server started at :" + port)
	r.Run(":" + port)
}
