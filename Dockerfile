FROM golang:1.23-alpine AS builder


COPY . .
RUN GO_ENABLED=0  go build \
	-o /go/bin/main ./cmd/server/main.go


FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /go/bin/main .
RUN chown root:root main

EXPOSE 8080
