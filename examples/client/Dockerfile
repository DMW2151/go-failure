FROM golang:1.19-alpine as builder

COPY ./ /build/
RUN cd /build/examples/client && go build -o client

FROM alpine:latest
COPY --from=builder /build/examples/client/client ./client
