FROM golang:1.17

RUN go get github.com/go-delve/delve/cmd/dlv

ADD . /go/src/github.com/liuliqiang/go-demos
WORKDIR /go/src/github.com/liuliqiang/go-demos
RUN make server

FROM busybox:glibc

COPY --from=0 /lib/x86_64-linux-gnu/libdl.so.2 /lib/libdl.so.2
COPY --from=0 /go/bin/dlv /home/liqiang/
COPY --from=0 /go/src/github.com/liuliqiang/go-demos/.build/server /home/liqiang/

EXPOSE 2345 9000
WORKDIR /home/liqiang

CMD ["/home/liqiang/dlv", "--listen=:2345", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", \
     "/home/liqiang/server", "--", "-iden", "grpc-server"]