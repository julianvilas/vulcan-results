# Copyright 2019 Adevinta

FROM --platform=$BUILDPLATFORM golang:1.21-alpine3.18 as builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

ARG TARGETOS TARGETARCH
WORKDIR /app/cmd/vulcan-results
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -tags netgo -ldflags '-w' .

FROM alpine:3.18

RUN apk add --no-cache --update gettext ca-certificates

ARG BUILD_RFC3339="1970-01-01T00:00:00Z"
ARG COMMIT="local"

ENV BUILD_RFC3339 "$BUILD_RFC3339"
ENV COMMIT "$COMMIT"

WORKDIR /app
COPY --from=builder /app/cmd/vulcan-results/vulcan-results .
COPY config.toml .
COPY run.sh .
CMD ["./run.sh"]
