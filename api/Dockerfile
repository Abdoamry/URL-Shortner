# Stage 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /builder

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main main.go

# Stage 2: Runtime
FROM alpine

WORKDIR /app

RUN adduser -S -D -H -h /app appuser

COPY --from=builder /builder/main /app/main

USER appuser

EXPOSE 3000

CMD ["./main"]
