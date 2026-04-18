.PHONY: deps build clean

# Download dependencies
deps:
	go mod tidy

# Build the Lambda binary for AWS (Linux amd64, static, named 'bootstrap' for the provided.al2 runtime)
build:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bootstrap .
	zip lambda.zip bootstrap

clean:
	rm -f bootstrap lambda.zip
