FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o api ./cmd/api/

FROM alpine:3.21
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/api /usr/local/bin/api

RUN mkdir -p /data

ENTRYPOINT ["api"]
