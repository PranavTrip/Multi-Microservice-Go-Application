# ---------- Build Stage ----------
FROM golang:1.21-alpine3.18 AS build

# Install build dependencies
RUN apk --no-cache add gcc g++ make ca-certificates

# Set working directory
WORKDIR /go/src/github.com/PranavTrip/go-grpc-graphql-ms

# Copy dependency-related files and source code
COPY go.mod go.sum ./
COPY vendor vendor
COPY catalog catalog

# Build the catalog service binary
RUN go build -mod=vendor -o /go/bin/catalog ./catalog/cmd/catalog

# ---------- Runtime Stage ----------
FROM alpine:3.18

# Create a non-root user for better security
RUN addgroup -S app && adduser -S app -G app

# Set working directory
WORKDIR /usr/bin

# Copy the built binary
COPY --from=build /go/bin/catalog .

# Run as non-root user
USER app

# Expose application port
EXPOSE 8080

# Run the catalog service binary
CMD ["app"]
