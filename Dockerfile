FROM golang:1.27.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /ptt-alertor .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates
WORKDIR /

COPY public ./public
COPY --from=builder /ptt-alertor /ptt-alertor

EXPOSE 9090 6060

ENTRYPOINT ["/ptt-alertor"]
