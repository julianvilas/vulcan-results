# Copyright 2019 Adevinta

FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

ARG TARGETOS TARGETARCH

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -tags netgo -ldflags '-w' ./cmd/vulcan-results

FROM alpine:3.21

RUN apk add --no-cache gettext ca-certificates

WORKDIR /app
COPY --from=builder /app/vulcan-results .
COPY config.toml .
COPY run.sh .
CMD ["./run.sh"]

