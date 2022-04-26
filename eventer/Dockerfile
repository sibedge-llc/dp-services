# build image
FROM golang:1.18-alpine as builder
RUN apk add --update --no-cache git make gcc musl-dev

WORKDIR /app

COPY / ./

RUN go mod download

RUN GOOS=linux GOARCH=amd64 go build -a -tags musl -o /app/eventer ./cmd/eventer/


# executable image
FROM alpine:latest

RUN apk add --update --no-cache bash netcat-openbsd curl

WORKDIR /

COPY --from=builder /app/eventer ./

CMD ["./eventer", "start", "--config", "config.yaml"]

