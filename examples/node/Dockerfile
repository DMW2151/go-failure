FROM golang:1.19-alpine as builder

COPY ./ /build/
RUN cd /build/examples/node && go build -o node

FROM alpine:latest
COPY --from=builder /build/examples/node/node ./node
