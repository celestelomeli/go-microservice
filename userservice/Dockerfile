# Start from the official Golang image
FROM golang:1.24.2

# Set the working directory
WORKDIR /app

# Copy Go modules
COPY go.mod ./
# guarantees every time you build your Go app, using the exact same safe versions of every dependency
#COPY go.sum ./

# Download dependencies based on whats in go.mod & go.sum
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the Go app
RUN go build -o userservice .

# Expose port
EXPOSE 8083

# Command to run the executable
CMD ["./userservice"]
