# ---------- Build Stage ----------
FROM golang:1.24-alpine3.22 AS build

# Install build dependencies
RUN apk --no-cache add gcc g++ make ca-certificates

# Set working directory
WORKDIR /go/src/github.com/PranavTrip/go-grpc-graphql-ms

# Copy only the necessary files for dependency resolution and build
COPY go.mod go.sum ./
# COPY vendor vendor
COPY account account

# Build the account service binary
RUN GO111MODULE=on go build -o /go/bin/account ./account/cmd/account

# ---------- Runtime Stage ----------
FROM alpine:3.18

# Create a non-root user for better security
RUN addgroup -S app && adduser -S app -G app

# Set working directory
WORKDIR /usr/bin

# Copy the compiled binary from build stage
COPY --from=build /go/bin/account .

# Use non-root user
USER app

# Expose the default application port
EXPOSE 8080

# Run the application
CMD ["account"]
