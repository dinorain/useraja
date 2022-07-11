FROM golang:1.16

WORKDIR /app

ENV CONFIG=docker

COPY ./ /app

RUN go mod download
RUN go get github.com/githubnemo/CompileDaemon

EXPOSE 5000

ENTRYPOINT CompileDaemon --build="go build cmd/grpc/main.go" --command=./main