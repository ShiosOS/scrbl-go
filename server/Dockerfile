FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build deps for CGO-free sqlite (modernc.org/sqlite is pure Go)
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o scrbl-server .

# --- Runtime ---
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/scrbl-server .

# Data volume for SQLite database
RUN mkdir -p /data
VOLUME /data

ENV DB_PATH=/data/scrbl.db
ENV PORT=8080

EXPOSE 8080

CMD ["./scrbl-server"]
