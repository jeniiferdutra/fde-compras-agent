
FROM golang:1.26.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o mcp-server .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/mcp-server .

EXPOSE 8001

CMD ["./mcp-server"]
