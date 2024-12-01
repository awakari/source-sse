FROM golang:1.23.3-alpine3.20 AS builder
WORKDIR /go/src/source-sse
COPY . .
RUN \
    apk add protoc protobuf-dev make git && \
    make build

FROM alpine:3.20
RUN apk --no-cache add ca-certificates \
    && update-ca-certificates
COPY --from=builder /go/src/source-sse/source-sse /bin/source-sse
ENTRYPOINT ["/bin/source-sse"]
