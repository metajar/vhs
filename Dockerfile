# Start from the latest golang base image
FROM golang:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o vhs .

# This container exposes port 8080 to the outside world
EXPOSE 8080

# Install git
RUN apt-get update && apt-get install -y git
RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts

# Set Git global configuration
RUN git config --global user.email "vhs@henrynetworks.com"
RUN git config --global user.name "VHS"

# Command to run the executable
CMD ["./vhs"]
