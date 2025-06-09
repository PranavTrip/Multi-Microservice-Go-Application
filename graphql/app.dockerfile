# Stage 1: Build
FROM golang:1.24-alpine3.22 AS build

RUN apk --no-cache add gcc g++ make ca-certificates

WORKDIR /go/src/github.com/PranavTrip/go-grpc-graphql-ms

COPY go.mod go.sum ./
COPY account account
COPY catalog catalog
COPY order order
COPY graphql graphql

WORKDIR /go/src/github.com/PranavTrip/go-grpc-graphql-ms/graphql

RUN go build -o /go/bin/app .

# Stage 2: Final image
FROM alpine:3.11

WORKDIR /usr/bin

# Optional: Add ca-certs if needed for HTTPS
RUN apk --no-cache add ca-certificates

COPY --from=build /go/bin/app .

EXPOSE 8080

# ✅ Ensure correct binary runs
CMD ["./app"]
