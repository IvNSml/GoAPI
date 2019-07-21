FROM golang:alpine

COPY . .
WORKDIR $GOPATH/src/github.com/IvNSml/GoAPI

RUN apk add git
RUN go get -d ./...
CMD ["go","run","api.go"]

EXPOSE 8080
EXPOSE 5432
