FROM golang:1.23 AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go.mod download

COPY . .

FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .

CMD ["./main"]