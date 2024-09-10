# Start with a base Go image to build the application
FROM golang:1.22 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download the Go module dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

ENV CGO_ENABLED=1
# Expose the port the app will run on (optional, depends on your app)
EXPOSE 3000

RUN go install github.com/air-verse/air@latest

# Command to run the application
CMD ["air", "-c", ".air.toml"]
