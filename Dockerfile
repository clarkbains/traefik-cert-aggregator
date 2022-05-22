FROM golang:1.18.2-buster AS builder
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY Makefile .
COPY aggregator aggregator
COPY clients clients
COPY cmd cmd
COPY config config
COPY util util
RUN make all

FROM alpine
LABEL org.opencontainers.image.source https://github.com/clarkbains/traefik-cert-aggregator
WORKDIR /app
COPY --from=builder /lib64/ld-linux-x86-64.so.2 /lib64/ld-linux-x86-64.so.2
COPY --from=builder /lib/x86_64-linux-gnu/libpthread.so.0 /lib/x86_64-linux-gnu/libpthread.so.0
COPY --from=builder /lib/x86_64-linux-gnu/libc.so.6 /lib/x86_64-linux-gnu/libc.so.6
COPY --from=builder /build/build/cert-agg main
CMD ["/app/main"]

