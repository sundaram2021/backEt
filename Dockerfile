# Stage 1: Build the Go application with Go 1.22.6 on Ubuntu
FROM ubuntu:20.04 AS builder

# Set environment variables for Go installation
ENV GO_VERSION 1.22.6
ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV PATH $GOPATH/bin:$GOROOT/bin:$PATH

# Install required packages
RUN apt-get update && apt-get install -y \
    wget \
    git \
    build-essential \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Download and install Go 1.22.6
RUN wget https://go.dev/dl/go$GO_VERSION.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go$GO_VERSION.linux-amd64.tar.gz && \
    rm go$GO_VERSION.linux-amd64.tar.gz

WORKDIR /app

# Copy the source code
COPY . .

# Build the Go application
RUN go mod tidy
RUN go build -o /app/etcd-backup

# Stage 2: Use a Debian-based image for the final container
FROM debian:buster-slim

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/etcd-backup /app/etcd-backup

# Make sure the binary has execute permissions
RUN chmod +x /app/etcd-backup

# Run the binary
ENTRYPOINT ["/app/etcd-backup"]
