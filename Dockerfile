FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/hr-system ./cmd/api/main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/bin/hr-system .

EXPOSE 8080
CMD ["./hr-system"]
