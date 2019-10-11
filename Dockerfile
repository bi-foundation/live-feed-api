# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:latest

# Set the Current Working Directory inside the container, where the live-feed api sources live
WORKDIR $GOPATH/src/github.com/FactomProject/live-feed-api

# Copy go mod and sum files
COPY go.mod go.sum ./
COPY EventRouter/go.mod EventRouter/go.sum ./EventRouter/

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build ./factom-live-feed-api.go

COPY factom-live-feed.conf /etc/factom-live-feed/factom-live-feed.conf

# Expose port 8080 to the outside world
EXPOSE 8700 8040

# Command to run the executable
CMD ["./factom-live-feed-api"]