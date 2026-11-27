############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder

RUN apk add --no-cache git

ENV GO111MODULE=on
ENV CGO_ENABLED=0
WORKDIR /app
COPY . .

RUN go mod download

RUN go build -trimpath -ldflags="-s -w" -o ./main ./cmd/api-server-gin/

############################
# STEP 2 build a small image
############################
FROM alpine:3

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/web/ ./web/

ENV PORT=8080
ENV GIN_MODE=release

EXPOSE 8080
ENTRYPOINT ["./main"]
