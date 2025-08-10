package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/aphiwit1/notes-app/ent"
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

	api := r.Group("/api")
	{
		api.GET("/notes", func(c *gin.Context) {
			notes, err := client.Note.Query().Order(ent.Desc("created_at")).All(ctx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, notes)
		})

		api.POST("/notes", func(c *gin.Context) {
			var body struct {
				Title   string `json:"title" binding:"required"`
				Content string `json:"content"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
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

		api.GET("/notes/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))
			n, err := client.Note.Get(ctx, id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.JSON(http.StatusOK, n)
		})

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

		api.DELETE("/notes/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))
			if err := client.Note.DeleteOneID(id).Exec(ctx); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})
	}

	// อ่าน PORT จาก environment variable (ถ้าไม่มี ให้ default เป็น 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server started at :" + port)
	r.Run(":" + port)
}
