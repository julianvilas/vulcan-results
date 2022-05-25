# Copyright 2019 Adevinta

FROM golang:1.18.2-alpine3.15 as builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
RUN go install github.com/goadesign/goa/goagen@v1.4.3
COPY . .
RUN goagen app -d github.com/adevinta/vulcan-results/design
RUN goagen client -d github.com/adevinta/vulcan-results/design

RUN CGO_ENABLED=1 go install -a -tags netgo -ldflags '-w' ./...

FROM alpine:3.15

RUN apk add --no-cache --update gettext ca-certificates

ARG BUILD_RFC3339="1970-01-01T00:00:00Z"
ARG COMMIT="local"

ENV BUILD_RFC3339 "$BUILD_RFC3339"
ENV COMMIT "$COMMIT"

WORKDIR /app
COPY --from=builder /go/bin/vulcan-results .
COPY config.toml .
COPY run.sh .
CMD ["./run.sh"]
