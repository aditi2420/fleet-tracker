# syntax=docker/dockerfile:1

####### STAGE 1: build the Go binary #######
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Cache modules
COPY go.mod go.sum ./
RUN go mod download

# Bring in the rest of the code, compile static binary
COPY . .
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -o fleet-tracker ./cmd/server

####### STAGE 2: runtime image #######
FROM alpine:3.18

# Optional: install CA certs if you call external HTTPS endpoints
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy the compiled binary
COPY --from=builder /build/fleet-tracker .

EXPOSE 8080

ENTRYPOINT ["./fleet-tracker"]
