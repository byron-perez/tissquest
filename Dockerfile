############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache 'git=~2'
# add c compiler kit
RUN apk add build-base

# Install dependencies
ENV GO111MODULE=on
WORKDIR $GOPATH/src/packages/goginapp/
COPY . .

# Fetch dependencies.
# Using go get.
RUN go get -d -v

# Build the binary.
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o /go/main cmd/api-server-gin/main.go

############################
# STEP 2 build a small image
############################
FROM alpine:3

WORKDIR /

# Copy our static executable.
COPY --from=builder /go/main /go/main

# Copy assets
COPY web /go/web

ENV PORT=8080
ENV GIN_MODE=debug

WORKDIR /go
COPY .env.example .env

EXPOSE 8080
# Run the Go Gin binary.
ENTRYPOINT ["/go/main"]
