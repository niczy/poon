# Multi-stage build for poon-server
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git protobuf-dev

# Install Go protobuf plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Copy proto files and generate
COPY poon-proto/ ./poon-proto/
RUN cd poon-proto && \
    mkdir -p gen/go && \
    protoc --go_out=gen --go_opt=paths=source_relative \
           --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
           --proto_path=. monorepo.proto && \
    mv gen/monorepo*.pb.go gen/go/

# Copy go mod files
COPY go.mod go.sum ./
COPY poon-server/go.mod poon-server/go.sum ./poon-server/

# Download dependencies
RUN cd poon-server && go mod download

# Copy source code
COPY poon-server/ ./poon-server/

# Build the binary
RUN cd poon-server && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o poon-server .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary
COPY --from=builder /app/poon-server/poon-server .

# Create data directory
RUN mkdir -p /app/data

# Expose port
EXPOSE 50051

# Health check support
RUN apk add --no-cache curl

# Run the binary
CMD ["./poon-server"]