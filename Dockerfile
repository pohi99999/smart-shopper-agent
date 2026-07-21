# First stage: builder
FROM golang:alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o smart-shopper-agent cmd/app/main.go

# Second stage: production
FROM scratch

# Copy the binary
COPY --from=builder /app/smart-shopper-agent /smart-shopper-agent

# Copy the required assets and config files
COPY internal/data/prices.json /internal/data/prices.json

# Expose the application port
EXPOSE 8080

# Start the application
CMD ["/smart-shopper-agent"]
