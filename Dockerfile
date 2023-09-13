FROM golang:1.20-alpine

WORKDIR /usr/local/src

# dependencies
COPY ["go.mod", "go.sum", "./"]
RUN go mod download

# build
COPY ./ ./
RUN go build -o avito-test ./cmd/app/main.go