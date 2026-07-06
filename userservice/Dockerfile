FROM golang:1.24.2

WORKDIR /app

# Copy module files first so dependency download is cached between builds
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o userservice .

EXPOSE 8083

CMD ["./userservice"]
