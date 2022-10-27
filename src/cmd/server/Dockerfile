FROM golang:1.19-alpine as builder

COPY ./src /build/
RUN cd /build/cmd/server && go build -o server

FROM alpine:latest
COPY --from=builder /build/cmd/server/server ./server

ENTRYPOINT ./server