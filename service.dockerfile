FROM golang:1.19.0-alpine3.16 as builder
WORKDIR /go/src/github.com/warm-metal/ms-demo-gen
COPY cmd/service ./cmd/service/
COPY pkg ./pkg
COPY go.mod go.sum Makefile ./
RUN go mod tidy && go build -o _output/service ./cmd/service

FROM alpine:3.16
COPY --from=builder /go/src/github.com/warm-metal/ms-demo-gen/_output/service /usr/local/bin/
ENTRYPOINT [ "/usr/local/bin/service" ]