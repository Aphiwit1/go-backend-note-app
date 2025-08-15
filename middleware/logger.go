package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// TimerMiddleware บันทึกเวลาที่ใช้ในการประมวลผลแต่ละคำขอ
func TimerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()
		latency := time.Since(t)
		log.Printf("[GIN] %s %s took %v", c.Request.Method, c.Request.URL.Path, latency)
	}
}
