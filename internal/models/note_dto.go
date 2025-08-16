package models

import "github.com/go-playground/validator/v10"

// CreateNoteInput เป็น struct สำหรับรับข้อมูลตอนสร้าง Note
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
func GetErrorMsg(fe validator.FieldError) string {
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
