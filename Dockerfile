FROM golang:1.10-alpine as builder

WORKDIR /go/src/github.com/dhiltgen/go-ioutil
COPY . /go/src/github.com/dhiltgen/go-ioutil
RUN go build -a -tags "netgo static_build" -installsuffix netgo -ldflags  "-w -extldflags '-static'" -o /go/bin/ioutil

FROM alpine:3.7
COPY --from=builder /go/bin/ioutil /bin/ioutil
ENTRYPOINT ["/bin/ioutil"]

