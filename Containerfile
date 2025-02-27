FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy go.mod first for better layer caching
COPY go.mod .
RUN go mod download

# Copy source code
COPY main.go .
COPY pkg/ pkg/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dnf-update-api ./

FROM fedora:latest

# Install dnf utilities
RUN dnf install -y dnf-utils && \
    dnf clean all && \
    rm -rf /etc/yum.repos.d/*

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/dnf-update-api .

# Create the entrypoint script
RUN printf "#!/bin/bash\n\
exec /app/dnf-update-api" > /app/entrypoint.sh && \
chmod +x /app/entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]