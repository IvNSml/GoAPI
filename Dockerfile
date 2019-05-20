FROM golang:alpine

WORKDIR $GOPATH/src/final

COPY . .

RUN go get -d -v ./...

RUN go-wrapper install

CMD ["go","run","my_final.go"]

EXPOSE 8080
