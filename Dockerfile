FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o log_analyzer main.go

FROM alpine:latest

ENV TERM=xterm-256color

RUN apk add --no-cache ncurses

WORKDIR /app

COPY --from=builder /app/log_analyzer .
COPY --from=builder /app/config ./config

RUN mkdir ./docs && chmod 755 ./docs

ENTRYPOINT ["./log_analyzer"]