FROM golang:1.16

WORKDIR /go/src/app
COPY . .

RUN go version

ENV GOPATH=/

CMD go run ./cmd/main/main.go
