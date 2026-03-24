FROM golang:alpine 

WORKDIR /app

COPY . .

RUN go build -o main ./cmd/server

EXPOSE 8080

CMD ["./main"]