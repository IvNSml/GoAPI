FROM golang:alpine

WORKDIR /api_docker

COPY . api_docker

RUN apk add git
RUN go get -d -v ./...

CMD ["go","run","api.go"]

EXPOSE 8080
EXPOSE 5432
