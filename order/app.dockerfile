# Build stage
FROM golang:1.21-alpine3.18 AS build

# Install necessary packages
RUN apk --no-cache add gcc g++ make ca-certificates

# Set working directory inside the container
WORKDIR /go/src/github.com/PranavTrip/go-grpc-graphql-ms

# Copy module files and source code
COPY go.mod go.sum ./
COPY vendor vendor
COPY account account
COPY catalog catalog
COPY order order

# Build the application using vendored dependencies
RUN go build -mod=vendor -o /go/bin/app ./order/cmd/order

# Runtime stage
FROM alpine:3.18

# Create a non-root user for better security
RUN addgroup -S app && adduser -S app -G app

WORKDIR /usr/bin

# Copy the built binary from the build stage
COPY --from=build /go/bin/app .

# Use non-root user
USER app

# Expose application port
EXPOSE 8080

# Run the binary
CMD ["app"]
