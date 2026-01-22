# ==========================================
# Stage 1: The Builder
# ==========================================
FROM golang:1.24-alpine AS builder

# install git and certificates (needed to fetch dependencies)
RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Build the binary
# -o main: name the output binary "main"
# CGO_ENABLED=0: Create a statically linked binary (no dependency on C libraries)
# -ldflags="-w -s": Strip debug symbols to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/server

# ==========================================
# Stage 2: The Runner (Production Ready)
# ==========================================
FROM alpine:3.19

# Add standard CA certificates so we can call HTTPS endpoints (OpenAI)
RUN apk add --no-cache ca-certificates

# Security: Create a non-root user. 
# Never run containers as root in production!
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Switch to the non-root user
USER appuser

# Expose the port (Documentation only, technically)
EXPOSE 8080

# The command to run when the container starts
CMD ["./main"]