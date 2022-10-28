FROM golang:1.19-alpine as builder

COPY ./ /build/
RUN cd /build/examples/server && go build -o server

FROM alpine:latest
COPY --from=builder /build/examples/server/server ./server
