FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /agrolang-api ./cmd/api

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /agrolang-api /agrolang-api
EXPOSE 8080
ENTRYPOINT ["/agrolang-api"]
