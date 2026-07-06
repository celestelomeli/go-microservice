FROM golang:1.24.2

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o gateway .

EXPOSE 8080

CMD ["./gateway"]
