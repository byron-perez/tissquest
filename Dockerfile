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
WORKDIR /app
COPY . .

# Fetch dependencies.
# Using go get.
RUN go get -d -v

# Build the binary.
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o ./main cmd/api-server-gin/main.go

############################
# STEP 2 build a small image
############################
FROM alpine:3

WORKDIR /app

# Copy our static executable.
COPY --from=builder /app/main .

# Copy assets
COPY --from=builder /app/web/ ./web/
# insecure!
COPY --from=builder /app/.env.example ./.env

ENV PORT=8080
ENV GIN_MODE=debug

EXPOSE 8080
# Run the Go Gin binary.
ENTRYPOINT ["./main"]
