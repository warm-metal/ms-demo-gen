FROM golang:1.21.3-alpine3.18 as builder
RUN apk add --no-cache make build-base
WORKDIR /go/src/github.com/warm-metal/ms-demo-gen
COPY pkg ./pkg
COPY cmd ./cmd
COPY go.mod go.sum Makefile ./
RUN go mod tidy && make

FROM alpine:3.18 as service
COPY --from=builder /go/src/github.com/warm-metal/ms-demo-gen/_output/service /usr/local/bin/
ENTRYPOINT [ "/usr/local/bin/service" ]

FROM alpine:3.18 as traffic-gen
COPY --from=builder /go/src/github.com/warm-metal/ms-demo-gen/_output/traffic_gen /usr/local/bin/
ENTRYPOINT [ "/usr/local/bin/traffic_gen" ]