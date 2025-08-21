package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aphiwit1/notes-app/ent"
	"github.com/aphiwit1/notes-app/internal/handlers"
	"github.com/aphiwit1/notes-app/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	_ "github.com/mattn/go-sqlite3"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET")) // 🔑 secret สำหรับ sign

// สร้าง token
func GenerateToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 1).Unix(), // 1 ชั่วโมงหมดอายุ
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
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

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.Use(middleware.TimerMiddleware())

	// Login endpoint
	r.POST("/login", func(c *gin.Context) {
		var body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
			return
		}

		log.Println("Login attempt:", body.Username)
		log.Println("Password:", body.Password)

		// สมมุติ username=admin password=1234
		if body.Username != "admin" || body.Password != "1234" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong username or password"})
			return
		}

		// สร้าง token
		token, _ := GenerateToken(body.Username)
		c.JSON(http.StatusOK, gin.H{"token": token})
	})

	r.GET("/profile", func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		// ตัดคำว่า "Bearer "
		tokenStr := authHeader[7:]

		// ตรวจสอบ token
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Hello, you are authorized"})
	})

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware()) // // ✅ ป้องกันเฉพาะ /api/*
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
