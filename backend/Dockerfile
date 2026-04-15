# ── Development image with air hot-reload ─────────────────────────────────────
FROM golang:1.22-alpine

RUN apk add --no-cache git curl

# Install air for hot-reload in development
RUN go install github.com/air-verse/air@latest

WORKDIR /app

# Pre-download dependencies (cached Docker layer)
COPY go.mod go.sum ./
RUN go mod download

# Source is volume-mounted in docker-compose for dev, copied here for prod builds
COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
