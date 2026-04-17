# ── Development image with air hot-reload ─────────────────────────────────────
FROM golang:1.24-alpine

# Allow Go to auto-download newer toolchains required by dependencies
ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git curl

# Install air for hot-reload in development
RUN go install github.com/air-verse/air@v1.61.7

WORKDIR /app

# Pre-download dependencies (cached Docker layer)
COPY go.mod go.sum ./
RUN go mod download

# Source is volume-mounted in docker-compose for dev, copied here for prod builds
COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
