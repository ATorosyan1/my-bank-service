FROM golang:1.16.3-alpine3.13 AS GO_BUILD
COPY . .
ENV GOPATH=/
RUN ls
CMD go run ./cmd/main/main.go
USER root
