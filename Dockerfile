# Go sürümünü 1.25 veya daha yeni bir sürüme yükseltin
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Bağımlılıkları kopyala ve indir
COPY go.mod go.sum ./
RUN go mod download

# Kaynak kodları kopyala
COPY . .

# Uygulamayı derle
RUN CGO_ENABLED=0 GOOS=linux go build -o log_analyzer main.go

# Çalıştırma aşaması
FROM alpine:latest

WORKDIR /app

# Derlenen binary'yi ve gerekli config dosyalarını kopyala
COPY --from=builder /app/log_analyzer .
COPY --from=builder /app/config ./config

# Raporların kaydedileceği docs klasörünü oluştur
RUN mkdir ./docs && chmod 755 ./docs

# Uygulamayı çalıştır
ENTRYPOINT ["./log_analyzer"]