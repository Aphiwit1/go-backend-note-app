# ---------- Stage 1: Build ----------
FROM golang:1.24.2-bullseye AS builder

# ติดตั้ง gcc เพื่อแก้ปัญหา runtime/cgo
RUN apt-get update && apt-get install -y gcc libc6-dev && rm -rf /var/lib/apt/lists/*

# ตั้งค่า working directory
WORKDIR /app

# คัดลอก go.mod และ go.sum ก่อน เพื่อ cache dependency
COPY go.mod go.sum ./

# ดาวน์โหลด dependencies
RUN go mod download

# คัดลอก source code ทั้งหมด
COPY . .

# สร้าง binary
RUN go build -o main .

# ---------- Stage 2: Run ----------
FROM debian:bullseye-slim

# ตั้ง working directory
WORKDIR /app

# คัดลอก binary จาก stage แรก
COPY --from=builder /app/main .

# เปิดพอร์ต
EXPOSE 8080

# คำสั่งเริ่มต้น
CMD ["./main"]
