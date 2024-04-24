# building binary
FROM golang:1.22-alpine as builder

ENV GOPATH /go/
ENV GO_WORKDIR $GOPATH/src/github.com/watain666/ptt-alertor/
ENV GO111MODULE=on
ENV CGO_ENABLED=0

WORKDIR $GO_WORKDIR

ADD . $GO_WORKDIR

RUN go get
RUN go install

# building executable image
FROM alpine:latest as ptt-alertor

RUN set -eux; \
	apk add --no-cache --virtual ca-certificates

COPY public/ public/
COPY .env ./
COPY --from=builder /go/bin/ptt-alertor .

ENTRYPOINT /ptt-alertor

EXPOSE 9090 6060