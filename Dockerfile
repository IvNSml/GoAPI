FROM golang:alpine

WORKDIR /app
COPY . /app

RUN apk add git
RUN go get -d ./...
RUN  GO111MODULE=on CGO_ENABLED=0 go build -o api

CMD ["./api"]

EXPOSE 8080