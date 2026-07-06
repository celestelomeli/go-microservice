FROM golang:1.24.2

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY main.go ./
RUN go build -o frontendservice main.go

EXPOSE 3000

CMD ["./frontendservice"]
