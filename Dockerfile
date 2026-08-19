# Multi-stage build for Chatwoot Lite WhatsApp Gateway
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

WORKDIR /app

# Install git and certificates
RUN apk add --no-cache git ca-certificates tzdata

# Cache go modules
COPY go.mod go.sum* ./
RUN go mod download || true

# Copy source code
COPY . .

# Build API & Worker binaries for the target architecture
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -ldflags="-w -s" -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -ldflags="-w -s" -o /bin/worker ./cmd/worker

# Final minimal runtime image
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata ffmpeg

WORKDIR /app

# Copy binaries
COPY --from=builder /bin/api /app/api
COPY --from=builder /bin/worker /app/worker

# Copy migrations and static web assets
COPY migrations/ /app/migrations/
COPY web/static/ /app/web/static/

EXPOSE 8080

CMD ["/app/api"]
