# Use Go official image
FROM golang:1.24.2

# Set working directory inside container
WORKDIR /app

# copy your Go module files
COPY go.mod ./

# Download dependencies
RUN go mod tidy

# Copy the rest of the app's source code
COPY main.go ./

# Build the Go app into a binary executable
RUN go build -o frontendapp main.go

# Tell Docker how to run it: run the compiled binary
CMD ["./frontendapp"]
