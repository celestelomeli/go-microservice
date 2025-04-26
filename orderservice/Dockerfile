# Start from the official Golang image
FROM golang:1.24.2

# Set the working directory
WORKDIR /app

# Copy Go modules
COPY go.mod ./
#COPY go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the Go app
RUN go build -o orderservice .

# Expose port
EXPOSE 8082

# Command to run the executable
CMD ["./orderservice"]
