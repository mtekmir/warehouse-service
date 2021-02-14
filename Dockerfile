FROM golang:latest

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o main ./cmd/server/main.go

WORKDIR /dist

COPY internal/postgres/migrations ./internal/postgres/migrations

RUN cp /build/main .

EXPOSE 8080

CMD ["/dist/main"]