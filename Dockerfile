# Use the official Go image as the base image
FROM golang:1.24.3-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to the working directory
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
# CGO_ENABLED=0 is important for creating a statically linked executable, which is good for Docker
# -o /app/main specifies the output path and name of the executable
RUN CGO_ENABLED=0 go build -o /app/main .

# Expose the port your application will listen on
EXPOSE 8080

# Command to run the executable
CMD ["/app/main"]
