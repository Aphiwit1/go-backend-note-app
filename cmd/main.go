package main

import (
	"context"
	"log"
	"os"

	"github.com/aphiwit1/notes-app/ent"
	"github.com/aphiwit1/notes-app/internal/handlers"
	"github.com/aphiwit1/notes-app/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

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
		// ใช้ handlers จากไฟล์ note.go
		noteHandlers := handlers.NewNoteHandlers(client, ctx)

		api.GET("/notes", noteHandlers.GetNotes)
		api.POST("/notes", noteHandlers.CreateNote)
		api.GET("/notes/:id", noteHandlers.GetNoteByID)
		api.PUT("/notes/:id", noteHandlers.UpdateNote)
		api.DELETE("/notes/:id", noteHandlers.DeleteNote)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server started at :" + port)
	r.Run(":" + port)
}
